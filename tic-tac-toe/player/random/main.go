package main

import (
	"log"
	"google.golang.org/grpc"
	"golang.org/x/net/context"

	"github.com/arenaio/woodhack2018/tic-tac-toe/proto"
)

func main() {
	address := ":8000"

	conn, err := grpc.Dial(address)
	if err != nil {
		log.Fatalf("unable to connect on port %s: %v", address, err)
	}
	defer conn.Close()

	client := proto.NewTicTacToeClient(conn)

	ctx := context.Background()
	stateResult, err := client.NewGame(ctx, &proto.Empty{})
	print(stateResult)
}
