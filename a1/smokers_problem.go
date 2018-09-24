// =========================================
// Example 1: The Smokers Problem (Exercise 4.5)
// =========================================
// Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
// Spencer Rose (ID V00124060)

package main

import (
	"sync"
)

// Smokers loop forever, first waiting for ingredients, then making and
// smoking cigarettes. The ingredients are tobacco, paper, and matches.

func smoker (ingredient string, tobacco chan string, paper chan string,
			matches chan string, loop chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	// Pick unprovided ingredients
	switch ingredient {
	case "tobacco":
		waitForIngredients(paper, matches, loop)
	case "paper":
		waitForIngredients(tobacco, matches, loop)
	case "matches":
		waitForIngredients(tobacco, paper, loop)
	}

}

// Customer requests a haircut; if no chairs are available in the waiting room
// s/he leaves the barbershop.
func agent (tobacco chan string, paper chan string, matches chan string, loop chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		for i := 1; i <= 3; i++ {
			switch i {
			case 1:
				//fmt.Printf("agent puts out tobacco and paper\n")
				tobacco <- "tobacco"
				paper <- "paper"
			case 2:
				//fmt.Printf("agent puts out paper and matches\n")
				paper <- "paper"
				matches <- "matches"
			case 3:
				//fmt.Printf("agent puts out tobacco and matches\n")
				tobacco <- "tobacco"
				matches <- "matches"
			}
			// Wait for smoker to finish
			<-loop
		}
	}
}

func waitForIngredients(ingr1Ch chan string, ingr2Ch chan string, loop chan bool) {
	// smoker waits for other two ingredients
	for {

		select {
		case ingr1 := <-ingr1Ch:
			//fmt.Printf("smoker has %s.\n", ingr1)
			select {
			case <-ingr2Ch:
				//fmt.Printf("smoker with has all ingredients.\n")
				makeCigarette()
				smokeCigarette()
				loop <- true
				continue
			default:
				// Replace ingredient 1
				//fmt.Printf("Replace ingredient %s.\n", ingr1)
				ingr1Ch <- ingr1
				continue
			}
		case ingr2 := <-ingr2Ch:
			//fmt.Printf("smoker has %s.\n", ingr2)
			select {
			case <-ingr1Ch:
				//fmt.Printf("smoker with has all ingredients.\n")
				makeCigarette()
				smokeCigarette()
				loop <- true
				continue
			default:
				// Replace ingredient 2
				//fmt.Printf("Replace ingredient %s.\n", ingr2)
				ingr2Ch <-ingr2
				continue
			}
		}
	}
}

func makeCigarette() {
	//time.Sleep(100 * time.Millisecond)
	return
}

func smokeCigarette() {
	// time.Sleep(100 * time.Millisecond)
	return
}

func main() {

	// Buffered channels for cigarette ingredients
	tobacco := make(chan string)
	paper := make(chan string)
	matches := make(chan string)
	loop := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(4)

	//fmt.Printf("Begin agent thread\n")

	// Agent thread
	go agent(tobacco, paper, matches, loop, &wg)
	// time.Sleep(1000 * time.Millisecond)

	// Smoker threads
	go smoker("tobacco", tobacco, paper, matches, loop, &wg)
	go smoker("paper", tobacco, paper, matches, loop, &wg)
	go smoker("matches", tobacco, paper, matches, loop, &wg)

	wg.Wait()
}
