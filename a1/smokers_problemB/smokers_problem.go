/*
=========================================
Example 1: Cigarette Smokers Problem (Exercise 4.5)
=========================================
Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
Spencer Rose (ID V00124060)
=========================================
SUMMARY: Four threads are involved: an agent and three smokers. The smokers loop
forever, first waiting for ingredients, then making and smoking cigarettes. The
ingredients are tobacco, paper, and matches. We assume that the agent has an infinite
supply of all three ingredients, and each smoker has an infinite supply of one of the
ingredients; that is, one smoker has matches, another has paper, and the third has
tobacco. The agent repeatedly chooses two different ingredients at random and makes
them available to the smokers. Depending on which ingredients are chosen, the smoker
with the complementary ingredient should pick up both resources and proceed.

Reference: Downey, Allen B., The Little Book of Semaphores,  Version 2.2.1, pp 101-111.
*/

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ingredientTypes struct {
	tobacco bool
	paper bool
	matches bool
}

type smokerTypes struct {
	tobacco chan bool
	paper chan bool
	matches chan bool
}

// The agent thread repeatedly chooses two different ingredients at random and makes
// them available to the smokers (via the helper).
func agent(ingCh chan ingredientTypes, signal chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		// Generate random selection of ingredients
		rand.Seed(time.Now().UTC().UnixNano())
		var x, i, j, k = rand.Intn(2) + 1, false, false, false
		switch x {
		case 1:
			i, j, k = true, true, false
		case 2:
			i, j, k = true, false, true
		case 3:
			i, j, k = false, true, true
		}

		ingredients := ingredientTypes{
			tobacco: i,
			matches: j,
			paper: k,
		}
		ingCh <- ingredients
		<-signal
		fmt.Printf("Next round...\n")
	}

}

// Helper transfer ingredients pushed by agent to correct smoker
func helper(ingCh chan ingredientTypes, smokers smokerTypes, signal chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	for {

		// wait for ingredients from agent
		ing := <-ingCh
		fmt.Printf("tobacco: %t\n", ing.tobacco)
		fmt.Printf("paper: %t\n", ing.paper)
		fmt.Printf("matches: %t\n", ing.matches)

		if ing.tobacco && ing.paper  {
			// Signal "matches" smoker
			fmt.Printf("Signal matches smoker.\n")
			smokers.matches <- true
			<-smokers.matches
		} else if ing.tobacco && ing.matches {
			// Signal "paper" smoker
			fmt.Printf("Signal paper smoker.\n")
			smokers.paper <- true
			<-smokers.paper
		} else if ing.paper && ing.matches {
			// Signal "tobacco" smoker
			fmt.Printf("Signal tobacco smoker.\n")
			smokers.tobacco <- true
			<-smokers.tobacco
		} else {
			continue
		}
		signal<- true
	}
}

func smoke() {
	// time.Sleep(100 * time.Millisecond)
	return
}

func main() {

	ingredients := make(chan ingredientTypes)
	smokers := smokerTypes{
		tobacco: make(chan bool),
		matches: make(chan bool),
		paper: make(chan bool),
	}
	signal := make(chan bool)


	var wg sync.WaitGroup
	wg.Add(5)

	// Agent thread
	go agent(ingredients, signal, &wg)

	// Helper (Pusher) thread
	go helper(ingredients, smokers, signal, &wg)

	// Smoker threads loop forever, first waiting for ingredients, then
	// smoking cigarettes. The ingredients are tobacco, paper, and matches.
	go func() {
		defer wg.Done()
		for {
			// Wait for signal from helper
			<-smokers.tobacco
			smoke()
			fmt.Printf("Tobacco smoker finshed smoking.\n")
			smokers.tobacco <- true
		}
	}()
	go func() {
		defer wg.Done()
		for {
			// Wait for signal from helper
			<-smokers.paper
			smoke()
			fmt.Printf("Paper smoker finshed smoking.\n")
			smokers.paper <- true
		}
	}()
	go func() {
		defer wg.Done()
		for {
			// Wait for signal from helper
			<-smokers.matches
			smoke()
			fmt.Printf("Matches smoker finshed smoking.\n")
			smokers.matches <- true
		}
	}()

	wg.Wait()
}
