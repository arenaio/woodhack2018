package main

import (
	"flag"
	"log"
	"math/rand"
        "os"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	ttt "github.com/arenaio/woodhack2018/tic-tac-toe"
	"github.com/arenaio/woodhack2018/tic-tac-toe/proto"
	"strings"
	"strconv"
	"math"
	"encoding/json"
	"io/ioutil"
	"fmt"
)

var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(199))
}

func main() {
	address := flag.String("address", ":8000", "server address")
	name := flag.String("name", "Q-Table", "bot name")
        file := flag.String("file", "", "File with Q-Table dataset.")
	flag.Parse()

	q := &Qlearning{
		Table:           make(map[string]map[int64]float64),
		ExplorationRate: 1,
		LearningRate:    0.001,
		DiscountFactor:  1,
                Epoches:         0,
	}

        if len(*file) > 0 {
                log.Printf("Fetching from state file: %s", *file)
                q.fetchFromFile(*file)
                q.ExplorationRate = 0
        }

	for gameCount := 0; ; gameCount++ {
                log.Printf("Current Epoche: %d", q.Epoches)
		q.runGameOnServer(*address, *name)

                // read the epoches if set
                if q.Epoches > 0 && gameCount == 0 {
                        gameCount = q.Epoches
                }

		if gameCount%1000 == 0 && q.ExplorationRate > 0 {

                        q.Epoches = gameCount/1000
			q.storeTable(gameCount)
		}
	}
}

type Qlearning struct {
	Table           map[string]map[int64]float64 `json:"Table"`
	ExplorationRate float64                      `json:"ExplorationRate"`
	LearningRate    float64                      `json:"LearningRate"`
	DiscountFactor  float64                      `json:"DiscountFactor"`
        Epoches         int                          `json:"Epoches"`
}

func (q *Qlearning) storeTable(gameCount int) {
	bytes, _ := json.Marshal(q)

	err := ioutil.WriteFile(fmt.Sprintf("./%d.json", gameCount), bytes, 0644)
	if err != nil {
		log.Printf("Error storing table %d: %s", gameCount, err)
	}
}

func (q *Qlearning) fetchFromFile(filePath string) {
        raw, err := ioutil.ReadFile(filePath)
        if err != nil {
                log.Printf("Error loading state file: %s, %s", filePath, err)
                os.Exit(1)
        }

        json.Unmarshal(raw, &q)
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

	estimatedOptimalFuture := float64(ttt.InvalidMove)
	for _, qvalue := range futureActionTable {
		if qvalue > estimatedOptimalFuture {
			estimatedOptimalFuture = qvalue
		}
	}

	learnedValue := float64(reward) + q.DiscountFactor*estimatedOptimalFuture
	actionTable[action] = (1-q.LearningRate)*actionTable[action] + q.LearningRate*learnedValue

	q.ExplorationRate *= 0.999999
}

func (q *Qlearning) runGameOnServer(address, name string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect on port %s: %v", address, err)
	}
	defer conn.Close()
	client := proto.NewTicTacToeClient(conn)
	ctx := context.Background()

	log.Print("Starting new game")
	stateResult, err := client.NewGame(ctx, &proto.New{GameType: ttt.RegularTicTacToe, Name: name})

	displayState(stateResult.State)

	id := stateResult.Id
	ongoingGame := true

	for ongoingGame {
		action := q.makeMove(stateResult.State)
		print("\nMoving to: ", action, "\n")

		lastState := stateResult.State

		stateResult, err = client.Move(ctx, &proto.Action{Id: id, Move: action})

		if err != nil {
			log.Fatal(err)
		}

                // don't train when the exploration rate is set to zero
                if q.ExplorationRate > 0 {
                        q.train(lastState, action, stateResult.State, stateResult.Result)
                }

		switch stateResult.Result {
		case ttt.InvalidMove:
			print("Made an illegal move\n")
		case ttt.Won:
			print("Won the game!\n")
			ongoingGame = false
		case ttt.Lost:
			print("Lost the game!\n")
			ongoingGame = false
		case ttt.Draw:
			print("Draw game!\n")
			ongoingGame = false
		default:
			// valid move
			//displayState(stateResult.State)
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

	displayState(state)
	// log.Printf(
		// "%.4f %.4f %.4f\n%.4f %.4f %.4f\n%.4f %.4f %.4f",
		// actionTable[0],
		// actionTable[1],
		// actionTable[2],
		// actionTable[3],
		// actionTable[4],
		// actionTable[5],
		// actionTable[6],
		// actionTable[7],
		// actionTable[8],
	// )

	if r.Float64() < q.ExplorationRate {
		log.Printf("Explore (%.2f)", q.ExplorationRate)
		return int64(r.Intn(len(state)))
	}

	log.Printf("I know what's best... (%.2f)", q.ExplorationRate)
	bestMove := int64(-1)
	bestValue := -math.MaxFloat64
	for action, value := range actionTable {
		// log.Printf("current: %f, best: %f, action: %d, best: %d", value, bestValue, action, bestMove)
		if value > bestValue {
			bestValue = value
			bestMove = action
		}
	}

	return bestMove
}

func displayState(state []int64) {
	for index, element := range state {
		print(" ", element, " ")
		if (index+1)%3 == 0 {
			print("\n")
		}
	}
}
