/*
=========================================
Example 4: Baboon Crossing Problem (Exercise 6.3
=========================================
Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
Spencer Rose (ID V00124060)

*/

package main

import (
	"sync"
	"time"
)

type BaboonCh struct {
	req chan int
	cross chan int
	done chan int
}

// BaboonCh constructor
func NewBaboonCh() *BaboonCh {
	a := BaboonCh{}
	a.req = make(chan int, 1)
	a.cross = make(chan int, 1)
	a.done = make(chan int, 1)
	return &a
}

// baboon thread routine
func baboon(i int, wg *sync.WaitGroup, rope BaboonCh) {
	defer wg.Done()

	// send request for the rope
	rope.req <- i
	// wait for access to the rope
	<-rope.cross
	//log.Printf("Baboon %d is crossing.\n", i)
	// signal end of crossing
	rope.done <- i
}

func main() {
	var n = 10

	// Create tracker
	trk := *NewTracker()
	tracker(trk)
	time.Sleep(1)

	// Waiting group for baboon and rope threads
	var wg sync.WaitGroup
	//wg.Add(1)

	// Channels representing crossing direction and state
	eastCh := *NewBaboonCh()
	westCh := *NewBaboonCh()

	// Launch rope thread
	go func() {
		// Counters
		var eastCrossing = 0
		var eastWaiting = 0
		var westCrossing = 0
		var westWaiting = 0
		var i = 0

		// Lightswitch for rope
		var ropeLock = 0

		for {
			select {
			// Eastbound baboon waiting to cross
			case <-eastCh.req:
				if westCrossing == 0 && eastCrossing < 5 && ropeLock == 0 {
					eastCrossing++
					trk.reCh <- 1
					eastCh.cross <- 1
					// Rope is at capacity: empty the rope
					if eastCrossing == 5 {
						ropeLock = 1
					}
				} else {
					eastWaiting++
					trk.eCh <- 1
				}

			// Westbound baboon waiting to cross
			case <-westCh.req:
				if westCrossing < 5 && eastCrossing == 0 && ropeLock == 0 {
					westCrossing++
					trk.rwCh <- 1
					westCh.cross <- 1
					// 5th baboon turns rope lock 'on'
					if westCrossing == 5 {
						ropeLock = 1
					}
				} else {
					westWaiting++
					trk.wCh <- 1
				}

			// Eastbound baboon finished crossing
			case <-eastCh.done:
				eastCrossing--
				// Turn rope lock 'off'
				if eastCrossing == 0 {
					ropeLock = 0
				}
				trk.reCh <- -1
				// Release west waiting list
				for i = 0; i < 5 && westWaiting > 0 && eastCrossing == 0; i++ {
					westCh.cross <- 1
					westWaiting--
					westCrossing++
					trk.wCh <- -1
					trk.rwCh <- 1
				}

			// Westbound baboon finished crossing
			case <-westCh.done:
				westCrossing--
				// Turn rope lock 'off'
				if westCrossing == 0 {
					ropeLock = 0
				}
				trk.rwCh <- -1
				// Release east waiting list
				for i = 0; i < 5 && eastWaiting > 0 && westCrossing == 0; i++ {
					eastCh.cross <- 1
					eastWaiting--
					eastCrossing++
					trk.eCh <- -1
					trk.reCh <- 1
				}
			// Check waitlists if no requests are available
			default:
				if westWaiting > eastWaiting {
					// Release west waiting list
					for i = 0; i < 5 && westWaiting > 0 && eastCrossing == 0; i++ {
						westCh.cross <- 1
						westWaiting--
						westCrossing++
						trk.wCh <- -1
						trk.rwCh <- 1
					}
				} else {
					// Release east waiting list
					for i = 0; i < 5 && eastWaiting > 0 && westCrossing == 0; i++ {
						eastCh.cross <- 1
						eastWaiting--
						eastCrossing++
						trk.eCh <- -1
						trk.reCh <- 1
					}

				}
			}
		}
	}()

	// Launch westward and eastward baboon threads
	for i := 1; i <= n; i++ {
		wg.Add(2)
		go baboon(i, &wg, eastCh)
		go baboon(i, &wg, westCh)
	}

	wg.Wait()

}

