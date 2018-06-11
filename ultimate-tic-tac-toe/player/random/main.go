package main

import (
	"flag"
	"log"
	"math/rand"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	ttt "github.com/arenaio/woodhack2018/ultimate-tic-tac-toe"
	"github.com/arenaio/woodhack2018/ultimate-tic-tac-toe/proto"
)

var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(199))
}

func main() {
	address := flag.String("address", ":8000", "server address")
	name := flag.String("name", "Random", "bot name")
	flag.Parse()

	for {
		runGameOnServer(*address, *name)
	}
}

func runGameOnServer(address string, name string) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect on port %s: %v", address, err)
	}
	defer conn.Close()

	client := proto.NewTicTacToeClient(conn)
	ctx := context.Background()

	stateResult, err := client.NewGame(ctx, &proto.New{GameType: ttt.UltimateTicTacToe, Name: name})
	if err != nil {
		log.Fatal(err)
	}

	id := stateResult.Id
	ongoingGame := true

	for ongoingGame {
		stateResult, err = client.Move(ctx, &proto.Action{
			Id: id,
			Move: int64(r.Intn(len(stateResult.State))),
		})
		if err != nil {
			log.Fatal(err)
		}

		switch stateResult.Result {
		case ttt.Won:
			ongoingGame = false
		case ttt.Lost:
			ongoingGame = false
		case ttt.Draw:
			ongoingGame = false
		case ttt.InvalidMove:
		default:
			// valid move
		}
	}
}
