/*
=========================================
Example 2: The Barbershop Problem (Exercise 5.2)
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
*/

package main

import (
	"fmt"
	"strings"
	"sync"
)

type BarbershopComm struct {
	customers chan int
	barber chan bool
}

// BarbershopComm constructor
func NewBarbershopComm(n int) *BarbershopComm {
	b := BarbershopComm{}
	b.customers = make(chan int, n)
	b.barber = make(chan bool)
	return &b
}


func main() {
	var customerCount = 25
	var n = 4

	// Waiting group for m customer goroutines
	var wg sync.WaitGroup

	// Create Barber-Customer buffered (n) channels
	shop := NewBarbershopComm(n)

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
	go func (shop *BarbershopComm, c chan int) {
		// Wait for next customer
		for {
			select {
			case <-shop.customers:
				c <- 1
				<- shop.doneCut
				c <- -1
			}
		}
	}(shop, bCh)

	// Queue the customers
	for i := 0; i < customerCount; i++ {
		wg.Add(1)
		go func (shop *BarbershopComm, wg *sync.WaitGroup, c chan int) {
			defer wg.Done()
			// Wait for the barber
			select {
			case shop.customers <- 1:
				c <- 1
				<-shop.barberDone
				c <- -1
			default:
				return
				}
		}(shop, &wg, cCh)
	}
	wg.Wait()
}
