// =========================================
// Example 4: Baboon Crossing Problem (Exercise 6.3)
// =========================================
// Completed for CSC 564 (Concurrency), Prof. Yvonne Coady, Fall 2018
// Spencer Rose (ID V00124060)

package main

import (
	"fmt"
	"sync"
	"math/rand"
)

func cross(i int, dir int) {
	fmt.Printf("baboon %d crossing %s!", i, dir)
	return
}

func main() {
	var n = 12
	var crossing = 0
	dir[0] = "west"
	dir[1] = "east"

	var mutex = &sync.Mutex{}

	// Waiting group for baboon threads
	var wg sync.WaitGroup

	// channel represents current crossing direction
	dirCh := make(chan int, 5)

	// Launch n baboon threads
	for i := 1; i <= n; i++ {
		wg.Add(1)
		// direction east = 0; west = 1
		d := rand.Intn(1)
		fmt.Printf("baboon %d goes in direction: %d\n", i, dir[d])

			go func() {
				// get ready to cross
				<- start
				// check direction of crossing
				for {
					if (d == 0) {
						<- eastDir
					}
					else {
						<- westDir
					}
					// check number of baboons crossing
					mutex.Lock()
					if crossing < 5 {
						crossing++
						// signal direction crossing
						if (d == 0) {
							eastDir <- i
						}
						else {
							westDir <- i
						}
					}
					mutex.Unlock()

				}
			}()

				// Launch "rope" thread



				select {
				case cross := <- dirCh:
					<-done
					fmt.Printf("Haircut done. Customer #%d leaves\n", c)
				default:
					fmt.Printf("No seats! Customer #%d leaves\n", c)
				}



  }
	close(start)


	wg.Wait()
	close(direction)
}
