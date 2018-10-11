/*
~barbershopB.go
=========================================
Example 2: FIFO Barbershop Problem (Exercise 5.3)
Implementation A
=========================================
Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
Spencer Rose (ID V00124060)
=========================================
SUMMARY: A barbershop consists of a waiting room with n chairs, and the barber room
containing the barber chair. If there are no customers to be served, the barber goes
to sleep. If a customer enters the barbershop and all chairs are occupied, then the
customer leaves the shop. If the barber is busy, but chairs are available, then the
customer sits in one of the free chairs. If the barber is asleep, the customer wakes
up the barber. Write a program to coordinate the barber and the customers.

References
Downey, Allen B., The Little Book of Semaphores,  Version 2.2.1, pp 101-111.
Implementation is an adaption of Downey's 5.2.2 Barbershop solution
*/

package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {

	customerCount, _ := strconv.Atoi(os.Args[1])
	fmt.Printf("Generating %d customers... \n", customerCount)

	// Number of waiting room seats
	var n = 4

	// Waiting group for customer goroutines
	var wg sync.WaitGroup

	// Initialize the customer queue
	q := make(chan chan int, n)

	// Create tracker thread
	// lightweight console printout to show interleaving of goroutines
	trk := *newTracker()
	runTracker(trk)

	// Barber goroutine
	go func(q chan chan int, trk tracker) {
		for {
			select {
				// wait for next customer in the queue (waiting room)
				case customer := <-q:
				// barber signals customer
				customer <- 1
				// wait for response from customer
				id := <-customer
				trk.ch <- event{time.Now(), id, "dequeue"}
				trk.ch <- event{time.Now(), id, "cut hair"}

				// cutHair()

				// barber signals haircut is done
				customer <- 1
				trk.ch <- event{time.Now(), id, "done haircut"}
				// barber waits for customer to leave the shop
				<- customer
				trk.ch <- event{time.Now(), id, "exit"}
			}
		}
	}(q, trk)

	// Customer goroutines
	for i := 0; i < customerCount; i++ {
		wg.Add(1)

		go func(wg *sync.WaitGroup, id int, q chan chan int, trk tracker) {
			defer wg.Done()

			// Create a customer channel
			customer := make(chan int)
			// Enqueue customer
			select {
			// Customer enqueued (buffered channel balks when full)
			case q <- customer:
				trk.ch <- event{time.Now(), id, "enqueue"}
				break
			// Otherwise waiting room (queue) is full: customer leaves
			default:
				trk.ch <- event{time.Now(), id, "balk"}
				return
			}

			/* Wait for barber to signal */
			<-customer
			// Respond to barber
			customer <- id
			// getHaircut()
			// Wait for barber to finish
			<-customer
			close(customer)
		}(&wg, i, q, trk)
	}
	wg.Wait()
}
