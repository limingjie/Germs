package germ

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Germ - A germ
type Germ struct {
	id      uint
	active  bool
	energy  uint          // The energy of the germ
	cycle   time.Duration // Heartbeat cycle
	input   chan uint     // The channel to absorb energy
	outputs []*chan uint  // The channels of surrounding germs
	mux     sync.Mutex
}

// Absorb - Absorbs energy emit from surrounding germs
func (g *Germ) Absorb(energy uint) {
	g.mux.Lock()
	g.energy += energy
	g.mux.Unlock()
}

// Emit - Emits energy to surrounding germs
func (g *Germ) Emit() {
	// e := rand.Intn(3) + 3 // 3 - 5
	e := 1

	canEmit := false
	g.mux.Lock()
	eTotal := uint(e * len(g.outputs))
	if g.energy >= eTotal {
		canEmit = true
		g.energy -= eTotal
	}
	g.mux.Unlock()

	if canEmit {
		for _, out := range g.outputs {
			*out <- uint(e)
		}
	}
}

// Run - Make it alive
func (g *Germ) Run(wg *sync.WaitGroup, stopSignal *bool) {
	defer wg.Done()

	heartbeat := time.Tick(g.cycle)
	for {
		select {
		case e := <-g.input:
			g.Absorb(e)
		case <-heartbeat:
			if g.active {
				g.Emit()
			}
		}
		if *stopSignal {
			// fmt.Printf("Germ %d quit.\n", g.id)
			return
		}
	}
}

// Link - Link 2 germs
func (g *Germ) Link(o *Germ) {
	g.outputs = append(g.outputs, &o.input)
	o.outputs = append(o.outputs, &g.input)
}

// Print - Print germ
func (g *Germ) Print() {
	// fmt.Printf("%03d ", g.energy)
	fmt.Printf("%04d ", g.energy)
}

// GetID - Return ID
func (g *Germ) GetID() uint {
	return g.id
}

// GetEnergy - Return energy
func (g *Germ) GetEnergy() uint {
	return g.energy
}

// GetCycle - Return cycle
func (g *Germ) GetCycle() time.Duration {
	return g.cycle
}

var counter uint = 0
var mux sync.Mutex

// NewGerm - Create a Germ
func NewGerm() *Germ {
	var id uint
	mux.Lock()
	id = counter
	counter++
	mux.Unlock()

	return &Germ{
		id:      id,
		active:  true,
		energy:  300,
		cycle:   time.Duration(rand.Intn(20)+1) * time.Millisecond,
		input:   make(chan uint, 10),
		outputs: make([]*chan uint, 0, 8),
	}
}
