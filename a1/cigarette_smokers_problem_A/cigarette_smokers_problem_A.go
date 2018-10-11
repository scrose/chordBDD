/*
=========================================
Example 1: Cigarette Smokers Problem (Exercise 4.5)
=========================================
Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
Author: Spencer Rose
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
Profiling: https://www.programming-books.io/essential/go/
*/

package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

import _ "net/http/pprof"

type Scoreboard struct {
	n int
	tobacco int
	paper int
	matches int
	smokes int
	mtx sync.Mutex
}

type smokerTypes struct {
	tobacco chan bool
	paper chan bool
	matches chan bool
}

type helperTypes struct {
	tobacco chan bool
	paper chan bool
	matches chan bool
}

// The agent thread repeatedly chooses two different ingredients at random and makes
// them available to the smokers (via the helper).
func agent(scoreboard *Scoreboard, helpers helperTypes, wg *sync.WaitGroup) {
	defer wg.Done()
	for scoreboard.smokes < scoreboard.n {
		// Generate random selection of ingredients
		rand.Seed(time.Now().UTC().UnixNano())
		var x = rand.Intn(3) + 1
		switch x {
		case 1:
			helpers.tobacco <- true
		case 2:
			helpers.paper <- true
		case 3:
			helpers.matches <- true
		}
	}
}

// Helper transfer ingredients pushed by agent to correct smoker
func helperA(scoreboard *Scoreboard, smokers smokerTypes, helperCh chan bool) {
	for scoreboard.smokes < scoreboard.n {
		// wait for tobacco from agent
		_, more := <-helperCh
		if more {
			scoreboard.mtx.Lock()
			if scoreboard.paper > 0 {
				smokers.matches <- true
				<-smokers.matches
				scoreboard.smokes++
			} else if scoreboard.matches > 0 {
				smokers.paper <- true
				<-smokers.paper
				scoreboard.smokes++
			} else {
				scoreboard.tobacco++
			}
			scoreboard.mtx.Unlock()
		} else {
			return
		}
	}
}

// Helper transfer ingredients pushed by agent to correct smoker
func helperB(scoreboard *Scoreboard, smokers smokerTypes, helperCh chan bool) {
	for scoreboard.smokes < scoreboard.n {
		// wait for paper from agent
		_, more := <-helperCh
		if more {
			scoreboard.mtx.Lock()
			if scoreboard.tobacco > 0 {
				smokers.matches <- true
				<-smokers.matches
				scoreboard.smokes++
			} else if scoreboard.matches > 0 {
				smokers.tobacco <- true
				<-smokers.tobacco
				scoreboard.smokes++
			} else {
				scoreboard.paper++
			}
			scoreboard.mtx.Unlock()
		} else {
			return
		}
	}
}

// Helper transfer ingredients pushed by agent to correct smoker
func helperC(scoreboard *Scoreboard, smokers smokerTypes, helperCh chan bool) {
	for scoreboard.smokes < scoreboard.n {
		// wait for matches from agent
		_, more := <-helperCh
		if more {
			scoreboard.mtx.Lock()
			if scoreboard.tobacco > 0 {
				smokers.paper <- true
				<-smokers.paper
				scoreboard.smokes++
			} else if scoreboard.paper > 0 {
				smokers.tobacco <- true
				<-smokers.tobacco
				scoreboard.smokes++
			} else {
				scoreboard.matches++
			}
			scoreboard.mtx.Unlock()
		} else {
			return
		}
	}
}

func smoke(smoker string) {
	/* fmt.Printf("%s smoker finshed smoking.\n", smoker) */
	return
}

func main() {
	n, _ := strconv.Atoi(os.Args[1])
	fmt.Printf("Generating %d cigarettes... \n", n)

	smokers := smokerTypes{
		tobacco: make(chan bool, 1),
		matches: make(chan bool, 1),
		paper: make(chan bool, 1),
	}

	helpers := helperTypes{
		tobacco: make(chan bool, 1),
		matches: make(chan bool, 1),
		paper: make(chan bool, 1),
	}

	scoreboard := Scoreboard{
		tobacco: 0,
		paper: 0,
		matches: 0,
		smokes: 0,
		n: n,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Helper (Pusher) threads
	go helperA(&scoreboard, smokers, helpers.tobacco)
	go helperB(&scoreboard, smokers, helpers.paper)
	go helperC(&scoreboard, smokers, helpers.matches)

	// Agent thread
	go agent(&scoreboard, helpers, &wg)

	// Smoker threads loop forever, first waiting for ingredients, then
	// smoking cigarettes. The ingredients are tobacco, paper, and matches.
	go func() {
		for {
			_, more := <-smokers.tobacco
			if more {
				smoke("tobacco")
				smokers.tobacco <- true
			} else {
				return
			}
		}
	}()
	go func() {
		for {
			_, more := <-smokers.paper
			if more {
				smoke("paper")
				smokers.paper <- true
			} else {
				return
			}
		}
	}()
	go func() {
		for {
			_, more := <-smokers.matches
			if more {
				smoke("matches")
				smokers.matches <- true
			} else {
				return
			}
		}
	}()
	wg.Wait()
}
