package main

import (
	"time"
)

type event struct {
	ts time.Time
	id int
	action string
}

type data struct {
	events []event
}

type tracker struct {
	ch chan event // tracker channel
	actions []string
}

// tracker constructor
func newTracker() *tracker {
	return &tracker{
		ch: make(chan event),
	}
}

func runTracker(t tracker) {
	d := data{}
	go func() {
		for {
			select {
			case e := <-t.ch:
				d.events = append(d.events, e)
				printEvent(e)
			}

		}
	}()
}

func printEvent(e event) {
	// fmt.Printf("%-25d %-15s %-10d\n", e.ts.UnixNano(), e.action, e.id)
}