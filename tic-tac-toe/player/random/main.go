package main

import (
	"log"
	"math/rand"
	"time"

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
			print("Made an illegal move\n")
			break
		case ttt.Won:
			// game won
			print("Won the game!\n")
			ongoingGame = false
			break
		case ttt.Lost:
			print("Lost the game!\n")
			// game lost
			ongoingGame = false
			break
		default:
			// valid move
			turnCount++
			displayState(stateResult.State)
		}
		if !ongoingGame || turnCount > fieldCount {
			break
		}
		// state size? assuming 81 for now
		moveTarget := r.Int63n(int64(fieldCount))
		print("\nMoving to: ", moveTarget, "\n")
		stateResult, err := client.Move(ctx, &proto.Action{Id: id, Move: moveTarget})

		if stateResult != nil && err == nil {
			stateResult = nil
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func displayState(state []int64) {
	for index, element := range state {
		print(" ", element, " ")
		if (index+1)%3 == 0 {
			print("\n")
		}
	}
}
