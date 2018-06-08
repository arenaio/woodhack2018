package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	ttt "github.com/arenaio/woodhack2018/ultimate-tic-tac-toe"
	"github.com/arenaio/woodhack2018/ultimate-tic-tac-toe/proto"
	"sync"
	"flag"
)

func main() {
	address := flag.String("address", ":8000", "server address")
	flag.Parse()

	listener, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("unable to listen on port %s: %v", *address, err)
	}

	srv := grpc.NewServer()
	proto.RegisterTicTacToeServer(srv, NewServer())
	log.Printf("listening on %s", *address)
	log.Print(srv.Serve(listener))
}

type Server struct {
	games        map[int64]*Game
	m            sync.Mutex
	nextPlayerId int64
	stats        map[string]*Stats
}

func NewServer() *Server {
	return &Server{
		games: make(map[int64]*Game),
		stats: make(map[string]*Stats),
	}
}

func (s Server) String() string {
	games := make([]string, len(s.games))
	for i, g := range s.games {
		games[i] = g.String()
	}
	return strings.Join(games, "\n")
}

func (s Server) getStats() string {
	stats := []string{"Current Standing"}
	for name, stat := range s.stats {
		stats = append(stats, fmt.Sprintf(
			"%s\t%.0f%% (%d / %d / %d)",
			name,
			float64(stat.won)/float64(stat.won+stat.draw+stat.lost)*100,
			stat.won,
			stat.draw,
			stat.lost,
		))
	}

	return strings.Join(stats, "\n")
}

func (s *Server) NewGame(ctx context.Context, new *proto.New) (*proto.StateResult, error) {
	//log.Printf("Server.NewGame(%d)", new.GameType)

	s.m.Lock()
	_, found := s.stats[new.Name]
	if !found {
		s.stats[new.Name] = &Stats{}
	}

	playerId := s.nextPlayerId

	var g *Game
	var ok bool
	if playerId%2 == 0 {
		log.Print(s.getStats())

		g = &Game{
			p1:       playerId,
			p1Name:   new.Name,
			p2:       -1,
			state:    make([]int64, 81),
			turn:     1,
			d1:       make(chan struct{}),
			d2:       make(chan struct{}),
			lastMove: -1,
		}
		s.games[playerId/2] = g
		//log.Printf("new game created: %s", g)
	} else {
		//log.Printf("trying to join game: %d", (playerId-1)/2)
		g, ok = s.games[(playerId-1)/2]
		if !ok {
			return nil, errors.New("game not found")
		}
		g.p2 = playerId
		g.p2Name = new.Name
		//log.Printf("found game: %s", g)
	}

	s.nextPlayerId++
	s.m.Unlock()

	if playerId%2 != 0 {
		g.WaitForPlayer1() // since player 1 begins the game wait for first move
	}

	return &proto.StateResult{
		Id:     playerId,
		State:  mapOutput(g.state, playerId == g.p1),
		Result: ttt.ValidMove,
	}, nil
}

func (s *Server) Move(ctx context.Context, a *proto.Action) (*proto.StateResult, error) {
	log.Printf("Server.Move(Id: %d, Move: %d)", a.Id, a.Move)

	gameId := (a.Id - a.Id%2) / 2
	//log.Printf("Looking for game #%d", gameId)

	g, ok := s.games[gameId]
	if !ok {
		log.Printf("game #%d not found", gameId)
		return nil, errors.New("game not found")
	}
	//log.Printf("game found: %s", g)

	if g.isOver() {
		log.Printf("game #%d is already over", gameId)
		return nil, errors.New("game is already over")
	}

	isFirstPlayer := a.Id == g.p1
	if (g.turn == 1 && !isFirstPlayer) || (g.turn == 2 && isFirstPlayer) {
		log.Printf("player #%d tried to make a move, but it was not his turn", a.Id)
		return nil, errors.New("it's not your turn")
	}

	if g.state[a.Move] != 0 {
		//log.Printf("player #%d tried to make an invalid move (%d)", a.Id, a.Move)
		return &proto.StateResult{
			Id:       a.Id,
			State:    mapOutput(g.state, isFirstPlayer),
			Result:   ttt.InvalidMove,
			LastMove: g.lastMove,
		}, nil
	}

	if g.lastMove > -1 {
		subBoardStart := (g.lastMove % 9) * 9
		subBoardEnd := subBoardStart + 9

		subState := g.state[subBoardStart:subBoardEnd]
		if !isStateOver(subState) && (a.Move < subBoardStart || a.Move > subBoardEnd) {
			//log.Printf("player #%d tried to make an invalid move (%d)", a.Id, a.Move)
			return &proto.StateResult{
				Id:       a.Id,
				State:    mapOutput(g.state, isFirstPlayer),
				Result:   ttt.InvalidMove,
				LastMove: g.lastMove,
			}, nil
		}
	}

	g.state[a.Move] = g.turn
	g.turn = 3 - g.turn
	g.lastMove = a.Move

	//log.Printf("Game had received move: %s", g)

	if g.IsWon(a.Id) {
		if isFirstPlayer {
			s.stats[g.p1Name].won++
			s.stats[g.p2Name].lost++
			log.Printf("Game #%d is won by %s", gameId, g.p1Name)
			g.Player1Done()
		} else {
			s.stats[g.p1Name].lost++
			s.stats[g.p2Name].won++
			log.Printf("Game #%d is won by %s", gameId, g.p2Name)
			g.Player2Done()
		}

		return &proto.StateResult{
			Id:       a.Id,
			State:    mapOutput(g.state, a.Id == g.p1),
			Result:   ttt.Won,
			LastMove: g.lastMove,
		}, nil
	}

	if g.IsDraw() {
		if isFirstPlayer {
			g.Player1Done()
		} else {
			g.Player2Done()
		}
		s.stats[g.p1Name].draw++
		s.stats[g.p2Name].draw++
		return &proto.StateResult{
			Id:       a.Id,
			State:    mapOutput(g.state, a.Id == g.p1),
			Result:   ttt.Draw,
			LastMove: g.lastMove,
		}, nil
	}

	if isFirstPlayer {
		g.Player1Done()
		g.WaitForPlayer2()
	} else {
		g.Player2Done()
		g.WaitForPlayer1()
	}

	result := ttt.ValidMove
	if g.IsLost(a.Id) {
		if isFirstPlayer {
			log.Printf("Game #%d is lost for %s", gameId, g.p1Name)
		} else {
			log.Printf("Game #%d is lost for %s", gameId, g.p2Name)
		}
		result = ttt.Lost
	} else if g.IsDraw() {
		log.Printf("Game #%d is a draw", gameId)
		result = ttt.Draw
	}

	return &proto.StateResult{
		Id:       a.Id,
		State:    mapOutput(g.state, a.Id == g.p1),
		Result:   result,
		LastMove: g.lastMove,
	}, nil
}

// map user numbers to me: 1 and -1: opponent
func mapOutput(state []int64, p1Active bool) []int64 {
	var mapping map[int64]int64
	if p1Active {
		mapping = map[int64]int64{0: 0, 1: 1, 2: -1}
	} else {
		mapping = map[int64]int64{0: 0, 1: -1, 2: 1}
	}

	out := make([]int64, len(state))
	for i, v := range state {
		out[i] = mapping[v]
	}
	return out
}

type Game struct {
	p1, p2         int64
	p1Name, p2Name string
	d1, d2         chan struct{}
	state          []int64
	lastMove       int64
	turn           int64
}

func (g Game) String() string {
	//state := ""
	//for index, element := range g.state {
	//	state += fmt.Sprintf("%d ", element)
	//	if (index+1)%3 == 0 {
	//		state += "\t"
	//	}
	//}

	return fmt.Sprintf(
		"Game #%d: %d vs. %d turn: %d",
		g.p1/2, g.p1, g.p2, g.turn,
	)
}

func (g Game) IsWon(pId int64) bool {
	var p int64
	if g.p1 == pId {
		p = 1
	} else {
		p = 2
	}

	subBoards := make([]int64, 9)
	for i := 0; i < 9; i++ {
		subBoards[i] = getSubResult(g.state[i*9:i*9+9])
	}

	places := [][]int64{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 4, 8},
		{6, 4, 2},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
	}

	for _, ps := range places {
		if p == subBoards[ps[0]] && p == subBoards[ps[1]] && p == subBoards[ps[2]] {
			return true
		}
	}

	wonBoards := []int64{0, 0, 0}
	for _, result := range subBoards {
		if result == 0 {
			return false
		}

		if result > 0 {
			wonBoards[result]++
		}
	}

	return wonBoards[p] > wonBoards[3 - p]
}

func (g Game) IsLost(p int64) bool {
	if g.p1 == p {
		return g.IsWon(g.p2)
	} else {
		return g.IsWon(g.p1)
	}
}

func (g Game) IsDraw() bool {
	subBoards := make([]int64, 9)
	for i := 0; i < 9; i++ {
		subBoards[i] = getSubResult(g.state[i*9:i*9+9])
	}

	wonBoards := []int64{0, 0, 0}
	for _, result := range subBoards {
		if result == 0 {
			return false
		}

		if result > 0 {
			wonBoards[result]++
		}
	}

	return wonBoards[1] == wonBoards[2]
}

func isStateOver(state [] int64) bool {
	return getSubResult(state) != 0
}

func getSubResult(state[] int64) int64 {
	places := [][]int64{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 4, 8},
		{6, 4, 2},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
	}

	for _, ps := range places {
		if 0 != state[ps[0]] && state[ps[0]] == state[ps[1]] && state[ps[1]] == state[ps[2]] {
			return state[ps[0]]
		}
	}

	for _, v := range state {
		if v == 0 {
			return 0 // unfinished
		}
	}

	return -1 // draw
}

func (g Game) isOver() bool {
	return g.IsWon(g.p1) || g.IsWon(g.p2) || g.IsDraw()
}

func (g Game) Player1Done() {
	g.d1 <- struct{}{}
}

func (g Game) Player2Done() {
	g.d2 <- struct{}{}
}

func (g Game) WaitForPlayer1() {
	<-g.d1
}

func (g Game) WaitForPlayer2() {
	<-g.d2
}

type Stats struct {
	won  int64
	lost int64
	draw int64
}
