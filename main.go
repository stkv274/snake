package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"

	"golang.org/x/term"
)

type coord struct {
	x, y int
}

type Direction int

const (
	North Direction = 1
	East            = 2
	South           = 3
	West            = 4
)

type Snake struct {
	body      []coord
	direction Direction
	dead      bool
}

var (
	vertSlow = 140
	horzSlow = 100
)

func main() {
	// canonical mode -> non-canonical mode
	// TODO: re-write using syscalls
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic("can't get old term state...")
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// get terminal size
	// TODO: re-write using syscalls
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		panic("can't get term size")
	}

	// creating screen buffer
	scr := newScreen(w, h)
	defer scr.reset()

	snake := Snake{
		body:      []coord{{50, 20}, {51, 20}, {52, 20}, {53, 20}, {54, 20}},
		direction: West,
		dead:      false,
	}
	foodCoord := coord{40, 20}

	score := 0

	quit := false
	// handle kb input
	nitro := false
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				fmt.Println("can't read rune from reader...")
			}
			switch r {
			case 'A':
				if snake.direction%2 != 1 {
					snake.direction = North
				}
			case 'B':
				if snake.direction%2 != 1 {
					snake.direction = South
				}
			case 'D':
				if snake.direction%2 == 1 {
					snake.direction = West
				}
			case 'C':
				if snake.direction%2 == 1 {
					snake.direction = East
				}
			case 's':
				if !nitro {
					nitro = true
				} else {
					nitro = false
				}
			case 'q':
				quit = true
				return
			}
		}
	}()

	// game loop
	for !quit {
		scr.clear()

		if !snake.dead {
			switch snake.direction {
			case North:
				head := coord{snake.body[0].x, snake.body[0].y - 1}
				snake.body = append([]coord{head}, snake.body[:len(snake.body)-1]...)
			case East:
				head := coord{snake.body[0].x + 1, snake.body[0].y}
				snake.body = append([]coord{head}, snake.body[:len(snake.body)-1]...)
			case South:
				head := coord{snake.body[0].x, snake.body[0].y + 1}
				snake.body = append([]coord{head}, snake.body[:len(snake.body)-1]...)
			case West:
				head := coord{snake.body[0].x - 1, snake.body[0].y}
				snake.body = append([]coord{head}, snake.body[:len(snake.body)-1]...)
			}
		}

		// collision with borders
		if snake.body[0].x < 1 || snake.body[0].x > w || snake.body[0].y < 1 || snake.body[0].y > h {
			snake.dead = true
		}

		// collision with itself
		for i := 1; i < len(snake.body); i++ {
			if snake.body[i].x == snake.body[0].x && snake.body[i].y == snake.body[0].y {
				snake.dead = true
			}
		}

		// collision with food
		if snake.body[0].x == foodCoord.x && snake.body[0].y == foodCoord.y {
			score += 1
			// place new food
			foodCoord = coord{rand.Int() % (w - 1), rand.Int() % (h - 1)}

			for i := 0; i < 10; i++ {
				snake.body = append(snake.body, snake.body[len(snake.body)-1])
			}
		}

		// show score
		scr.moveCursor(coord{1, 1})
		scr.drawf("SCORE: %v", score)

		// draw snake
		headCell := '■'
		bodyCell := '█'
		if snake.dead {
			headCell = 'X'
			bodyCell = '░'
		}
		for _, crd := range snake.body {
			scr.setCell(bodyCell, crd)
		}
		// draw head
		scr.setCell(headCell, snake.body[0])

		// draw food
		scr.setCell('■', foodCoord)

		scr.render()

		// timing
		var (
			vertSleep = time.Duration(vertSlow)
			horzSleep = time.Duration(horzSlow)
		)
		if nitro {
			vertSleep /= 2
			horzSleep /= 2
		}

		if snake.direction%2 == 1 {
			time.Sleep(vertSleep * time.Millisecond)
		} else {
			time.Sleep(horzSleep * time.Millisecond)
		}
	}
}

type screen struct {
	w, h int
	buf  *bufio.Writer
}

func newScreen(w, h int) *screen {
	s := &screen{
		w:   w,
		h:   h,
		buf: bufio.NewWriter(os.Stdout),
	}
	s.clear()
	s.moveCursor(coord{1, 1})
	s.showCursor(false)
	s.render()
	return s
}

func (s *screen) render() {
	s.buf.Flush()
}

func (s *screen) draw(str string) {
	fmt.Fprint(s.buf, str)
}

func (s *screen) drawf(str string, args ...any) { // TODO: change func name
	fmt.Fprintf(s.buf, str, args...)
}

func (s *screen) setCell(ch rune, c coord) {
	s.moveCursor(c)
	fmt.Fprint(s.buf, string(ch))
}

func (s *screen) clear() {
	fmt.Fprint(s.buf, "\033[2J")
}

func (s *screen) moveCursor(c coord) {
	fmt.Fprintf(s.buf, "\033[%d;%dH", c.y, c.x)
}

func (s *screen) showCursor(flag bool) {
	if flag {
		fmt.Fprint(s.buf, "\033[?25h")
		return
	}
	fmt.Fprint(s.buf, "\033[?25l")
}

func (s *screen) reset() {
	s.clear()
	fmt.Fprint(s.buf, "\033[?25h") // show cursor
	fmt.Fprint(s.buf, "\033[m")    // reset colors
	fmt.Fprint(s.buf, "\033[H")    // move cursor to home position
	s.render()
}
