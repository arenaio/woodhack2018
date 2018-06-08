package main
import (
	"fmt"
	term "github.com/nsf/termbox-go"
)

func reset() {
    term.Sync() // cosmestic purpose
}

var positionXOld int = 1
var positionYOld int = 1
var positionX int = 1
var positionY int = 1
func main() {
	err := term.Init()
	if err != nil {
			panic(err)
	}
	defer term.Close()
	drawState()
	goTo(positionX, positionY)
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
				setX()
			}
		case term.EventError:
			panic(ev.Err)
		}
	}
}

func goTo(x int, y int) {
	fmt.Printf("\033[%v;%vH \033[%v;%vH ", positionXOld*2, positionYOld*4-2, positionXOld*2, positionYOld*4)
	positionXOld = positionX
	positionYOld = positionY
	fmt.Printf("\033[0;31m\033[%v;%vH[\033[%v;%vH]\033[0m", x*2, y*4-2, x*2, y*4)
}

func setX() {
	fmt.Printf("\033[0;34m\033[%v;%vHX\033[0m", positionX*2, positionY*4-1)
}

func drawState() {
	// go to pos 0/0
	fmt.Printf("\033[0;0H")
	fmt.Printf("┌───┬───┬───┐\n")
	fmt.Printf("│ %v │ %v │ %v │\n", " ", " ", " ")
	fmt.Printf("├───┼───┼───┤\n")
	fmt.Printf("│ %v │ %v │ %v │\n", " ", " ", " ")
	fmt.Printf("├───┼───┼───┤\n")
	fmt.Printf("│ %v │ %v │ %v │\n", " ", " ", " ")
	fmt.Printf("└───┴───┴───┘")
}
