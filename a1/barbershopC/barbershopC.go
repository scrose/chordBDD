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
	barber Semaphore
	barberDone Semaphore
	customer Semaphore
	customerDone Semaphore
}

// Barbershop constructor
func NewBarbershop() *Barbershop {
	b := Barbershop{
		room:         sync.Mutex{},
		customers:    0,
	}
	return &b
}

// Emulate semaphores using channels
type Semaphore struct {
	sync.Mutex
	cond *sync.Cond
	n, wakeups int
}

// Semaphore constructor
func NewSemaphone() *Semaphore {
	s := Semaphore{
		n: 0,
		wakeups: 0,
	}
	s.cond = sync.NewCond(&s)
	return &s
}

// Semaphore: Wait (P) (adapted from Downey, p. 265)
func (s *Semaphore) P() {
	s.cond.L.Lock()
	s.n--
	if s.n < 0 {
		for s.wakeups < 1 {
			s.cond.Wait()
		}
		s.wakeups--
	}
	s.cond.L.Unlock()
}

// Semaphore: Signal (V) (adapted from Downey, p. 265)
func (s Semaphore) V() {
	s.cond.L.Lock()
	s.n++
	if s.n <= 0 {
		s.wakeups++
		s.cond.Signal()
	}
	s.cond.L.Unlock()
}

func main() {

	var customerCount = 90000

	// Number of waiting room seats
	var n = 4

	// Waiting group for customer goroutines
	var wg sync.WaitGroup

	// Create shop instance
	shop := NewBarbershop()

	barber := NewSemaphone()
	barberDone := NewSemaphone()
	customer := NewSemaphone()
	customerDone := NewSemaphone()

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
			fmt.Printf("Barber is waiting...\n")
			customer.P()

			// Signal barber is with a customer
			barber.V()
			fmt.Printf("Barber is busy...\n")
			c <- 1
			customerDone.P()
			barberDone.V()
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
			customer.V()
			barber.P()
				//fmt.Printf("Customer #%d is getting a haircut.\n", cID)
			customerDone.V()
			barberDone.P()
			shop.room.Lock()
				shop.customers--
				c <- -1
			shop.room.Unlock()

		}(&wg, i, shop, n, cCh)
	}
	wg.Wait()
}
