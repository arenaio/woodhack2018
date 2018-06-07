package main

import (
	"log"
	"math/rand"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	ttt "github.com/arenaio/woodhack2018/tic-tac-toe"
	"github.com/arenaio/woodhack2018/tic-tac-toe/proto"
)

func main() {
	address := ":8000"

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect on port %s: %v", address, err)
	}
	defer conn.Close()

	client := proto.NewTicTacToeClient(conn)

	ctx := context.Background()
	stateResult, err := client.NewGame(ctx, &proto.New{GameType: ttt.RegularTicTacToe})
	id := stateResult.Id
	ongoingGame := true
	r := rand.New(rand.NewSource(199))

	turnCount := 0
	fieldCount := 9

	for {
		switch stateResult.Result {
		case ttt.InvalidMove:
			// invalid move
			break
		case ttt.Won:
			// game won
			ongoingGame = false
			break
		case ttt.Lost:
			// game lost
			ongoingGame = false
			break
		default:
			// valid move
		}
		if !ongoingGame || turnCount > fieldCount {
			break
		}
		turnCount++
		// state size? assuming 81 for now
		stateResult, err := client.Move(ctx, &proto.Action{Id: id, Move: r.Int63n(int64(fieldCount))})

		if stateResult != nil && err == nil {
			stateResult = nil
		}
	}
}
