package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"main/germ"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

// ROWS - rows of the network
const ROWS int = 10

// COLS - cols of the network
const COLS int = 10

func puts(s tcell.Screen, style tcell.Style, x, y int, msg string) {
	for _, c := range msg {
		s.SetCell(x, y, style, c)
		x += runewidth.RuneWidth(c)
	}
}

func fill(s tcell.Screen, style tcell.Style, x, y, w, h int) {
	for ly := 0; ly < h; ly++ {
		for lx := 0; lx < w; lx++ {
			s.SetCell(x+lx, y+ly, style)
		}
	}
}

func calcColor(e uint) tcell.Color {
	var r, g, b, v int32 = 0, 0, 0, int32(e)
	if v < 100 { // yellow - turquoise
		r = (100 - v) * 200 / 100 // 200 -   0
		g = (100-v)*55/100 + 200  // 255 - 200
		b = v * 255 / 100         //   0 - 255
	} else if v < 2000 { // blue - pink
		r = (v - 100) * 255 / 1900 //   0 - 255
		g = 0                      //   0
		b = 255                    // 255
	} else if v < 4000 { // pink - red
		r = 255                     // 255
		g = 0                       //   0
		b = (4000 - v) * 255 / 2000 // 255 - 0
	} else { // red
		r, g, b = 255, 0, 0
	}

	return tcell.NewRGBColor(r, g, b)
}

func printGerms(s tcell.Screen, germs []*germ.Germ) {
	for {
		x, y := 10, 5
		for i, g := range germs {
			style := tcell.StyleDefault.
				Foreground(tcell.ColorWhite).
				Background(calcColor(g.GetEnergy()))
			fill(s, style, x, y, 4, 2)
			puts(s, style, x, y, fmt.Sprintf("%4d", g.GetEnergy()))
			puts(s, style, x+1, y+1, fmt.Sprintf("%2d", g.GetCycle()/1000000))
			x += 4
			if (i+1)%COLS == 0 {
				x = 10
				y += 2
			}
		}
		s.Show()
		time.Sleep(10 * time.Millisecond)
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

	// Init screen
	s, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err = s.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorBlack))
	s.Clear()

	// Termbox redrew
	go printGerms(s, germs)

loop:
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyEnter:
				break loop
			default:
				if ev.Rune() == 'q' || ev.Rune() == 'Q' {
					break loop
				}
			}
		case *tcell.EventResize:
			s.Sync()
		}
	}

	s.Fini()

	stopSignal = true
	wg.Wait()
}
