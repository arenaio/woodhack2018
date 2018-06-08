package main
import (
	"fmt"
	"log"

	term "github.com/nsf/termbox-go"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	ttt "github.com/arenaio/woodhack2018/tic-tac-toe"
	"github.com/arenaio/woodhack2018/tic-tac-toe/proto"
)

func reset() {
    term.Sync() // cosmestic purpose
}

var positionXOld int = 1
var positionYOld int = 1
var positionX int = 1
var positionY int = 1

func main() {
	parseField := func (field int64) string {
		if field == 0 {return " "}
		if field == 1 {return "\033[0;34mX\033[0m"}
		if field == 2 {return "\033[0;32mO\033[0m"}
		panic("EINVALID VALUE RECEIVED")
	}
	goTo := func (x int, y int) {
		fmt.Printf("\033[%v;%vH \033[%v;%vH ", positionXOld*2, positionYOld*4-2, positionXOld*2, positionYOld*4)
		positionXOld = positionX
		positionYOld = positionY
		fmt.Printf("\033[0;31m\033[%v;%vH[\033[%v;%vH]\033[0m", x*2, y*4-2, x*2, y*4)
	}

	drawState := func (state []int64) {
		// go to pos 0/0
		fmt.Printf("\033[0;0H")
		fmt.Printf("┌───┬───┬───┐\n")
		fmt.Printf("│ %s │ %s │ %s │\n", parseField(state[0]), parseField(state[1]), parseField(state[2]))
		fmt.Printf("├───┼───┼───┤\n")
		fmt.Printf("│ %s │ %s │ %s │\n", parseField(state[3]), parseField(state[4]), parseField(state[5]))
		fmt.Printf("├───┼───┼───┤\n")
		fmt.Printf("│ %s │ %s │ %s │\n", parseField(state[6]), parseField(state[7]), parseField(state[8]))
		fmt.Printf("└───┴───┴───┘")
		goTo(positionX, positionY)
	}


	goTo(positionX, positionY)
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

	termErr := term.Init()
	if termErr != nil {
			panic(termErr)
	}
	drawState(stateResult.State)


	defer term.Close()
	keyPressListenerLoop:
	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyEsc:
				break keyPressListenerLoop
			case term.KeyArrowUp:
				fmt.Printf("\033[10;0H => Arrow Up pressed       ")
				if positionX > 1 {
					positionX--
					goTo (positionX, positionY)
				}
			case term.KeyArrowDown:
				fmt.Printf("\033[10;0H => Arrow Down pressed       ")
				if positionX < 3 {
					positionX++
					goTo (positionX, positionY)
				}
			case term.KeyArrowLeft:
				fmt.Printf("\033[10;0H => Arrow Left pressed       ")
				if positionY > 1 {
					positionY--
					goTo (positionX, positionY)
				}
			case term.KeyArrowRight:
				fmt.Printf("\033[10;0H => Arrow Right pressed       ")
				if positionY < 3 {
					positionY++
					goTo (positionX, positionY)
				}
			case term.KeySpace:
				fmt.Printf("\033[10;0H => Space pressed       ")
				moveTarget := (positionX-1)*3 + positionY -1
				stateResult, err = client.Move(ctx, &proto.Action{Id: id, Move: int64(moveTarget)})
				drawState(stateResult.State)
			}
		case term.EventError:
			panic(ev.Err)
		}
	}
}
