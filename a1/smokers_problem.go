// =========================================
// Example 1: Cigarette Smokers Problem (Exercise 4.5)
// =========================================
// Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
// Spencer Rose (ID V00124060)
// =========================================
// SUMMARY: Four threads are involved: an agent and three smokers. The smokers loop
// forever, first waiting for ingredients, then making and smoking cigarettes. The
// ingredients are tobacco, paper, and matches. We assume that the agent has an infinite
// supply of all three ingredients, and each smoker has an infinite supply of one of the
// ingredients; that is, one smoker has matches, another has paper, and the third has
// tobacco. The agent repeatedly chooses two different ingredients at random and makes
// them available to the smokers. Depending on which ingredients are chosen, the smoker
// with the complementary ingredient should pick up both resources and proceed.
// Reference: Downey, Allen B., The Little Book of Semaphores,  Version 2.2.1, pp 101-111.

package main

import (
	"sync"
)


// Customer requests a haircut; if no chairs are available in the waiting room
// s/he leaves the barbershop.
func agent (tobacco chan string, paper chan string, matches chan string, loop chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		for i := 1; i <= 3; i++ {
			switch i {
			case 1:
				// Agent puts out tobacco and paper
				tobacco <- "tobacco"
				paper <- "paper"
			case 2:
				// Agent puts out paper and matches
				paper <- "paper"
				matches <- "matches"
			case 3:
				// Agent puts out tobacco and matches
				tobacco <- "tobacco"
				matches <- "matches"
			}
			// Wait for a smoker to finish, then loop
			<-loop
		}
	}
}

func waitForIngredients(ingr1Ch chan string, ingr2Ch chan string, loop chan bool) {
	for {
		// Smoker waits for other two ingredients
		select {
		case ingr1 := <-ingr1Ch:
			// Needed ingredient acquired; acquire third ingredient
			select {
			case <-ingr2Ch:
				// Smoker has all the ingredients to smoke
				makeCigarette()
				smokeCigarette()
				loop <- true
				continue
			default:
				// Replace ingredient since third not available
				ingr1Ch <- ingr1
				continue
			}
		case ingr2 := <-ingr2Ch:
			// Needed ingredient acquired; acquire third ingredient
			select {
			case <-ingr1Ch:
				// Smoker has all the ingredients to smoke
				makeCigarette()
				smokeCigarette()
				loop <- true
				continue
			default:
				// Replace ingredient since third not available
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

	// Channels to "transfer" cigarette ingredients
	tobacco := make(chan string)
	paper := make(chan string)
	matches := make(chan string)
	// Channel to notify agent when smoker has taken two ingredients
	loop := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(4)

	// Agent thread
	go agent(tobacco, paper, matches, loop, &wg)
	// time.Sleep(1000 * time.Millisecond)

	// Smokers threads loop forever, first waiting for ingredients, then making and
	// smoking cigarettes. The ingredients are tobacco, paper, and matches.
	go func() {
		defer wg.Done()
		waitForIngredients(tobacco, paper, loop)
	}()
	go func() {
		defer wg.Done()
		waitForIngredients(tobacco, matches, loop)
	}()
	go func() {
		defer wg.Done()
		waitForIngredients(paper, matches, loop)
	}()

	wg.Wait()
}
