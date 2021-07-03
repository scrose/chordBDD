# Binary Decision Diagram Package using a Chord-Based Distributed Hash Table

## Abstract
Binary Decision Diagrams (BDDs) are the method of choice for symbolic verification and manipulation of Boolean functions. For large and complex functions, distributed (and parallelized) BDDs provide a compact, scalable and canonical data structure supported by a distributed hash-table (DHT) for node lookups and insertions. This paper describes a BDD package implementation that uses the Chord distributed lookup protocol for partitioned BDD node data across a cluster of multi-core machines. This package offers a scalable, load-balanced memory storage and O(logN) lookup time.

## Keywords

Symbolic model checking, distributed hash tables, binary decision diagrams, Chord protocol

## Overview

Binary-Decision Diagrams (originally Binary-Decision Programs) were formally introduced in 1959 by C. Y. Lee at Bell Lab [4] as a successful method for analyzing the logical operation of switching circuits. Later research established the BDD as a canonical way to verify arbitrary Boolean expressions, with numerous application areas. BDDs are used extensively in CAD software for the design of digital systems; and more generally for formal verification, combinatorial problems, and context-sensitive program analysis. The seminal work of Bryant et al. [2] helped to translate decades of conceptual work on BDDs into practice through the useful property of the canonicity of Reduced Ordered BDDs (ROBDDs). In his highly-cited 1990 paper, he describes the efficient ITE algorithm used to synthesize and manipulate BDDs for basic Boolean operations.

Since the 1990s, BDDs have been thoroughly explored, and numerous open-source implementations – both parallel and sequential – are currently available. This package adapts design decisions from a selection of these: (1) a parallel BDD package developed by Stornetta and Brewer [9], for non-shared distributed memory multi-processing environments; (2) Milvang-Jensen's [5] BDDNOW applied to a workstation network using a breadth-first parallel implementation; (3) Oortwijn's recent work on algorithms that exploit Remote Direct Memory Access (RDMA), a high-throughput, low- latency networking that bypasses the cluster node's operating system entirely [7].

An important advantage of the BDD is its compactness in representing complex binary and unary operations. This is because BDDs do not explicitly enumerate all paths, where each path in the BDD represents a particular variable assignment. Instead, a BDD represents paths by a graph whose size is measured by the number of the nodes.1 BDD computations typically apply operations to combine multiple BDDs. However, this often results in intermediate BDD computations that can be impractically large even if the final BDD is small.

This implementation addresses the challenge of how to distribute very large BDD data across a network cluster with minimal access times. BDD manipulations are hard to parallelize because they are memory intensive and their memory access patterns are irregular [9]. Initially, this project was to explore parallelized BDD operations using Apache Spark, which came to mixed results, and did not allow an exploration of the underlying distributed hash table. This paper instead focuses on a data partition strategy of node data across a single distributed hash table. There is an implicit understanding that, in a more complete analysis, the two aspects – distributed hashing and explicitly parallelized operations – should be considered interdependent.

In my implementation, I have adapted the Chord-based distributed hash table (DHT) initially developed by Stoica et al., for BDD applications [8]. The Chord protocol offers a number of features favourable to BDD processing: (1) Flat key-space for flexible hash naming; (2) Scalability: To handle large BDD datasets; (3) Availability: Node failure resilience through stabilization and replication; (4) O(logN) lookup time for N Chord nodes; (5) Effective load balancing through random hashing. These features support the distributed storage of BDD data and memoization tables needed to hold the intermediate computations of BDD operations. I also address some of the significant drawbacks encountered in this experiment, and in particular those raised by Milvang-Jensen [5] and Oortwijn [7], on high
1 Knuth offers an excellent breakdown of the many other virtues of BDDs [3].

## Binary-Decision Diagrams (BDDs)
BDDs are compact graph representations of Boolean expressions that decompose formulae into their component Shannon cofactors.

## Chord Protocol

To facilitate BDD memoization and unique tables, this implementation uses a Chord-based distributed hash table, which forms a secondary focus of my project. My implementation is closer to Stornetta and Brewer's design in that the Chord DHT assumes a randomly distributed key-space, however the Chord protocol offers significant differences in handling BDD node lookups and insertions.

Chord is a protocol and algorithm for a peer-to-peer distributed hash table, and hence decentralized in its storage and retrieval, rather than (for example) a client/server system that requires a central server. It uses an m-bit hash key to identify Chord nodes in a network cluster. Each node3 points to another node called its successor, and the last node points to the first node to form a “network overlay”.

DHT's store data by mapping a key to each data item (or partition) and assigning the resultant key/data pair to a node to which the key maps. Each node stores the data key identifers in the system (node) closest in value to the node’s identifer, also called the successor node. The Chord uses consistent hashing on the node (or peer's) IP address as follows: (1) the IP address is inputed to a SHA-1 hash function to generate a unique 160-bit string; (2) The SHA-1 string is truncated to m-bits resulting in a node id between 0 and 2m- 1, which is mapped to a discrete point on a logical m-bit circle or ring. DHTs such as Chord are designed to distribute keyspace fairly across participating nodes in the cluster. Unbalanced distributions, where some nodes perform an disproportionate amount of work or data storage, are avoided by a random allocation of node keys on the ring. In fact, however, the probability density function of keys to nodes is exponential [8], and Stoica offers some improvements using virtual nodes to distribute more evenly.

To improve the efficiency (and decrease latency) of lookup of successors, each node has a local lookup table called a finger table. The finger table holds up to m successor nodes for “finger” or chord values that successively “hops” a pre-calculated distance away from the local node (i.e. forming a ring chord – see the diagram below). Each finger entry maps to the next successor of that key value. This improves key lookup time over a linear search of the successors from O(m) to O(log m). Given k keys and n nodes, each node (peer) can store ~ O(k/n) data keys.

## Implementation 

In this package, both the BDD and Chord data structures are implemented in Scala (JVM) with Akka libraries. Scala is a general-purpose programming language, though well-suited for concurrent applications. Akka is a JVM toolkit used to construct concurrent and distributed applications.

### BDD Operations

The most fundamental BDD operations are the ITE (If-Then-Else) operator and the BDD constructor build(). Package testing therefore consisted of constructing BDDs and applying a given operator (AND or OR) using ite(). The indexes of the unique node table are hashed on the Chord DHT, along with the (v,lo,hi) triples (inverse lookup), and the intermediate computation table used in the ITE algorithm (i.e. all three tables could occupy the same Chord DHT).

A sample output file is provided in the the code repository (results/project-1.output  and results/project-2.output) showing a printout of the Chord initialization, stabilization of the finger tables, and example BDD builds for x 0 <=> x 2 formula and x 1 <=> x 3 formulaas well as the resultant ITE AND operation. Tables 2-1, 2-2, 2-3 and Figure 2-1 summarize the example and printed results. Four nodes are created for a 12-bit Chord ring and 10 Chord nodes. As each node joins the chord by communicating with the gateway node, the values of the finger tables are stabilized and propagated along the ring, shown by the finger table updates.
 
### Chord Operations

The Chord implementation presented here is adapted from the protocol and algorithm proposed by Stoica [3]. A network was simulated using Akka's peer-to-peer based cluster service. Nodes in a network cluster were assigned top-level Akka Actors to act as Chord nodes in the DHT. Akka Actors are analogous to objects in object-oriented design, but for concurrent systems. They serve as containers for behaviour and state. Built in Scala, the Akka's actor system allows for interaction and information sharing between independent actors.

In this implementation, the Chord node actors are identified on an 12-bit/10 node Chord ring created as follows:

- An initial node (gateway node) instantiates the ring and initializes its finger table.  
- Node identifiers are created through a truncated SHA-1 hash key of the local system IP address.  
- A list of remote nodes assigned port numbers join the ring by looking up a generated identifier key on a gateway node. 
- Once joined, the node updates its finger table and successor/predecessor keys by looking up successor keys. 
- To maintain key correctness, each node periodically runs stabilization tests on successor nodes. 
- When a node leaves or aborts, the key-values are transfer to its successor. 
- Node failure is handles by a “Successor Overlap” strategy (see Node Stabilization and Failure below) 

DHT lookups are essentially successor finds (findSuccessor(key)) followed by a return of the data at that key. Insertions are preceded by a look up of the successor node for the hash key of the data key (an arbitrary string). The value is then added to the data bank of the successor and to the replicated data bank the successor's successor – the latter action to handle node failures (see Node Failure below).

For join and insert operations, the implementation does not handle hash key collisions. Given the small number of nodes and data keys used in this test project (much less than the total 4096 keys of the 12-bit ring), it is assumed the risk of collisions can be ignored [10].
 
#### ROBDD API and Node Operations

| Operation | Description |
| --- | --- |
| bdd.build(exp \<string\>, vars \<Array\>) | Constructs an ROBDD object in if-then-else normal form from a Boolean expression and ordered list of variable identifiers. The expression is parsed and evaluated using JEval Java library. |
| bdd.isMember(key <int>) | Returns closest node preceding the key value among all the fingers of node. |
| bdd.insert() | Verifies local node's successor. Runs periodically. |
| bdd.mkNode(n' \<node>\) | Inserts a BDD node (v,lo,hi) triplet to the unique table. |
| bdd.lookup(index \<int>\, low <int>, high \<int>\) | Look up unique node using the variable, low and high indexes. |
| bdd.ite(bdd1 \<Node>\, bdd2 \<Node>\, op \<enum>\) | Computes logical operation op on two ROBDDs. |
| bdd.printTable() | Prints a tabular representation of the ROBDD. |
  

### Chord Distributed Hash Table API and Node Operations

| Operation | Description |
| --- | --- |
| dht.lookup(key \<string\>) | Returns value in dht (distributed hash table object) stored at location key. |
| dht.insert(key \<string>, data <list>) | Stores data in dht and stores it at location key. |
| n.join(n' \<node>\) | Joins node n to the Chord ring using connected node n' as a gateway node. |
| n.findSuccessor(key \<int>\) | Finds successor node responsible for data at key. Uses n as a gateway node.
| n.findPrecedingNode(key \<int>\) | Returns closest node preceding the key value among all the fingers of node.
| n.stabilize() | Verifies local node's successor. Runs periodically.
| n.notify(n' \<node>\) | Notifies n'  that n might be its predecessor. Runs periodically as part of stabilize().
| n.checkPredecessor() | Checks status of predecessor and reacts to a node failure.
| n.updateFingers() | Verifies the successor node stored in each entry in n 's finger table. Runs periodically
| n.moveKeys(n \<node>\) | Following node failure, method relocates replicated key-values to main data bank. 
  
  
## References

1.  Andersen, H. R. (1997). An introduction to binary decision diagrams. Lecture notes, available online, IT University of Copenhagen, 5. 
2.  Brace, K. S., Rudell, R. L., & Bryant, R. E. (1990, June). Efficient implementation of a BDD package. In Design Automation Conference, 1990. Proceedings., 27th ACM/IEEE (pp. 40-45). IEEE. 
3.  Knuth, Donald E. "The Art of Computer Programming: Bitwise Tricks & Techniques." Binary Decision Diagrams 4 (2009). 
4.  Lee, C. Y. (1959). Representation of Switching Circuits by Binary‐Decision Programs. Bell system Technical journal, 38(4), 985-999. 
5.  Milvang-Jensen, K., & Hu, A. J. (1998, November). BDDNOW: a parallel BDD package. In International Conference on Formal Methods in Computer-Aided Design (pp. 501-507). Springer, Berlin, Heidelberg. 
6.  Narayan, A., Jain, J., Fujita, M., & Sangiovanni-Vincentelli, A. (1997, January). Partitioned ROBDDs—a compact, canonical and efficiently manipulable representation for Boolean functions. In Proceedings of the 1996 IEEE/ACM international conference on Computer-aided design (pp. 547-554). IEEE Computer Society. 
7.  Oortwijn, W., Dijk, T. V., & Pol, J. V. D. (2017, July). Distributed binary decision diagrams for symbolic reachability. In Proceedings of the 24th ACM SIGSOFT International SPIN Symposium on Model Checking of Software (pp. 21-30). ACM. 
8.  Stoica, I., Morris, R., Karger, D., Kaashoek, M. F., & Balakrishnan, H. (2001). Chord: A scalable peer-to-peer lookup service for internet applications. ACM SIGCOMM Computer Communication Review, 31(4), 149-160. 
9.  Stornetta, T., & Brewer, F. (1996, June). Implementation of an efficient parallel BDD package. In Design Automation Conference Proceedings 1996, 33rd (pp. 641-644). IEEE. 
10.  Zave, P. (2015). How to make Chord correct. arXiv preprint arXiv:1502.06461. 
