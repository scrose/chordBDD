/*
=========================================
Example 4: Baboon Crossing Problem (Exercise 6.3
=========================================
Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
Spencer Rose (ID V00124060)

*/

package main

import (
	"log"
	"sync"
	"time"
)

type Rope struct {
	resp chan int
}

// baboon thread routine
func baboon(i int, wg *sync.WaitGroup, dirCh chan Rope, actionCh chan int) {
	defer wg.Done()
	resp := make(chan int)

	// send request for the rope
	dirCh <- Rope{resp}
	log.Printf("Baboon %d is ready to cross...\n", i)
	<-resp
	log.Printf("Baboon %d is crossing.\n", i)
	time.Sleep(time.Second)
	actionCh <- i
}

func main() {
	var n = 3

	// Waiting group for baboon threads
	var wg sync.WaitGroup

	// channels represent crossing direction and state
	westCh := make(chan Rope)
	eastCh := make(chan Rope)
	actionCh := make(chan int)

	// Launch 2n baboon threads
	for i := 1; i <= n; i++ {
		wg.Add(2)

		go baboon(i, &wg, eastCh, actionCh)
		//go baboon(i*10, &wg, westCh, actionCh)
	}

	// Launch rope thread
	// go func() {
		//wg.Add(1)
		var dir = ""
		var crossing = 0

		// loop
		log.Println("Waiting for a baboon.")
		for {
			log.Println("loop!")
			// check number and direction of baboons crossing
				// wait for baboon to start/finish crossing
				select {
				case b := <-eastCh:
					if crossing < 5 && (dir == "east" || dir == "") {
						crossing++
						dir = "east"
						b.resp <- 1
					}


				case b := <-westCh:
					if crossing < 5 && (dir == "west" || dir == "") {
						crossing++
						dir = "west"
						b.resp <- 1
					}

				case b := <-actionCh:
					log.Printf("Baboon %d finished crossing %s.\n", b, dir)
					crossing--
					if crossing == 0 {
						dir = ""
						log.Println("Waiting for a baboon.")
					}

				}
			wg.Wait()

		}
	//}()
	close(eastCh)
	close(westCh)
	close(actionCh)


}

