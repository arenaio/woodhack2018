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

var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(199))
}

func main() {
	address := ":8000"

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect on port %s: %v", address, err)
	}
	defer conn.Close()

	client := proto.NewTicTacToeClient(conn)

	ctx := context.Background()
	stateResult, err := client.NewGame(ctx, &proto.New{GameType: ttt.RegularTicTacToe, Name: 'Random'})
	id := stateResult.Id
	ongoingGame := true

	turnCount := 0
	fieldCount := len(stateResult.State)

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

		moveTarget := makeMove(stateResult.State)
		print("\nMoving to: ", moveTarget, "\n")
		stateResult, err = client.Move(ctx, &proto.Action{Id: id, Move: moveTarget})

		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func makeMove(state []int64) int64 {
	// Fully random
	// moveTarget := r.Int63n(int64(len(state)))
	// Randomly select valid targets
	validTargets := getValidTargets(state)
	moveTarget := validTargets[r.Int63n(int64(len(validTargets)))]
	return moveTarget
}

func getValidTargets(state []int64) []int64 {
	tmp := make([]int64, 0)
	for i, v := range state {
		if v == 0 {
			tmp = append(tmp, int64(i))
		}
	}
	return tmp
}

func displayState(state []int64) {
	for index, element := range state {
		print(" ", element, " ")
		if (index+1)%3 == 0 {
			print("\n")
		}
	}
}
