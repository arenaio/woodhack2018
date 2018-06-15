package main

import (
	"flag"
	"log"
	"math/rand"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/arenaio/woodhack2018/proto"
)

var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(199))
}

func main() {
	address := flag.String("address", ":8000", "server address")
	name := flag.String("name", "Random", "bot name")
	flag.Parse()

	conn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect on port %s: %s", *address, err)
	}
	defer conn.Close()

	client := proto.NewTicTacToeClient(conn)
	ctx := context.Background()

	for {
		runGameOnServer(client, ctx, *name)
	}
}

func runGameOnServer(client proto.TicTacToeClient, ctx context.Context, name string) {
	stateResult, err := client.NewGame(ctx, &proto.New{GameType: proto.UltimateTicTacToe, Name: name})
	if err != nil {
		log.Fatal(err)
	}

	id := stateResult.Id
	ongoingGame := true

	for ongoingGame {
		stateResult, err = client.Move(ctx, &proto.Action{
			Id:   id,
			Move: int64(r.Intn(len(stateResult.State))),
		})
		if err != nil {
			log.Fatal(err)
		}

		switch stateResult.Result {
		case proto.Won:
			ongoingGame = false
		case proto.Lost:
			ongoingGame = false
		case proto.Draw:
			ongoingGame = false
		case proto.InvalidMove:
		default:
			// valid move
		}
	}
}
