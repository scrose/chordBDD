/*
=========================================
Example 2: The Barbershop Problem (Exercise 5.2)
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
*/

package main

import (
	"fmt"
	"time"
	"math/rand"
	"sync"
)

// Barber receives next customer and cuts her/his hair
func barber(queue chan int, cut_done chan bool, stop chan bool) {
	// waiting for next customer
	for {
		select {
		case <-stop:
			return
		// Take next customer in waiting room
		default:
			c := <-queue
			cutHair(c)
			cut_done <- true
		}
  }
}

// Customer requests a haircut; if no chairs are available in the waiting room
// s/he leaves the barbershop.
func customer(c int, wg *sync.WaitGroup, queue chan int, cut_done chan bool) {
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

func cutHair(c int) {
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
	cut_done := make(chan bool)
	cust_ready := make(chan bool)
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
