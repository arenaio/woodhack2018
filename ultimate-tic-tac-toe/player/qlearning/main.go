package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/arenaio/woodhack2018/proto"
)

var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func main() {
	address := flag.String("address", ":8000", "server address")
	name := flag.String("name", "Q-Table", "bot name")
	flag.Parse()

	conn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect on port %s: %s", *address, err)
	}
	defer conn.Close()

	client := proto.NewTicTacToeClient(conn)
	ctx := context.Background()

	q := &Qlearning{
		Table:           make(map[string]map[int64]float64),
		ExplorationRate: 1,
		LearningRate:    0.001,
		DiscountFactor:  1,
	}

	for gameCount := 1; ; gameCount++ {
		q.runGameOnServer(client, ctx, *name)

		if gameCount%1000 == 0 {
			log.Printf("%d Episodes - Exploration Rate: %.4f", gameCount, q.ExplorationRate)
		}
		if gameCount%10000 == 0 {
			q.storeTable(gameCount)
			log.Printf("%d.json saved", gameCount)
		}
	}
}

type Qlearning struct {
	Table           map[string]map[int64]float64 `json:"Table"`
	ExplorationRate float64                      `json:"ExplorationRate"`
	LearningRate    float64                      `json:"LearningRate"`
	DiscountFactor  float64                      `json:"DiscountFactor"`
}

func (q *Qlearning) storeTable(gameCount int) {
	bytes, _ := json.Marshal(q)

	err := ioutil.WriteFile(fmt.Sprintf("./%d.json", gameCount), bytes, 0644)
	if err != nil {
		log.Printf("Error storing table %d: %s", gameCount, err)
	}
}

func hashState(state []int64) string {
	stateStr := make([]string, len(state))
	for i, v := range state {
		stateStr[i] = strconv.Itoa(int(v))
	}
	return strings.Join(stateStr, "")
}

func (q *Qlearning) train(lastState []int64, action int64, futureState []int64, reward int64) {
	actionTable := q.getActionTable(lastState)
	futureActionTable := q.getActionTable(futureState)

	estimatedOptimalFuture := float64(proto.InvalidMove)
	for _, qvalue := range futureActionTable {
		if qvalue > estimatedOptimalFuture {
			estimatedOptimalFuture = qvalue
		}
	}

	learnedValue := float64(reward) + q.DiscountFactor*estimatedOptimalFuture
	actionTable[action] = (1-q.LearningRate)*actionTable[action] + q.LearningRate*learnedValue

	q.ExplorationRate *= 0.9999999
}

func (q *Qlearning) runGameOnServer(client proto.TicTacToeClient, ctx context.Context, name string) {
	//log.Print("Starting new game")
	stateResult, err := client.NewGame(ctx, &proto.New{GameType: proto.UltimateTicTacToe, Name: name})
	if err != nil {
		log.Fatal(err)
	}

	displayState(stateResult.State)

	id := stateResult.Id
	ongoingGame := true

	for ongoingGame {
		action := q.makeMove(stateResult.State)
		//print("\nMoving to: ", action, "\n")

		lastState := stateResult.State

		stateResult, err = client.Move(ctx, &proto.Action{Id: id, Move: action})

		if err != nil {
			displayState(lastState)
			log.Fatal(err)
		}

		q.train(lastState, action, stateResult.State, stateResult.Result)

		switch stateResult.Result {
		case proto.InvalidMove:
			//print("Made an illegal move\n")
		case proto.Won:
			//print("Won the game!\n")
			ongoingGame = false
		case proto.Lost:
			//print("Lost the game!\n")
			ongoingGame = false
		case proto.Draw:
			//print("Draw game!\n")
			ongoingGame = false
		default:
			// valid move
			displayState(stateResult.State)
		}
	}

	displayState(stateResult.State)
}

func (q *Qlearning) getActionTable(state []int64) map[int64]float64 {
	hash := hashState(state)

	actionTable, found := q.Table[hash]
	if !found {
		actionTable = make(map[int64]float64)
		for i := 0; i < len(state); i++ {
			actionTable[int64(i)] = 0
		}
		q.Table[hash] = actionTable
	}

	return actionTable
}

func (q *Qlearning) makeMove(state []int64) int64 {
	actionTable := q.getActionTable(state)

	if r.Float64() < q.ExplorationRate {
		//log.Printf("Exploring (%.2f)", q.ExplorationRate)
		return int64(r.Intn(len(state)))
	}

	//log.Printf("I know what's best... (%.2f)", q.ExplorationRate)
	bestMove := int64(-1)
	bestValue := -math.MaxFloat64
	for action, value := range actionTable {
		if value > bestValue {
			bestValue = value
			bestMove = action
		}
	}

	return bestMove
}

func displayState(state []int64) {
	return
	lookup := map[int64]string{-1: "X", 0: " ", 1: "O"}

	i := 0
	print("\n")
	fmt.Printf("┌───┬───┬───┐   ┌───┬───┬───┐   ┌───┬───┬───┐ \n")
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", lookup[state[i]], lookup[state[i+1]], lookup[state[i+2]], lookup[state[i+9]], lookup[state[i+10]], lookup[state[i+11]], lookup[state[i+18]], lookup[state[i+19]], lookup[state[i+20]])
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", lookup[state[i]], lookup[state[i+1]], lookup[state[i+2]], lookup[state[i+9]], lookup[state[i+10]], lookup[state[i+11]], lookup[state[i+18]], lookup[state[i+19]], lookup[state[i+20]])
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", lookup[state[i]], lookup[state[i+1]], lookup[state[i+2]], lookup[state[i+9]], lookup[state[i+10]], lookup[state[i+11]], lookup[state[i+18]], lookup[state[i+19]], lookup[state[i+20]])
	fmt.Printf("└───┴───┴───┘   └───┴───┴───┘   └───┴───┴───┘ \n")
	print("\n")
	fmt.Printf("┌───┬───┬───┐   ┌───┬───┬───┐   ┌───┬───┬───┐ \n")
	i += 21
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", lookup[state[i]], lookup[state[i+1]], lookup[state[i+2]], lookup[state[i+9]], lookup[state[i+10]], lookup[state[i+11]], lookup[state[i+18]], lookup[state[i+19]], lookup[state[i+20]])
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", lookup[state[i]], lookup[state[i+1]], lookup[state[i+2]], lookup[state[i+9]], lookup[state[i+10]], lookup[state[i+11]], lookup[state[i+18]], lookup[state[i+19]], lookup[state[i+20]])
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", lookup[state[i]], lookup[state[i+1]], lookup[state[i+2]], lookup[state[i+9]], lookup[state[i+10]], lookup[state[i+11]], lookup[state[i+18]], lookup[state[i+19]], lookup[state[i+20]])
	fmt.Printf("└───┴───┴───┘   └───┴───┴───┘   └───┴───┴───┘ \n")
	print("\n")
	fmt.Printf("┌───┬───┬───┐   ┌───┬───┬───┐   ┌───┬───┬───┐ \n")
	i += 21
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", lookup[state[i]], lookup[state[i+1]], lookup[state[i+2]], lookup[state[i+9]], lookup[state[i+10]], lookup[state[i+11]], lookup[state[i+18]], lookup[state[i+19]], lookup[state[i+20]])
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", lookup[state[i]], lookup[state[i+1]], lookup[state[i+2]], lookup[state[i+9]], lookup[state[i+10]], lookup[state[i+11]], lookup[state[i+18]], lookup[state[i+19]], lookup[state[i+20]])
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", lookup[state[i]], lookup[state[i+1]], lookup[state[i+2]], lookup[state[i+9]], lookup[state[i+10]], lookup[state[i+11]], lookup[state[i+18]], lookup[state[i+19]], lookup[state[i+20]])
	fmt.Printf("└───┴───┴───┘   └───┴───┴───┘   └───┴───┴───┘ \n")
}
