package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"./germ"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

// ROWS - rows of the network
const ROWS int = 10

// COLS - cols of the network
const COLS int = 10

// PrintGerms - Show germs
func PrintGerms(germs []*germ.Germ) {
	for {
		var eTotal uint
		for i, g := range germs {
			if i != 0 && i%COLS == 0 {
				fmt.Println()
			}
			g.Print()
			eTotal += g.GetEnergy()
		}
		fmt.Printf("\nTotal Energy: %d\n", eTotal)
		time.Sleep(200 * time.Millisecond)
	}
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}

func fill(x, y, w, h int, cell termbox.Cell) {
	for ly := 0; ly < h; ly++ {
		for lx := 0; lx < w; lx++ {
			termbox.SetCell(x+lx, y+ly, cell.Ch, cell.Fg, cell.Bg)
		}
	}
}

var colors = []termbox.Attribute{
	termbox.ColorWhite,
	termbox.ColorYellow,
	termbox.ColorGreen,
	termbox.ColorCyan,
	termbox.ColorBlue,
	termbox.ColorMagenta,
	termbox.ColorRed,
}

func getColor(e uint) termbox.Attribute {
	e += 200
	e /= 300
	if e > 6 {
		e = 6
	}

	return colors[e]
}

// PrintGermsTermBox - Show germs
func PrintGermsTermBox(germs []*germ.Germ) {
	for {
		x, y := 0, 0
		for i, g := range germs {
			bg := getColor(g.GetEnergy())
			fill(x, y, 4, 2, termbox.Cell{Bg: bg})
			tbprint(x, y, termbox.ColorBlack, bg, fmt.Sprintf("%4d", g.GetEnergy()))
			tbprint(x+1, y+1, termbox.ColorYellow, bg, fmt.Sprintf("%2d", g.GetCycle()/1000000))
			x += 4
			if (i+1)%COLS == 0 {
				x = 0
				y += 2
			}
		}
		termbox.Flush()
		time.Sleep(50 * time.Millisecond)
	}
}

// LinkGerms - Link germs into a network
// o - o - o
// | X | X |
// o - o - o
// | X | X |
// o - o - o
func LinkGerms(germs []*germ.Germ) {
	for i := 0; i < ROWS*COLS; i++ {
		// link up, except the first row.
		if i >= COLS {
			germs[i].Link(germs[i-COLS])
			// link up-left, except the first column.
			if i%COLS != 0 {
				germs[i].Link(germs[i-COLS-1])
			}
			// link up-right, except the last column.
			if (i+1)%COLS != 0 {
				germs[i].Link(germs[i-COLS+1])
			}
		}
		// link left, except the first column.
		if i%COLS != 0 {
			germs[i].Link(germs[i-1])
		}
	}
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	// Create germs
	germs := make([]*germ.Germ, 0, ROWS*COLS)
	for i := 0; i < ROWS*COLS; i++ {
		germs = append(germs, germ.NewGerm())
	}

	// Link germs
	LinkGerms(germs)

	var wg sync.WaitGroup

	// Make germs alive
	stopSignal := false
	for _, germ := range germs {
		wg.Add(1)
		go germ.Run(&wg, &stopSignal)
	}

	// Display germs
	// go PrintGerms(germs)

	// Init termbox
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	// Termbox Event Queue
	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	// Termbox redrew
	go PrintGermsTermBox(germs)

loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break loop
			default:
				if ev.Ch == 'q' || ev.Ch == 'Q' {
					break loop
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}

	termbox.Close()

	stopSignal = true
	wg.Wait()
}
