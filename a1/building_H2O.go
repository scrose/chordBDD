// =========================================
// Example 3: Building H2O (Exercise 5.6)
// =========================================
// Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
// Spencer Rose (ID V00124060)
// Code modelled on examples given in Go in 5 Minutes blog:
// https://www.goin5minutes.com/blog/channel_over_channel/

package main

import (
	"fmt"
	"sync"
)

// oxygen threads require two hydrogen threads to form H2O
func oxygen(o int, wg *sync.WaitGroup, bondCh chan chan int) {
	defer wg.Done()
	// wait for bond channel with H1
	h1 := <-bondCh
	// wait for bond channel with H2
	h2 := <-bondCh
	// Confirm request on received channels
	h1 <- o
	 h1_b := <- h1
	h2 <- o
	 h2_b := <- h2
	 bond(h1_b, h2_b, o)
	//bond(1,2,o)
}

// hydrogen threads require one hydrogen and one oxygen thread to form an
// H2O bond; receives a channel on ch then sends id on channel it received
func hydrogen(h int, wg *sync.WaitGroup, bondCh chan chan int) {
	defer wg.Done()
	// create bond channel
	reqCh := make(chan int, 1)
	// send bond request and wait
	bondCh <- reqCh
	<- reqCh
	reqCh <- h
}

func bond(h1 int, h2 int, o int) {
	fmt.Printf("H_%d <> H_ %d <--> O_%d\n", h1, h2, o)
	return
}

func main() {
	var hCount = 100
	var oCount = 50

	// Waiting group for m customer goroutines
	var wg sync.WaitGroup

	// Oxygen channel to receive hydrogen channel
	bondCh := make(chan chan int)

	// Launch the oxygen threads
	for i := 1; i <= oCount; i++ {
		wg.Add(1)
		go oxygen(i, &wg, bondCh)
	}

	// Launch the hydrogen threads
	for i := 1; i <= hCount; i++ {
		wg.Add(1)
		go hydrogen(i, &wg, bondCh)
	}

	wg.Wait()
	close(bondCh)
}
