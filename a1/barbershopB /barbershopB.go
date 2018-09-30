/*
~barbershopB.go
=========================================
Example 2: The Barbershop Problem (Exercise 5.2)
Implementation B
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
Emulating semaphores: http://www.golangpatterns.info/concurrency/semaphores
*/

package main

import (
	"fmt"
	"strings"
	"sync"
)

type Barbershop struct {
	room sync.Mutex
	customers int
	barber semaphore
	barberDone semaphore
	customer semaphore
	customerDone semaphore
}

// Barbershop constructor
func NewBarbershop() *Barbershop {
	b := Barbershop{
		customers: 0,
	}
	b.barber = make(semaphore, 1)
	b.barberDone = make(semaphore, 1)
	b.customer = make(semaphore, 1)
	b.customerDone = make(semaphore, 1)
	return &b
}

// Emulate semaphores using channels
type empty struct {}
type semaphore chan empty

// Semaphore: Acquire n resources (wait)
func (s semaphore) P(n int) {
	e := empty{}
	for i := 0; i < n; i++ {
		s <- e
	}
}

// Semaphore: Release n resources (signal)
func (s semaphore) V(n int) {
	for i := 0; i < n; i++ {
		<-s
	}
}

func main() {

	var customerCount = 90000

	// Number of waiting room seats
	var n = 4

	// Waiting group for customer goroutines
	var wg sync.WaitGroup

	// Create shop instance
	shop := NewBarbershop()

	// Tracker thread
	var b, c int
	bCh := make(chan int)
	cCh := make(chan int)
	go func() {
		for {
			select {
			case n := <-bCh:
				b += n
			case n := <-cCh:
				c += n
			}
			fmt.Printf("%s|%s\n",
				strings.Repeat("B", b),
				strings.Repeat("C", c))
		}
	}()

	// Barber goroutine
	go func(wg *sync.WaitGroup, shop *Barbershop, n int, c chan int) {
		for {
			// Wait for next customer
			shop.customer.P(1)
			// Signal barber is with a customer
			shop.barber.V(1)
			//fmt.Printf("Barber is busy...\n")
			c <- 1
			shop.customerDone.P(1)
			shop.barberDone.V(1)
			//fmt.Printf("Barber is done...\n")
			c <- -1
		}
	}(&wg, shop, n, bCh)

	// Customer goroutines
	for i := 1; i <= customerCount; i++ {
		wg.Add(1)

		// Customer goroutine
		go func (wg *sync.WaitGroup, cID int, shop *Barbershop, n int, c chan int) {
			defer wg.Done()

			// Check that waiting room is not full
			shop.room.Lock()
			if shop.customers < n {
				shop.customers++
				c <- 1
			} else { // Balk (leave) if full
				shop.room.Unlock()
				//fmt.Printf("Waiting room full! Customer #%d leaves.\n", cID)
				return
			}
			shop.room.Unlock()

			// Signal customer is ready
			shop.customer.V(1)
			shop.barber.P(1)
				//fmt.Printf("Customer #%d is getting a haircut.\n", cID)
			shop.customerDone.V(1)
			shop.barberDone.P(1)
			shop.room.Lock()
				shop.customers--
				c <- -1
			shop.room.Unlock()

		}(&wg, i, shop, n, cCh)
	}
	wg.Wait()
}
