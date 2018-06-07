package main

import (
	"log"

	"math/rand"

	"github.com/arenaio/woodhack2018/tic-tac-toe/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
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
	stateResult, err := client.NewGame(ctx, &proto.New{GameType: 1})
	id := stateResult.Id
	ongoingGame := true
	r := rand.New(rand.NewSource(199))

	for {
		print(stateResult)
		switch stateResult.Result {
		case -2:
			// invalid move
			break
		case 1:
			// game won
			ongoingGame = false
			break
		case 2:
			// game lost
			ongoingGame = false
			break
		default:
			// valid move
		}
		if !ongoingGame {
			break
		}
		// state size? assuming 81 for now
		stateResult, err := client.Move(ctx, &proto.Action{Id: id, Move: r.Int63n(81)})

		print(stateResult)
		print(err)
	}
}
