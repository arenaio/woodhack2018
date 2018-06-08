package main
import (
	"fmt"
	"log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	term "github.com/nsf/termbox-go"
	ttt "github.com/arenaio/woodhack2018/tic-tac-toe"
	"github.com/arenaio/woodhack2018/tic-tac-toe/proto"
)

func main() {
	// player1Char := "X"
	// player2Char := "O"
	positionXOld := 1
	positionYOld := 1
	positionX := 1
	positionY := 1

	parseField := func (field int64) string {
		if field == 0 {return " "}
		if field == 1 {return "\033[0;34mX\033[0m"}
		if field == -1 {return "\033[0;32mO\033[0m"}
		panic("INVALID FIELD VALUE RECEIVED")
	}

	goTo := func (x int, y int) {
		fmt.Printf("\033[%v;%vH \033[%v;%vH ", positionXOld*2, positionYOld*4-2, positionXOld*2, positionYOld*4)
		positionXOld = positionX
		positionYOld = positionY
		fmt.Printf("\033[0;31m\033[%v;%vH[\033[%v;%vH]\033[0m", x*2, y*4-2, x*2, y*4)
	}

	drawState := func (state []int64) {
		fmt.Printf("\033[0;0H")	// go to pos 0/0
		fmt.Printf("┌───┬───┬───┐\n")
		fmt.Printf("│ %s │ %s │ %s │\n", parseField(state[0]), parseField(state[1]), parseField(state[2]))
		fmt.Printf("├───┼───┼───┤\n")
		fmt.Printf("│ %s │ %s │ %s │\n", parseField(state[3]), parseField(state[4]), parseField(state[5]))
		fmt.Printf("├───┼───┼───┤\n")
		fmt.Printf("│ %s │ %s │ %s │\n", parseField(state[6]), parseField(state[7]), parseField(state[8]))
		fmt.Printf("└───┴───┴───┘")
		goTo(positionX, positionY) // draw input brackets
		fmt.Printf("\033[8;0H => Your turn           ")
	}

	drawFinalState := func (state []int64, message string) {
		fmt.Printf("┌───┬───┬───┐\n")
		fmt.Printf("│ %s │ %s │ %s │\n", parseField(state[0]), parseField(state[1]), parseField(state[2]))
		fmt.Printf("├───┼───┼───┤\n")
		fmt.Printf("│ %s │ %s │ %s │\n", parseField(state[3]), parseField(state[4]), parseField(state[5]))
		fmt.Printf("├───┼───┼───┤\n")
		fmt.Printf("│ %s │ %s │ %s │\n", parseField(state[6]), parseField(state[7]), parseField(state[8]))
		fmt.Printf("└───┴───┴───┘\n")
		fmt.Println("  ", message)
	}

	address := ":8000"
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect on port %s: %v", address, err)
	}
	defer conn.Close()

	client := proto.NewTicTacToeClient(conn)

	ctx := context.Background()
	stateResult, err := client.NewGame(ctx, &proto.New{GameType: ttt.RegularTicTacToe, Name: "CLI"})
 	id := stateResult.Id

	termErr := term.Init()
	if termErr != nil {
			panic(termErr)
	}
	drawState(stateResult.State)

	keyPressListenerLoop:
	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyEsc:
				term.Close()
				break keyPressListenerLoop
			case term.KeyArrowUp:
				if positionX > 1 {
					positionX--
					goTo (positionX, positionY)
				}
			case term.KeyArrowDown:
				if positionX < 3 {
					positionX++
					goTo (positionX, positionY)
				}
			case term.KeyArrowLeft:
				if positionY > 1 {
					positionY--
					goTo (positionX, positionY)
				}
			case term.KeyArrowRight:
				if positionY < 3 {
					positionY++
					goTo (positionX, positionY)
				}
			case term.KeySpace:
				// undraw input brackets and place X in orange
				fmt.Printf("\033[0;94m\033[%v;%vH X \033[0m", positionX*2, positionY*4-2)
				fmt.Printf("\033[8;0H => Enemies turn           ")
				moveTarget := (positionX-1)*3 + positionY -1
				stateResult, err = client.Move(ctx, &proto.Action{Id: id, Move: int64(moveTarget)})
				switch stateResult.Result {
				case ttt.InvalidMove:
					drawState(stateResult.State)
					fmt.Printf("\033[8;0H => Invalid Move, Your turn           ")
				case ttt.Won:
					term.Close()
					drawFinalState(stateResult.State, "You Won")
					break keyPressListenerLoop
				case ttt.Lost:
					term.Close()
					drawFinalState(stateResult.State, "You Lost")
					break keyPressListenerLoop
				case ttt.Draw:
					term.Close()
					break keyPressListenerLoop
					drawFinalState(stateResult.State, "Game Draw")
				default:
					// valid move
					drawState(stateResult.State)
				}
			}
		case term.EventError:
			panic(ev.Err)
		}
	}
}
