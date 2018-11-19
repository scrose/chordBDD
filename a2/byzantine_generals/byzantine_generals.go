/*
=========================================
Byzantine Generals Problem (Assignment 2.2)
=========================================
Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
Author: Spencer Rose
=========================================
SUMMARY: If a general is loyal, and it has to relay an order O to general Gi, it always relays the value O.
If a general is a traitor, and it has to relay an order O to general Gi, it will relay the value O if i is odd,
and it will send the opposite of O if i is even. Note that, in the case of a traitorous commander general,
the order being relayed is still just Oc.

When performing a majority vote, only ATTACK and RETREAT votes are taken into account, and it is enough for one of
the two to have a relative majority (i.e., a plurality) to determine what action wins the vote. If the number of
ATTACK and RETREAT votes is the same, then the result of the vote is a tie.

PARAMETERS:
 - m: the level of recursion in the algorithm, and assume m > 0.
 - G: List of n generals: [G0,G1,G2,...,Gn] each of which is either loyal or a traitor. G0 is the commander,
      and the remaining n−1 generals are the lieutenants.
 - Oc: The order the commander gives to its lieutenants (Oc in {ATTACK,RETREAT})

Reference:
[1] Leslie Lamport, Robert Shostak, And Marshall Peasethe, Byzantine Generals Problem, SRI International, 1982.
[2] Sets in Golang: https://www.davidkaya.sk/2017/12/10/sets-in-golang/
*/

package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// -----------------------------------------
// Enum translators
const (
	RETREAT = 0
	ATTACK = 1
	TIE = 2
	TRAITOR = 0
	LOYAL = 1
)
var loyaltyKey = map[string]int{"Traitor": TRAITOR, "Loyal": LOYAL}
var loyaltyType = map[int]string{TRAITOR : "Traitor", LOYAL: "Loyal"}
var orderKey = map[string]int{"Attack": ATTACK, "Retreat": RETREAT, "Tie": TIE}
var orderType = map[int]string{ATTACK : "Attack", RETREAT: "Retreat", TIE: "Tie"}

// Array of channels
type vectorChannel struct {
	chs []chan Message
}

// Vector channel constructor (one ch per goroutine)
func NewVectorChannel(n int, m int) *vectorChannel {
	vch := vectorChannel{}
	for i := 0; i < n; i++ {
		vch.chs = append(vch.chs, make(chan Message, calcExpected(m,n - 2)))
	}
	return &vch
}

func calcExpected(m int, ttl int) int {
	if m > 1 {
		ttl += calcExpected(m - 1, ttl * (ttl - 1))
	}
	return ttl
}

// -----------------------------------------
// Oral message
type Message struct {
	id int
	rnd int
	order int
	senders []int
}
// Tree node
type Node struct {
	index int
	id int
	order int
	senders []int
	children map[int]Node
	level int
}
// Message scheduler
type Scheduler struct {
	queue map[int][]Message
}

// -----------------------------------------
// Set Interface
var exists = struct{}{}

type set struct {
	m map[int]struct{}
}

func NewSet() *set {
	s := &set{}
	s.m = make(map[int]struct{})
	return s
}
// initialize set with range of values
func (s *set) init(i int, j int) {
	for value := i; value < j; value++ {
		s.m[value] = exists
	}
}

func (s *set) add(value int) {
	s.m[value] = exists
}

func (s *set) remove(value int) {
	delete(s.m, value)
}

func (s *set) contains(value int) bool {
	_, c := s.m[value]
	return c
}


// -----------------------------------------
// Count traitors in generals
func counter(G []string) map[string]int {
	counter := make( map[string]int )
	for i, label := range G {
		if label == "Loyal" || label == "Traitor"{
			if i > 0 {counter[label]++}
		} else {
			fmt.Printf("ERROR: Input \"%s\" is not correctly labeled.", label)
			os.Exit(1)
		}
	}
	return counter
}

// Calculate majority consensus of orders
func majority(orders []int) int {
	counter := make( map[int]int )
	for _, i := range orders {
		counter[i]++
	}
	if counter[ATTACK] > counter[RETREAT] {
		return ATTACK
	} else if counter[RETREAT] > counter[ATTACK] {
		return RETREAT
	} else {
		return TIE
	}
}

func getConsensus(node Node) int {
	var election []int
	time.Sleep(50*time.Millisecond)
	if len(node.children) == 0 {
		return node.order
	}
	for _, u := range node.children {
		election = append(election, getConsensus(u))
	}
	return majority(election)
}

// Calculate majority vote consensus / print recursion tree
func consensus(tree Node) string {
	var v Node
	var stk = []Node{tree}
	visited := make(map[int]bool)
	var output = ""

	for len(stk) > 0 {
		v = stk[0] // pop stack
		stk = stk[1:]
		level := len(v.senders)

		if !visited[v.index] {
			output += fmt.Sprintf("%s|--G%d %s -> %s\n", strings.Repeat("\t",
				level), v.id, orderType[v.order], orderType[getConsensus(v)])
			visited[v.index] = true
		}
		// store keys in slice to sort
		var keys []int
		for k := range v.children {
			keys = append(keys, k)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(keys)))

		// traverse nodes of recursion tree
		for _, k := range keys {
			if !visited[v.children[k].index] {
				stk = append([]Node{v.children[k]}, stk...)
			}
		}
		// Calculate majority of OM tree
	}
	return output
}


// ---------------------------------------------------
func main() {

	var args = os.Args

	m, _ := strconv.Atoi(args[1]) // Level of recursion
	Oc := args[2] // Commander's order to lieutenants
	G := args[3:] // Generals array

	n := len(G) // Number of generals
	maxTraitors := (n - 1)/3
	counter := counter(G) // Count loyal/traitor

	// Error handling
	if n < 4 {
		fmt.Printf("ERROR: Number of generals (input: %d) must be greater than 3.\n", n)
		os.Exit(1)
	}
	if counter[loyaltyType[TRAITOR]] > maxTraitors {
		fmt.Printf("ERROR: Number of traitors (%d) exceeds maximum of %d.\n",
			counter[loyaltyType[TRAITOR]], maxTraitors)
		os.Exit(1)
	}
	if orderKey[Oc] > 2 || orderKey[Oc] < 0 {
		fmt.Println("Input is not valid.\n Usage: <level of recursion>, <command>, <generals list>")
	}
	fmt.Printf("\n * * * * * * * * * * * * * * * * * * * * * * * * * * *\n")
	fmt.Printf(" Level of Recursion: %d\n Order is to %s\n", m, Oc)
	fmt.Printf(" Commander: %s\n Total Lieutenants: %d (Loyals: %d | Traitors: %d)\n",
		G[0], n - 1, counter[loyaltyType[LOYAL]], counter[loyaltyType[TRAITOR]])
	fmt.Printf(" * * * * * * * * * * * * * * * * * * * * * * * * * * *\n\n")

	// Waiting group for processes
	var wg sync.WaitGroup
	wg.Add(n)

	// Create channel array
	msg := *NewVectorChannel(n, m)

	// ---------------------------------------------------
	// Launch Commander
	go func(n int, order string, msg vectorChannel, wg *sync.WaitGroup) {
		defer wg.Done()

		// Commander sends Oc regardless of loyalty
		fmt.Printf("Commander sends out order to %s!\n", order)
		for i := 1; i < n; i++ {
			msg.chs[i] <- Message{0, 1, orderKey[order], []int{}}
		}
	}(n, Oc, msg, &wg)

	// ---------------------------------------------------
	// Launch Lieutenants
	for i := 1; i < n; i++ {

		go func(i int, l string, msg vectorChannel, wg *sync.WaitGroup) {
			defer wg.Done()

			// Local recursion tree
			var root Node
			nodeIndex := 0
			// Received message
			var omIn Message
			expected := 1 // Commander oral message
			scheduler := Scheduler{make(map[int][]Message)}

			// Iterate rounds over level of recursion m
			for rnd := 1; rnd <= m + 1; rnd++ {
				// Receive oral messages until expected number reached
				received := 0
				for received < expected {

					// Check scheduler for waiting messages
					if len(scheduler.queue[rnd]) > 0 {
						omIn = scheduler.queue[rnd][0]
						scheduler.queue[rnd] = scheduler.queue[rnd][1:]
					} else {
						omIn = <-msg.chs[i]
					}

					// Check if message from current round
					if omIn.rnd == rnd {
						received++

						// Update local recursion tree
						newNode := Node{
							nodeIndex, omIn.id, omIn.order,
							make([]int, len(omIn.senders)), make(map[int]Node), omIn.rnd,}
						if nodeIndex == 0 {
							root = newNode
						} else {
							// Traverse tree by sender indexes
							node := root
							for _, index := range omIn.senders[1:] {
								node = node.children[index]
							}
							node.children[omIn.id] = newNode
							// Copy senders list as indexes in recursion tree
							copy(node.children[omIn.id].senders, omIn.senders)
						}
						nodeIndex++

						// Exit at recursion level m
						if rnd <= m {

							// Set receptors of outgoing message
							newReceptors := *NewSet()
							newReceptors.init(1, n)
							// For each order, relay message to all generals except:
							// (i) itself;
							newReceptors.remove(i)
							// (ii) nodes visited previously by the corresponding value (removed senders);
							for _, e := range omIn.senders {
								newReceptors.remove(e)
							}
							// (iii) the direct sender of the received message.
							newReceptors.remove(omIn.id)

							for r := range newReceptors.m {
								orderFiltered := omIn.order
								// Traitor: complement order for even-numbered lieutenants
								if r%2 == 0 && loyaltyKey[l] == 0 {
									orderFiltered = 1 - omIn.order
								}
								// Create outgoing message
								omOut := Message{
									i, rnd + 1, orderFiltered,make([]int, len(omIn.senders)),
								}
								// Copy senders list and add the current sender
								copy(omOut.senders, omIn.senders)
								omOut.senders = append(omOut.senders, omIn.id)
								// Relay message
								msg.chs[r] <- omOut
							}
						}
					} else {
						// Schedule the message for later round
						scheduler.queue[omIn.rnd] = append(scheduler.queue[omIn.rnd], omIn)
					}
				}

				// Update total expected messages
				if rnd == 1 {
					expected = n - 2
				} else {
					expected = expected * (n - rnd - 1)
				}
			}
			//if i == 1 {
				fmt.Printf("\n\nRecursion Tree for General %d\n\n%v\n\n",
					i, consensus(root))
				fmt.Printf("G%d Consensus Agreement: %s\n", i, orderType[getConsensus(root)])
			//}
		}(i, G[i], msg, &wg)
	}
	wg.Wait()
}