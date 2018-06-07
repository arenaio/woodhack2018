package main

import (
	"net"
	"log"

	"google.golang.org/grpc"
	"golang.org/x/net/context"

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

type Server struct{
	games map[int64]*Game
	nextPlayerId int64
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) NewGame(context.Context, *proto.Empty) (*proto.StateResult, error) {
	return nil, nil
}

func (s *Server) Move(context.Context, *proto.Action) (*proto.StateResult, error)   {
	return nil, nil
}

type Game struct {
	p1, p2 int64
	state []int64
}