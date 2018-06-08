package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	ttt "github.com/arenaio/woodhack2018/tic-tac-toe"
	"github.com/arenaio/woodhack2018/tic-tac-toe/proto"
)

func main() {
	address := ":8000"

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("unable to listen on port %s: %v", address, err)
	}

	srv := grpc.NewServer()
	proto.RegisterTicTacToeServer(srv, NewServer())
	srv.Serve(listener)
	log.Printf("listening on %s", address)
}

type Server struct {
	games        map[int64]*Game
	nextPlayerId int64
	//ultimateGames map[int64]*Game
}

func NewServer() *Server {
	return &Server{
		games: make(map[int64]*Game),
	}
}

func (s Server) String() string {
	games := make([]string, len(s.games))
	for i, g := range s.games {
		games[i] = g.String()
	}
	return strings.Join(games, "\n")
}

func (s *Server) NewGame(ctx context.Context, new *proto.New) (*proto.StateResult, error) {
	log.Printf("Server.NewGame(%d)", new.GameType)

	//ultimate := new.gameType == 1
	switch new.GameType {
	case ttt.RegularTicTacToe:
		break
	default:
		return nil, errors.New(fmt.Sprintf(
			"gametype %d not implemented yet",
			new.GameType,
		))
	}

	playerId := s.nextPlayerId

	var g *Game
	var ok bool
	if playerId%2 == 0 {
		g = &Game{
			p1:    playerId,
			p2:    -1,
			state: make([]int64, 9),
			turn:  1,
			d1:    make(chan struct{}),
			d2:    make(chan struct{}),
		}
		s.games[playerId/2] = g
		log.Printf("new game created: %s", g)
	} else {
		log.Printf("trying to join game: %d", (playerId-1)/2)
		g, ok = s.games[(playerId-1)/2]
		if !ok {
			return nil, errors.New("game not found")
		}
		g.p2 = playerId
		log.Printf("found game: %s", g)
		<-g.d1 // wait for player 1
	}

	s.nextPlayerId++

	return &proto.StateResult{
		Id:     playerId,
		State:  mapOutput(g.state, playerId == g.p1),
		Result: 0,
	}, nil
}

func (s *Server) Move(ctx context.Context, a *proto.Action) (*proto.StateResult, error) {
	log.Printf("Server.Move(Id: %d, Move: %d)", a.Id, a.Move)

	g, ok := s.games[a.Id-a.Id%2]
	if !ok {
		return nil, errors.New("game not found")
	}

	log.Printf("game found: %s", g)

	if g.IsDraw() || g.IsWon(g.p1) || g.IsWon(g.p2) {
		return nil, errors.New("game is already over")
	}

	if (g.turn == 1 && a.Id != g.p1) || (g.turn == 2 && a.Id != g.p2) {
		return nil, errors.New("it's not your turn")
	}

	if g.state[a.Move] != 0 {
		return &proto.StateResult{
			Id:     a.Id,
			State:  g.state,
			Result: ttt.InvalidMove,
		}, nil
	}

	g.state[a.Move] = g.turn
	g.turn = 3 - g.turn

	if g.IsWon(g.p1) {
		if a.Id == g.p1 {
			g.d2 <- struct{}{}
		} else {
			g.d1 <- struct{}{}
		}
		return &proto.StateResult{
			Id:     a.Id,
			State:  mapOutput(g.state, a.Id == g.p1),
			Result: ttt.Won,
		}, nil
	}

	if g.IsDraw() {
		if a.Id == g.p1 {
			g.d2 <- struct{}{}
		} else {
			g.d1 <- struct{}{}
		}
		return &proto.StateResult{
			Id:     a.Id,
			State:  mapOutput(g.state, a.Id == g.p1),
			Result: ttt.Draw,
		}, nil
	}

	if a.Id == g.p1 {
		g.d1 <- struct{}{} // player 1 done
		<-g.d2             // wait for for player 2
	} else {
		g.d2 <- struct{}{}
		<-g.d1
	}

	result := ttt.ValidMove
	if g.IsDraw() {
		result = ttt.Draw
	}

	if g.IsWon(a.Id) {
		result = ttt.Won
	}

	if g.IsLost(a.Id) {
		result = ttt.Lost
	}

	return &proto.StateResult{
		Id:     a.Id,
		State:  mapOutput(g.state, a.Id == g.p1),
		Result: result,
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
	p1, p2 int64
	d1, d2 chan struct{}
	state  []int64
	turn   int64
}

func (g Game) String() string {
	state := ""
	for index, element := range g.state {
		state += fmt.Sprintf(" %d ", element)
		if (index+1)%3 == 0 {
			state += "\n"
		}
	}

	return fmt.Sprintf(
		"Game #%d: %d vs. %d\n%sturn: %d",
		g.p1/2, g.p1, g.p2, state, g.turn,
	)
}

func (g Game) IsWon(p int64) bool {
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
		if p == ps[0] && ps[0] == ps[1] && ps[1] == ps[2] {
			return true
		}
	}

	return false
}

func (g Game) IsLost(p int64) bool {
	if g.p1 == p {
		return g.IsWon(g.p2)
	} else {
		return g.IsWon(g.p1)
	}
}

func (g Game) IsDraw() bool {
	for _, v := range g.state {
		if v == 0 {
			return false
		}
	}

	return true
}
