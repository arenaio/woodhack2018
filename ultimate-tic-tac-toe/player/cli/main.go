package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/nsf/termbox-go"
	"google.golang.org/grpc"

	"github.com/arenaio/woodhack2018/proto"
)

func main() {
	name := flag.String("name", "CLI", "your name")
	address := flag.String("address", ":8000", "server address")
	player1Char := flag.String("player1Char", "X", "Character to use for Player 1")
	player2Char := flag.String("player2Char", "O", "Character to use for Player 2")
	flag.Parse()

	err := termbox.Init()
	if err != nil {
		log.Fatalf("unable to initialize terminal interface: %s", err)
	}

	conn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect on port %s: %v", *address, err)
	}
	defer conn.Close()

	g := NewGame(*name, *player1Char, *player2Char)
	err = g.run(proto.NewTicTacToeClient(conn), context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

type Game struct {
	name         string
	player1Char  string
	player2Char  string
	positionXOld int
	positionYOld int
	positionX    int
	positionY    int
}

func NewGame(name, player1Char, player2Char string) *Game {
	return &Game{
		name:         name,
		player1Char:  player1Char,
		player2Char:  player2Char,
		positionX:    1,
		positionXOld: 1,
		positionY:    1,
		positionYOld: 1,
	}
}

func (g *Game) parseField(field int64) string {
	switch field {
	case 0:
		return " "
	case 1:
		return fmt.Sprintf("\033[0;34m%s\033[0m", g.player1Char)
	case -1:
		return fmt.Sprintf("\033[0;32m%s\033[0m", g.player2Char)
	default:
		panic(fmt.Sprintf("invalid field received: %d", field))
	}
}

func (g *Game) displayPos(x, y int) (pX, pY int) {
	pX = x * 2
	pY = y * 4
	if x > 3 {
		pX += 2
	}
	if x > 6 {
		pX += 2
	}
	if y > 3 {
		pY += 4
	}
	if y > 6 {
		pY += 4
	}
	return pX, pY
}

func (g *Game) goTo(x int, y int) {
	pX, pY := g.displayPos(g.positionXOld, g.positionYOld)
	fmt.Printf("\033[%v;%vH \033[%v;%vH ", pX, pY-2, pX, pY)
	g.positionXOld = g.positionX
	g.positionYOld = g.positionY
	pX, pY = g.displayPos(x, y)
	fmt.Printf("\033[0;31m\033[%v;%vH[\033[%v;%vH]\033[0m", pX, pY-2, pX, pY)
}

func (g *Game) drawState(state []int64) {
	i := 0
	fmt.Printf("┌───┬───┬───┐   ┌───┬───┬───┐   ┌───┬───┬───┐ \n")
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", g.parseField(state[i]), g.parseField(state[i+1]), g.parseField(state[i+2]), g.parseField(state[i+9]), g.parseField(state[i+10]), g.parseField(state[i+11]), g.parseField(state[i+18]), g.parseField(state[i+19]), g.parseField(state[i+20]))
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", g.parseField(state[i]), g.parseField(state[i+1]), g.parseField(state[i+2]), g.parseField(state[i+9]), g.parseField(state[i+10]), g.parseField(state[i+11]), g.parseField(state[i+18]), g.parseField(state[i+19]), g.parseField(state[i+20]))
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", g.parseField(state[i]), g.parseField(state[i+1]), g.parseField(state[i+2]), g.parseField(state[i+9]), g.parseField(state[i+10]), g.parseField(state[i+11]), g.parseField(state[i+18]), g.parseField(state[i+19]), g.parseField(state[i+20]))
	fmt.Printf("└───┴───┴───┘   └───┴───┴───┘   └───┴───┴───┘ \n")
	print("\n")
	fmt.Printf("┌───┬───┬───┐   ┌───┬───┬───┐   ┌───┬───┬───┐ \n")
	i += 21
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", g.parseField(state[i]), g.parseField(state[i+1]), g.parseField(state[i+2]), g.parseField(state[i+9]), g.parseField(state[i+10]), g.parseField(state[i+11]), g.parseField(state[i+18]), g.parseField(state[i+19]), g.parseField(state[i+20]))
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", g.parseField(state[i]), g.parseField(state[i+1]), g.parseField(state[i+2]), g.parseField(state[i+9]), g.parseField(state[i+10]), g.parseField(state[i+11]), g.parseField(state[i+18]), g.parseField(state[i+19]), g.parseField(state[i+20]))
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", g.parseField(state[i]), g.parseField(state[i+1]), g.parseField(state[i+2]), g.parseField(state[i+9]), g.parseField(state[i+10]), g.parseField(state[i+11]), g.parseField(state[i+18]), g.parseField(state[i+19]), g.parseField(state[i+20]))
	fmt.Printf("└───┴───┴───┘   └───┴───┴───┘   └───┴───┴───┘ \n")
	print("\n")
	fmt.Printf("┌───┬───┬───┐   ┌───┬───┬───┐   ┌───┬───┬───┐ \n")
	i += 21
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", g.parseField(state[i]), g.parseField(state[i+1]), g.parseField(state[i+2]), g.parseField(state[i+9]), g.parseField(state[i+10]), g.parseField(state[i+11]), g.parseField(state[i+18]), g.parseField(state[i+19]), g.parseField(state[i+20]))
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", g.parseField(state[i]), g.parseField(state[i+1]), g.parseField(state[i+2]), g.parseField(state[i+9]), g.parseField(state[i+10]), g.parseField(state[i+11]), g.parseField(state[i+18]), g.parseField(state[i+19]), g.parseField(state[i+20]))
	fmt.Printf("├───┼───┼───┤   ├───┼───┼───┤   ├───┼───┼───┤ \n")
	i += 3
	fmt.Printf("│ %v │ %v │ %v │   │ %v │ %v │ %v │   │ %v │ %v │ %v │ \n", g.parseField(state[i]), g.parseField(state[i+1]), g.parseField(state[i+2]), g.parseField(state[i+9]), g.parseField(state[i+10]), g.parseField(state[i+11]), g.parseField(state[i+18]), g.parseField(state[i+19]), g.parseField(state[i+20]))
	fmt.Printf("└───┴───┴───┘   └───┴───┴───┘   └───┴───┴───┘ \n")
}

func (g *Game) getMove(x, y int) int {
	fX := []int{0, 3, 6, 27, 30, 33, 54, 57, 60}
	fY := []int{0, 1, 2, 9, 10, 11, 18, 19, 20}
	return fX[x-1] + fY[y-1]
}

func (g *Game) drawInput(state []int64) {
	fmt.Printf("\033[0;0H") // go to pos 0/0
	g.drawState(state)
	g.goTo(g.positionX, g.positionY) // draw input brackets
	fmt.Printf("\033[25;0H => Your turn           ")
}

func (g *Game) drawFinal(state []int64, message string) {
	g.drawState(state)
	fmt.Println("  ", message)
}

func (g *Game) run(client proto.TicTacToeClient, ctx context.Context) error {
	stateResult, err := client.NewGame(ctx, &proto.New{GameType: proto.UltimateTicTacToe, Name: g.name})
	if err != nil {
		return errors.New(fmt.Sprintf("unable join game: %s", err))
	}

	id := stateResult.Id
	g.drawInput(stateResult.State)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				termbox.Close()
				return nil
			case termbox.KeyArrowUp:
				if g.positionX > 1 {
					g.positionX--
					g.goTo(g.positionX, g.positionY)
				}
			case termbox.KeyArrowDown:
				if g.positionX < 9 {
					g.positionX++
					g.goTo(g.positionX, g.positionY)
				}
			case termbox.KeyArrowLeft:
				if g.positionY > 1 {
					g.positionY--
					g.goTo(g.positionX, g.positionY)
				}
			case termbox.KeyArrowRight:
				if g.positionY < 9 {
					g.positionY++
					g.goTo(g.positionX, g.positionY)
				}
			case termbox.KeySpace:
				// undraw input brackets and place X in orange
				pX, pY := g.displayPos(g.positionX, g.positionY)
				fmt.Printf("\033[0;94m\033[%v;%vH %s \033[0m", pX, pY-2, g.player1Char)
				fmt.Printf("\033[25;0H => Enemies turn           ")
				moveTarget := g.getMove(g.positionX, g.positionY)
				stateResult, err = client.Move(ctx, &proto.Action{Id: id, Move: int64(moveTarget)})
				if err != nil {
					log.Fatalf("an error trying to make a move: %s", err)
				}
				switch stateResult.Result {
				case proto.InvalidMove:
					g.drawInput(stateResult.State)
					fmt.Printf("\033[25;0H => Invalid Move, Your turn           ")
				case proto.Won:
					termbox.Close()
					g.drawFinal(stateResult.State, "You Won")
					return nil
				case proto.Lost:
					termbox.Close()
					g.drawFinal(stateResult.State, "You Lost")
					return nil
				case proto.Draw:
					termbox.Close()
					g.drawFinal(stateResult.State, "Game Draw")
					return nil
				default: // valid move
					g.drawInput(stateResult.State)
				}
			}
		case termbox.EventError:
			termbox.Close()
			return ev.Err
		}
	}
}
