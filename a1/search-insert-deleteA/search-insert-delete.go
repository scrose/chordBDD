/*
=========================================
Example 5: The Search-Insert-Delete Problem (Exercise 6.1, pp. 165-169)
=========================================
Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
Spencer Rose (ID V00124060)

SUMMARY: Three kinds of threads share access to a singly-linked list: searchers,
inserters and deleters. Searchers merely examine the list; hence they can execute
concurrently with each other. Inserters add new items to the end of the list;
insertions must be mutually exclusive to preclude two inserters from inserting
new items at about the same time. However, one insert can proceed in parallel
with any number of searches. Finally, deleters remove items from anywhere in
the list. At most one deleter process can access the list at a time, and deletion
must also be mutually exclusive with searches and insertions.

References:
https://medium.com/golangspec/sync-rwmutex-ca6c6c3208a0
https://bbengfort.github.io/snippets/2017/09/08/lock-queueing.html
*/
package main

import (
	"container/list"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type turnstile struct {
	sync.Mutex
	d int
	s int
	ch chan bool
}

type linkedlist struct {
	searchMux sync.RWMutex
	insertMux sync.Mutex
	items *list.List
}

// Searcher: Multiple searchers can examine the list concurrently with other searchers and inserters;
// mutually exclusive of deletions.
func searcher(c chan int, l *linkedlist, t *turnstile, wg *sync.WaitGroup) {
	defer wg.Done()
	defer l.searchMux.RUnlock()

	rand.Seed(time.Now().Unix())
	var val = rand.Intn(10)

	// Check turnstile for writers enqueued
	if t.d > 1000 {
		t.Lock()
		t.s++
		t.Unlock()
		<-t.ch
		}

	l.searchMux.RLock()
	//fmt.Printf("Search RLock acquired: %d\n", l.items.Len())
	c <- 1
	// Search for element val in list
	for e := l.items.Front(); e != nil; e = e.Next() {
		if e.Value == val {
			c <- -1
			return
		}
	}
	c <- -1
	return
}

// Inserters add new items to the end of the list; insertions must be mutually
// exclusive to preclude two inserters from inserting new items at about the
// same time. However, one insert can proceed in parallel with any number of
// searches.
func inserter(c chan int, l *linkedlist, wg *sync.WaitGroup)  {
	defer wg.Done()

	rand.Seed(time.Now().Unix())
	var val = rand.Intn(10)

	//fmt.Printf("Insert enqueue: %d\n", l.items.Len())
	l.searchMux.RLock()
	l.insertMux.Lock()
	c <- 1
	l.items.PushBack(val)
	c <- -1
	l.insertMux.Unlock()
	l.searchMux.RUnlock()
}

// Deleter removes item randomly from the list. At most one deleter process
// can access the list at a time, and deletion must also be mutually exclusive
// with searches and insertions.
func deleter(c chan int, l *linkedlist, t *turnstile, wg *sync.WaitGroup)  {
	defer wg.Done()
	defer l.searchMux.Unlock()

	// Update turnstile
	t.Lock()
	t.d++
	t.Unlock()

	l.searchMux.Lock()
	c <- 1

	// Check list is not empty
	if l.items.Front() == nil {
		fmt.Printf("DELETE FAILED: List is empty!\n")

	} else {
		// Generate random index value
		var i = rand.Intn(l.items.Len())
		var j = 0

		// Find and delete random list element at index i
		e := l.items.Front()
		for ; e != nil && j < i; e, j = e.Next(), j + 1 {}
		if e != nil && j == i {
			l.items.Remove(e)
		} else {
			fmt.Printf("DELETE FAILED: Value at index %d [%d | %d] not found!\n", i, j, l.items.Len())
		}
	}
	// Update turnstile
	c <- -1
	t.Lock()
	t.d--
	for ; t.s != 0;  t.s-- {t.ch <- true}
	t.Unlock()
}


func main() {
	// Number of goroutine triplets
	var n = 300

	// Initialize empty singly-linked list
	l := linkedlist{
		items: list.New(),
	}
	t := turnstile{
		ch: make(chan bool),
		d: 0,
		s: 0,
	}

	// Waiting group for n goroutines
	var wg sync.WaitGroup

	// Tracker thread
	var s, i, d int
	sCh := make(chan int)
	iCh := make(chan int)
	dCh := make(chan int)
	go func() {
		for {
			select {
			case n := <-sCh:
				s += n
			case n := <-iCh:
				i += n
			case n := <-dCh:
				d += n
			}
			fmt.Printf("%s%s%s\n",
				strings.Repeat("S", s),
				strings.Repeat("I", i),
				strings.Repeat("D", d))
		}
	}()

	for k := 1; k < n; k++ {
		wg.Add(3)
		// Generete searcher, inserter and deleter threads
		go searcher(sCh, &l, &t, &wg)
		go inserter(iCh, &l, &wg)
		go deleter(dCh, &l, &t, &wg)
	}

	wg.Wait()


	}
