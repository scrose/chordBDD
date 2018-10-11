/*
=========================================
Example 3: Building H2O (Exercise 5.6)
Implementation A
=========================================
Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
Spencer Rose (ID V00124060)
=========================================
SUMMARY: There are two kinds of threads, oxygen and hydrogen. In order to
assemble these threads into water molecules, we have to create a barrier
that makes each thread wait until a complete molecule is ready to proceed.
As each thread passes the barrier, it should invoke bond. You must guarantee
that all the threads from one molecule invoke bond before any of the
threads from the next molecule do.

Reference: Downey, Allen B., The Little Book of Semaphores,  Version 2.2.1, pp 143-148.
*/
package main

import (
"fmt"
"os"
"strconv"
"sync"
)

type scoreboard struct {
	mtx sync.Mutex
	oCount int
	hCount int
	bondCount int
}

type barrier struct {
	mtx sync.Mutex
	ch chan bool
	n int
	count int
}

// Scoreboard constructor
func newScoreboard() *scoreboard {
	s := scoreboard{
		oCount: 0,
		hCount: 0,
		bondCount: 0,
	}
	return &s
}

// Barrier constructor
func newBarrier(threshold int) *barrier {
	b := barrier{
		n: threshold,
		count: 0,
		ch: make(chan bool, threshold),
	}
	return &b
}

func (b *barrier) Wait() {
	b.mtx.Lock()
	b.count++

	if b.count == b.n {
		close(b.ch)
		b.mtx.Unlock()
		return
	}
	b.mtx.Unlock()
	<-b.ch
}

// oxygen threads require two hydrogen threads to form H2O
func oxygen(wg *sync.WaitGroup, bondCh chan chan *barrier, sboard *scoreboard) {
	defer wg.Done()

	sboard.mtx.Lock()
	sboard.oCount++
	sboard.mtx.Unlock()

	// create barrier for bond
	barrier := newBarrier(3)
	// wait for bond request channel from H1 over bond channel
	h1 := <-bondCh
	// send barrier over bond channel to H1
	h1 <- barrier
	// wait for bond request channel from H2 over bond channel
	h2 := <-bondCh
	// send barrier over bond channel to H2
	h2 <- barrier

	sboard.mtx.Lock()
	sboard.oCount--
	sboard.bondCount++
	bond(sboard)
	sboard.mtx.Unlock()

	// close barrier
	barrier.Wait()
}

// hydrogen threads require one hydrogen and one oxygen thread to form an
// H2O bond; receives a channel on ch then sends id on channel it received
func hydrogen(wg *sync.WaitGroup, bondCh chan chan *barrier, sboard *scoreboard) {
	defer wg.Done()

	sboard.mtx.Lock()
	sboard.hCount++
	sboard.mtx.Unlock()

	// create request/response channel
	requestCh := make(chan *barrier, 1)
	// send bond request channel to Oxygen
	bondCh <- requestCh
	// receive barrier
	barrier := <-requestCh
	//bond()

	sboard.mtx.Lock()
	sboard.hCount--
	sboard.mtx.Unlock()

	// wait at barrier
	barrier.Wait()
}


func bond(sboard *scoreboard) {
	//fmt.Printf("New bond! [Bonds: %d]\n", sboard.bondCount)
	return
}

func main() {

	n, _ := strconv.Atoi(os.Args[1])
	fmt.Printf("Generating %d hydrogens and %d oxygen for n H2O molecules... \n", 2*n, n)

	// Waiting group for m customer goroutines
	var wg sync.WaitGroup

	// Oxygen channel used to receive hydrogen channel
	bondCh := make(chan chan *barrier, 1)

	// Create scoreboard
	sboard := newScoreboard()

	// Launch the oxygen threads
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go oxygen(&wg, bondCh, sboard)
	}

	// Launch the hydrogen threads
	for i := 1; i <= 2*n; i++ {
		wg.Add(1)
		go hydrogen(&wg, bondCh, sboard)
	}
	wg.Wait()
	close(bondCh)
}