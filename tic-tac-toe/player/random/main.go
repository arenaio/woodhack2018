package main

import (
	"flag"
	"log"
	"math/rand"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	ttt "github.com/arenaio/woodhack2018/tic-tac-toe"
	"github.com/arenaio/woodhack2018/tic-tac-toe/proto"
)

var r *rand.Rand
var clientType int
var clientName string

func init() {
	r = rand.New(rand.NewSource(199))
}

func main() {
	address := ":8000"

	clientTypePtr := flag.Int("type", 1, "an int")

	flag.Parse()

	clientType = *clientTypePtr

	switch clientType {
	case 1:
		clientName = "Random"
	case 2:
		clientName = "Random without invalid"
	case 3:
		clientName = "Random with clever moves"
	}

	for {
		runGameOnServer(address)
	}
}

func runGameOnServer(address string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect on port %s: %v", address, err)
	}
	defer conn.Close()
	client := proto.NewTicTacToeClient(conn)
	ctx := context.Background()
	log.Print("Starting new game")
	stateResult, err := client.NewGame(ctx, &proto.New{GameType: ttt.RegularTicTacToe, Name: clientName})
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
		case ttt.Draw:
			print("Draw game!\n")
			ongoingGame = false
		default:
			// valid move
			turnCount++
			displayState(stateResult.State)
		}

		if !ongoingGame || turnCount > fieldCount {
			displayState(stateResult.State)
			break
		}

		moveTarget := makeMove(stateResult.State)
		print("\nMoving to: ", moveTarget, "\n")
		stateResult, err = client.Move(ctx, &proto.Action{Id: id, Move: moveTarget})

		if err != nil {
			log.Fatal(err)
		}
		//time.Sleep(100 * time.Millisecond)
	}
}

func makeMove(state []int64) int64 {
	var moveTarget int
	switch clientType {
	case 1:
		// Fully random
		moveTarget = r.Intn(len(state))
	case 2:
		// Randomly select valid targets
		validTargets := getValidTargets(state)
		moveTarget = validTargets[r.Intn(len(validTargets))]
	case 3:
		// select winning targets
		cleverTargets := getCleverTargets(state)
		if len(cleverTargets) > 0 {
			moveTarget = cleverTargets[r.Intn(len(cleverTargets))]
		} else {
			validTargets := getValidTargets(state)
			moveTarget = validTargets[r.Intn(len(validTargets))]
		}
	}

	return int64(moveTarget)
}

func getValidTargets(state []int64) []int {
	tmp := make([]int, 0)
	for i, v := range state {
		if v == 0 {
			tmp = append(tmp, i)
		}
	}
	return tmp
}

func getCleverTargets(state []int64) []int {
	tmp := make([]int, 0)
	for i, v := range state {
		if v == 0 && wouldMoveEndGame(state, i) {
			tmp = append(tmp, i)
		}
	}
	return tmp
}

func wouldMoveEndGame(state []int64, position int) bool {
	switch true {
	case position == 0 && state[1] == state[2],
		position == 1 && state[0] == state[2],
		position == 2 && state[0] == state[1],
		position == 3 && state[4] == state[5],
		position == 4 && state[3] == state[5],
		position == 5 && state[3] == state[4],
		position == 6 && state[7] == state[8],
		position == 7 && state[6] == state[8],
		position == 8 && state[6] == state[7],
		position == 0 && state[4] == state[8],
		position == 2 && state[4] == state[6],
		position == 4 && state[0] == state[8],
		position == 4 && state[2] == state[6],
		position == 6 && state[4] == state[2],
		position == 8 && state[0] == state[4],
		position < 3 && state[position+3] == state[position+6],
		position >= 3 && position < 6 && state[position+3] == state[position-3],
		position >= 6 && state[position-3] == state[position-6]:
		return true
	}
	return false
}

func displayState(state []int64) {
	for index, element := range state {
		print(" ", element, " ")
		if (index+1)%3 == 0 {
			print("\n")
		}
	}
}
