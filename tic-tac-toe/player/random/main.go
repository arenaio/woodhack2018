package main

import (
	"net"
	"log"
	"google.golang.org/grpc"
	"github.com/arenaio/woodhack2018/tic-tac-toe/proto"
	"golang.org/x/net/context"
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

type Server struct{}

func NewServer() *Server{
	return &Server{}
}

func (s *Server) GetMove(ctx context.Context, in *proto.GetMoveRequest) (out *proto.GetMoveResponse, err error) {
	out = &proto.GetMoveResponse{}
	out.Move = &proto.Move{}

	return out, err
}

type Move struct {
	x int64
	y int64
}

type Moves []*Move
