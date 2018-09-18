// =========================================
// Example 2: The Search-Insert-Delete Problem (Exercise 6.1)
// =========================================
// Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
// Spencer Rose (ID V00124060)

package main

import (
	"fmt"
	"time"
	"math/rand"
	"sync"
)

// Searchers merely examine the list; hence they can execute concurrently
// with each other.
func searcher(queue chan int, done chan bool, stop chan bool) {
	// waiting for next customer
	for {
		select {
			// Take next customer in waiting room
		case <-stop:
			return
		default:
			c := <-queue
			cutHair(c)
			done <- true
		}
  }
}

// Inserters add new items to the end of the list; insertions must be mutually
// exclusive to preclude two inserters from inserting new items at about the
// same time. However, one insert can proceed in parallel with any number of
// searches.
func inserter(c int, wg *sync.WaitGroup, queue chan int, done chan bool) {
	defer wg.Done()
	fmt.Printf("Customer #%d arrives. Waiting room has %d seat(s) available.\n", c, cap(queue)-len(queue))
	select {
	case queue <- c:
		getHairCut(c)
		<-done
		fmt.Printf("Haircut done. Customer #%d leaves\n", c)
	default:
		fmt.Printf("No seats! Customer #%d leaves\n", c)
	}
}

// deleters remove items from anywhere in the list. At most one deleter process
// can access the list at a time, and deletion must also be mutually exclusive
// with searches and insertions.

func deleter(c int) {
	fmt.Printf("Customer #%d is getting a haircut.\n", c)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Done with customer #%d!\n", c)
}

func getHairCut(c int) {
	fmt.Printf("Customer #%d sits in barber chair.\n", c)
	time.Sleep(100 * time.Millisecond)
}

func main() {
	var count = 25
	var n = 4

	// Waiting group for m customer goroutines
	var wg sync.WaitGroup
	wg.Add(count)

	// Barber-Customer buffered communication channel
	queue := make(chan int, n)
	done := make(chan bool)
	stop := make(chan bool)

	fmt.Printf("Barbershop is open with %d chairs in the waiting room and one barber's chair...\n", n)

	// Get the barber ready
	go barber(queue, done, stop)
	time.Sleep(time.Second)

	// Queue the customers
	for i := 1; i < count + 1; i++ {
		r := rand.Intn(50)
		time.Sleep(time.Duration(r) * time.Millisecond)
		go customer(i, &wg, queue, done)
	}
	wg.Wait()
	fmt.Printf("Closing the barbershop... ")
	close(stop)
	fmt.Printf("done.\n")
}
