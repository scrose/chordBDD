package main

import (
	"fmt"
	"strings"
)

type Tracker struct {
	e int
	w int
	re int
	rw int
	wCh chan int
	eCh chan int
	reCh chan int
	rwCh chan int
}

// tracker constructor
func NewTracker() *Tracker {
	t := Tracker{
		e: 0,
		w: 0,
		rw: 0,
		re: 0,
	}
	t.wCh = make(chan int)
	t.eCh = make(chan int)
	t.rwCh = make(chan int)
	t.reCh = make(chan int)
	return &t
}

func tracker(t Tracker) {
	go func() {
		for {
			select {
			case n := <-t.eCh:
				t.e += n
			case n := <-t.wCh:
				t.w += n
			case n := <-t.reCh:
				t.re += n
			case n := <-t.rwCh:
				t.rw += n
			}
			fmt.Printf("%s|%s%s|%s\n",
				strings.Repeat("E", t.e),
				strings.Repeat("E", t.re),
				strings.Repeat("W", t.rw),
				strings.Repeat("W", t.w))
		}
	}()
}