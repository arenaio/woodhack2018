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

func (s *Server) NewGame(ctx context.Context, new *proto.New) (*proto.StateResult, error) {
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
	if playerId%2 == 0 {
		g = &Game{
			p1:    playerId,
			p2:    -1,
			state: make([]int64, 9),
			turn:  1,
		}
		s.games[playerId/2] = g
		print("New game created: ", playerId/2)
	} else {
		print("Trying to join game: ", (playerId-1)/2)
		g, ok := s.games[(playerId-1)/2]
		if !ok {
			return nil, errors.New("game not found")
		}
		g.p2 = playerId
	}

	s.nextPlayerId++

	return &proto.StateResult{
		Id:     playerId,
		State:  g.state,
		Result: 0,
	}, nil
}

func (s *Server) Move(ctx context.Context, a *proto.Action) (*proto.StateResult, error) {
	g, ok := s.games[a.Id-a.Id%2]
	if !ok {
		return nil, errors.New("game not found")
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
	g.turn = 2 - g.turn

	// TODO: wait until it's our turn again

	return &proto.StateResult{
		Id:     a.Id,
		State:  g.state,
		Result: ttt.ValidMove,
	}, nil
}

type Game struct {
	p1, p2 int64
	state  []int64
	turn   int64
}
