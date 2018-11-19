/*
=========================================
Vector Clocks (Assignment 2.1)
=========================================
Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
Author: Spencer Rose
=========================================
SUMMARY: Demonstrates the operation of vector clocks in n processes to identify events that might be causally concurrent.
 - Initially all clocks are zero.
 - Each event internal to a process increments that process' own logical clock in the vector by one.
 - Each time a process sends a message, it increments its own logical clock in the vector by one and then sends
   a copy of its own vector.
 - Each time a process receives a message it updates each clock value in its vector by choosing the maximum value of its
   own vector clock and the value in the received vector (for every element).

Reference: Fidge, Colin J., Timestamps in Message-Passing Systems That Preserve the Partial Ordering
*/

package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const clkLimit  = 5
const n  = 3

type message struct {
	id int
	vc [n]int
}

type vectorChannel struct {
	chs []chan message
}

// Vector channel constructor (one ch per goroutine)
func NewVectorChannel(n int) *vectorChannel {
	vch := vectorChannel{}
	for i := 0; i < n; i++ {
		vch.chs = append(vch.chs, make(chan message, n))
	}
	return &vch
}

// Returns maximum of two integers
func max(a, b int) int {
	if a > b {return a}
	return b
}

func makeSelection(i int, n int) map[int]int {
	var m = make(map[int]int)
	t := 0
	for k := 0; k < n; k++ {
		if k != i {
			m[t] = k
			t++
		}
	}
	return m
}

// Event tracker
func tracker(vch vectorChannel, done chan bool) {
	// Create history object
	var history = make(map[int][][n]int)
	var colHeader string
	//var rowHeader string
	var rows string
	var results string
	for {
		in, more := <-vch.chs[n]
		if more {
			history[in.id] = append(history[in.id], in.vc)
		} else { // Channel closed: Print history
			i := 0
			p := makeSelection(i, n)
			for i < n && len(p) > 0 {
				rows = "\n"
				for j := range p {
					colHeader += fmt.Sprintf( "%-10s", fmt.Sprintf("P%d/P%d",i, p[j]))
					for q, vj := range history[p[j]] {
						rows  += fmt.Sprintf("%-10s", fmt.Sprintf("%v",vj))
						for _, vi := range history[i] {

							if q == 0 {
								colHeader += fmt.Sprintf("%-10s", fmt.Sprintf("%v",vi))
							}

							concurrency := strconv.FormatBool(isConcurrent(vi, vj))
							rows += fmt.Sprintf("%-10s", concurrency)
						}
						rows += "\n"
					}
					results += colHeader + rows
					fmt.Println(results)
					colHeader = "\n"
					results = "\n"
					rows = "\n"
				}

				delete(p,i)
				i++
			}

			done <- true
			return
		}
	}
}


// Process with vector clock
func process(i int, vch vectorChannel, wg *sync.WaitGroup) {
	defer wg.Done()
	// Create local vector clock
	var vc [n]int
	var vcPrev [n]int

	// Create other goroutine selection array
	p := makeSelection(i, n)

	// Run updates until clock value reaches defined limit
	for vc[i] < clkLimit {

		// Update tracker history
		vcPrev = vc
		vc[i]++

		rand.Seed(time.Now().UnixNano())
		var j = p[rand.Intn(len(p))]
		select {
		// Check for incoming messages
		case in := <-vch.chs[i]:

			// Compare values of received vector with local vector
			for k := range vc {
				vc[k] = max(vc[k], in.vc[k])
			}
			fmt.Printf("P%d receives %v from P%d: Update %v -> %v\n", i, in.vc, in.id, vcPrev, vc)
			vch.chs[n] <- message{i, vc}
		// Send message to randomly selected process
		case vch.chs[j] <- message{i, vc}:
			vch.chs[n] <- message{i, vc}
			//fmt.Printf("P%d sends %v to P%d\n", i, vc, j)
		// No message handling: go to sleep for 2 seconds
		default:
			time.Sleep(2*time.Second)
		}
	}
}

/* 	Function determines whether two events are concurrent:
	If every v1[i] <= v2[i] then v1 -> v2 (not concurrent)
	If every v1[i] >= v2[i] then v2 -> v1 (not concurrent)
	Otherwise, if neither conditions apply for at least some ith timestamp: concurrent
*/
func isConcurrent(v1 [n]int, v2 [n]int) bool {
	v1v2 := false
	v2v1 := false

	for i := range v1 {
		if v1[i] < v2[i] {v1v2 = true}
		if v1[i] > v2[i] {v2v1 = true}
	}
	if v2v1 && v1v2 {return true} /* the vectors are concurrent */
	return false /* the vectors are not concurrent */
}


func main() {

	fmt.Printf("Generating %d goroutines... \n\n", n)

	// Waiting group for processes
	var wg sync.WaitGroup
	done := make(chan bool)

	// Create channel array
	vCh := *NewVectorChannel(n + 1)

	go tracker(vCh, done)

	// Launch n goroutines
	for i := 0; i < n; i++ {
		wg.Add(1)
		go process(i, vCh, &wg)
	}

	wg.Wait()

	// Close tracker channel to get results
	fmt.Println("\nConcurrent Events")
	close(vCh.chs[n])
	<-done
}