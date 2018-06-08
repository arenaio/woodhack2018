package main

import (
	"errors"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	ttt "github.com/arenaio/woodhack2018/tic-tac-toe"
	"github.com/arenaio/woodhack2018/tic-tac-toe/proto"
	"strings"
	"sync"
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

	log.Printf("xXXXX: %s", g)

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
	//g.m.Lock()

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

	//g.m.Unlock()

	if a.Id == g.p1 {
		g.d1 <- struct{}{} // player 1 done
		<-g.d2             // wait for for player 2
	} else {
		g.d2 <- struct{}{}
		<-g.d1
	}

	return &proto.StateResult{
		Id:     a.Id,
		State:  mapOutput(g.state, a.Id == g.p1),
		Result: ttt.ValidMove,
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
	m      *sync.Mutex
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
