CSC546 Concurrency (Fall 2018)
Assignment 1
Prof. Yvonne Coady
October 11, 2018


Specifications

Processor
MacBook Pro (15-inch, 2018) 2.2 GHz Intel Core i7

Operating System
Problems 1-5:

Linux precise64 3.2.0-23-generic #36-Ubuntu SMP Tue Apr 10 20:39:51 UTC 2012
x86_64 x86_64 x86_64 GNU/Linux

Problem 6:
Darwin Kernel Version 17.7.0: Fri Jul  6 19:54:51 PDT 2018; 
root:xnu-4570.71.3~2/RELEASE_X86_64 x86_64 

GNU Compiler
gcc version 4.6.3 (Ubuntu/Linaro 4.6.3-1ubuntu5)

Golang
go version go1.11 darwin/amd64 

Python
Python 3.6.5 64bits, Qt 5.9.4, PyQt5 5.9.2 on Darwin

Analysis Overview

Correctness
1. Safety: Code demonstrates the avoidance of deadlocks, livelocks and runtime errors;
2. Liveness: An action is eventually executed that follows fair choice and action priority (avoidance of starvation).
Fairness: Choice over a set of transitions is executed infinitely often, then every transition in the set will be executed infinitely often.

Comprehensibility
1. Code Readability: 
Lines of Code (LoC): Though not a well-liked metric, LoC gives a rough idea of the overall complexity of a system.
Numbers of synchronization primitives and channels offer indications of difficulty, and effort of a piece of code; flow complexity; lines of code. (NOTE: n is the number of goroutines/threads using a primitive concurrently)

2. Maintainability:
Code shows coupling/cohesion: smaller modular units and demonstrates ease of testing; 
Code is amenable to error detection

Performance
1. Running Time: Total time to complete benchmark tests.
2. Heap Allocation: Total heap allocation at runtime.
3. Waiting time: Performance issues in multithreaded programs such as competition for resources, synchronization and scheduling problems, increase the overall waiting time.ii

4. Profilers:
pprof [Golang] (https://golang.org/pkg/net/http/pprof/) to provide runtime profiling data (CPU usage, memory allocation, mutex contention)
mutrace [C] (http://0pointer.de/blog/projects/mutrace.html) to monitor runtime and mutex contention
valgrind [C] to monitor memory allocation.

PROBLEMS
Generalized Cigarette Smokers Problem (Exercise 4.5.4)

The original Smokers Problem was conceived by Suhas Patilii (1971) to illustrate a multi-process synchronization problem that could not be solved using semaphores (i.e. without the use of conditional statements). This is an impractical restriction, according to David Parnasiv. However, the problem highlights the value of the semaphore's “nondeterminism,” in that when there are multiple threads waiting on a semaphore, the “V” operation does not specify the process to be released.  This is useful for schedulers subject to uncertainty and change, or applications that may be removed from a scheduling queue, while maintaining smooth (deadlock-free) operation of resource production and consumption. Downey likens the problem to how operating systems allocate resources to applications: the agent representing an OS (scheduler) that allocates resources to running applications (smokers). Synchronization between the agent and applications (smokers), in this example, requires that waiting applications only proceed when the correct resources are available. At the same time, we wish to avoid waking applications that cannot proceed, demonstrating simultaneous wait.  The solution proposed by David Parnas: to create "helper" or "pusher" threads that wake waiting smokers, acts as a communication bridge between the agent and smoker. In the generalized version of the problem (implemented for the assignment), the agent does not wait for a response (uptake) when resources are made available. 
 
Implementation A (Golang)
Adapted from the solution in Downey, p.101. 
Uses six channels (two for each ingredient) to synchronize resources.
The agent only signals available resources along helper channels (one for each ingredient) that keep track track of available ingredients using a scoreboard.
Each smoker communicates with a designated helper through an independent channel.
The use of the select statement allows simultaneous selection of signals, which is not available in C.

Implementation B (C)
Similar adaptation from the solution in Downey, p.101, but using semaphores in a way analogous to channels. 
Each smoker communicates with the designated helper through an independent channel.
An additional “ready” semaphore is added to let the agent know that a resource was used.
Analysis: Generalized Smokers Problem
Correctness


FIFO Barbershop Problem (Exercise 5.3)

The Barbershop problem is a classic multi-process scheduling problem that emerges in operating systems. The problem specifically models difficulties in synchronizing actions (without deadlock) that have an unknown amount of time. For example, an arriving thread or process may be about to be queued by the scheduler while an OS is running an application. In the process of being queued, however, the running thread may end that results in an empty queue. The scheduler therefore finds no waiting threads and goes to sleep. The scheduler and the thread are both waiting, i.e. deadlock. A less likely example is when two threads are vying to be queued at the same time when there happens to be only space or “seat” in the waiting room. Furthermore, we may know, or be able to specify, the OS wakeup policy of waiting threads (e.g. some implementations of futex), and so we cannot assume the order that queued threads are released. Using a application FIFO queue, as illustrated in the following implementations, resolves this potential by ensuring the correct order of queued threads.

Implementation A (Golang)
A buffered channel (channel-in-channel) is used as a FIFO queue of bounded capacity (number of seats in the waiting room).
Channels are employed for the barber and customer to wait and signal the start and end of a haircut.

Implementation B (C)
Adapted from Downey, p.127.
Required a FIFO queue implementation locked by a mutex.
The customers enqueue themselves (unless the queue is full, and therefore exit); the barber pops customers off the queue.
Each customer has a semaphore that is passed to the queue and used by the barber to signal each customer separately.
Once the haircut is finished, the barber and customer interleave a receive-signal-confirmation pattern that ensures the threads proceed in lockstep (the customer exits, and the barber proceeds to the next customer in the queue).

Building H20 (Exercise 5.6)

The H2O problem focuses on the synchronization pattern of the barrier. Barriers block threads until a threshold number of threads are waiting, at which point all waiting threads are released. Barriers are useful when you want a number of tasks to be completed before an overall phase of work can proceed. For example, parallelized calculations require local sequential calculations to be computed before a parallel calculation can execute (see Naive Bayes Classifier below). Similarly, the separate phases of a multi-phase algorithm might require the coordination and barrier synchronization of independently running sub-tasks. A barrier can force all of the threads that are doing parallel computations to wait until all involved threads have reached the barrier. When the threads have reached the barrier, the threads are released and begin computing together.  In the following implementations, we can see that a simple wait call is insufficient to match threads in a structured triplet (H2O) that is released once all of the correct component threads are assembled.

Implementation A (Golang)
For each goroutine “bond”, a channel is sent within a channel to facilitate one-to-one communication between the three entities.
A simplified barrier that uses channels is employed to force the aggregating goroutines to wait for the correct adjoining goroutines.
The "oxygen" goroutine coordinates the bond of two "hydrogen".
Barrier is released once the bonding channel is closed by the last goroutine to join.
A scoreboard locked by a mutex tracks the number of atoms and formed H2O molecules.

Implementation B (C)
Adapted from Downey, p.148.
Uses a pthread barrier to synchronize the assembly of triplet threads.
A scoreboard locked by a mutex tracks the number of atoms and formed H2O molecules.


Search-Insert-Delete Problem (Exercise 6.1)

The Search-Insert-Delete problem is a variation of the readers-writers problem using multiple categorical mutual exclusion: mutexes determine which category of thread allowed in the critical section at any time. As Downey states, “A thread in the critical section does not necessarily exclude other threads, but the presence of one category in the critical section excludes other categories.”  Hence, the problem demonstrates different allowable types of concurrent access by the different types of computational entities. The problem is illustrative of any situation in which multiple processes try to access shared data, such as a data structure or database, where we want to ensure the reading of the data is consistent and valid. The Search-Insert-Delete problem includes an additional category: the Inserter, which runs concurrently with searchers (essentially readers), but excludes deleters (writers) and other inserters. As categorical exclusion problems are prone to asymmetric solutions, in which one category of thread blocks others from progressing, starvation is probable. Specifically, deleters must wait for searchers and inserters - which can run concurrently - to complete before locking the linked list. This problem can be eased by prioritizing one category (e.g. writers), or, as presented in Implementation B, by combining concurrent design patterns such as multiplexes, which limit the number of concurrent processes in the critical area, and turnstiles, which allow one thread in at a time. It's not clear whether an optimized solution to fairness is possible.

Implementation A (Golang)
Uses Go's RWMutex, which ensures a blocked Lock call excludes new readers from acquiring the lock.
If a goroutine holds a RWMutex for reading and another goroutine might call Lock, no goroutine should expect to be able to acquire a read lock until the initial read lock is released to ensure that the lock eventually becomes available. This has some effect on avoiding the starvation of deleters: if a searcher is in the critical section, it forces deleters to queue, but not other searchers. So, if a deleter arrives it can lock searchers, which will cause subsequent searchers to queue.

Implementation B (C)
Uses intersecting concurrency patterns of Lightswitch, a synchronization technique that employs a First-In-Last-Out principle for threads onto a semaphore, so that group of threads can collectively request access to the critical section, and release it when the group has finished processing.
Required a linkedlist implementation.


Baboons Crossing (Exercise 6.3)

Given two types of threads A and B, this problem models how a single resource can only be shared concurrently (i.e. accessed as read-only) by a fixed number of threads, and furthermore, one type at a time. Threads of one type, say type B, are locked out from accessing the resource until no threads of type A are in the critical section. This problem might model a two-way serial communication channel that is exchanging data between servers. As with the Search-Insert-Delete problem, starvation of one type of thread or process is a significant possibility, if, for example, a stream of type A threads locks the resource indefinitely. This is equivalent to only one side of a communication channel being able to efficiently transmit data. To mitigate the problem, we want to ensure a fair interleaving of both types of threads (baboons going East and West get equal rope time). In Implementation B, a synchronization scheme that combines design patterns turnstile, lightswitch and multiplex can facilitate fairness by ensuring one type does not dominate access to the resource.

Implementation A (Golang)

Mutex-free implementation that uses Go's select functionality to wait on multiple channels simultaneously, which I'm not aware of an analogue to that in C.
A “rope” goroutine is both the shared resource and acts as a monitor by tracking waitlists on both sides and the number of crossing baboons at one time, while granting the fixed number of concurrent baboons.
The baboon goroutine is quite minimal – it merely requests access to the rope, waits for a response, and confirms that it has crossed, all on one channel. The tradeoff is that the rope goroutine is long and complicated channel selection. (Note, however (as outlined in the Golang documentation3), if multiple requests are waiting, one of them is chosen by the select statement at random (pseudo-random). 

Implementation B (C)

Uses intersecting concurrency patterns: 
Lightswitch: Employs a First-In-Last-Out principle for threads onto a semaphore, so that group of threads can collectively request access to the critical section, and release it when the group has finished processing.
Turnstile: Semaphore wait-signal that allows only one thread to proceed at a time.
Multiplex: A pre-loaded semaphore that allows X number of threads in the critical section at one time. 
Inverse of Implementation A, since the baboon threads do all the work, and the rope thread is only used to (optionally) display scoreboard information at specific intervals.
Analysis: Baboons Crossing
Correctness


Naive Bayes Classifier
This problem might fall under the model of Single Program Multiple Data (SPMD): the parallelization of an algorithm by dividing up tasks and running them simultaneously on multiple processors. The goal is to achieve faster results for the computation than a serial implementation of the algorithm. In the following solution, a bott


Implementation A
TBA

Implementation B
TBA



