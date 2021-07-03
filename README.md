# Binary Decision Diagram Package using a Chord-Based Distributed Hash Table

## Abstract
Binary Decision Diagrams (BDDs) are the method of choice for symbolic verification and manipulation of Boolean functions. For large and complex functions, distributed (and parallelized) BDDs provide a compact, scalable and canonical data structure supported by a distributed hash-table (DHT) for node lookups and insertions. This paper describes a BDD package implementation that uses the Chord distributed lookup protocol for partitioned BDD node data across a cluster of multi-core machines. This package offers a scalable, load-balanced memory storage and O(logN) lookup time.

## Keywords

Symbolic model checking, distributed hash tables, binary decision diagrams, Chord protocol

## Overview

Binary-Decision Diagrams (originally Binary-Decision Programs) were formally introduced in 1959 by C. Y. Lee at Bell Lab [4] as a successful method for analyzing the logical operation of switching circuits. Later research established the BDD as a canonical way to verify arbitrary Boolean expressions, with numerous application areas. BDDs are used extensively in CAD software for the design of digital systems; and more generally for formal verification, combinatorial problems, and context-sensitive program analysis. The seminal work of Bryant et al. [2] helped to translate decades of conceptual work on BDDs into practice through the useful property of the canonicity of Reduced Ordered BDDs (ROBDDs). In his highly-cited 1990 paper, he describes the efficient ITE algorithm used to synthesize and manipulate BDDs for basic Boolean operations.

Since the 1990s, BDDs have been thoroughly explored, and numerous open-source implementations – both parallel and sequential – are currently available. In my discussion, I consider design decisions from a selection of these: (1) a parallel BDD package developed by Stornetta and Brewer [9], for non-shared distributed memory multi-processing environments; (2) Milvang-Jensen's [5] BDDNOW applied to a workstation network using a breadth-first parallel implementation; (3) Oortwijn's recent work on algorithms that exploit Remote Direct Memory Access (RDMA), a high-throughput, low- latency networking that bypasses the cluster node's operating system entirely [7].

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

Both the BDD and Chord data structures are implemented in Scala (JVM) with Akka libraries. Scala is a general-purpose programming language, though well-suited for concurrent applications. Akka is a JVM toolkit used to construct concurrent and distributed applications.

### Table 2-5. Chord Distributed Hash Table API and Node Operations

| Operation | Description |
| --- | --- |
| dht.lookup(key <string>) | Returns value in dht (distributed hash table object) stored at location key. |
| dht.insert(key <string>, data <list>) | Stores data in dht and stores it at location key. |
| n.join(n' <node>) | Joins node n to the Chord ring using connected node n' as a gateway node. |
| n.findSuccessor(key <int>) | Finds successor node responsible for data at key. Uses n as a gateway node.
| n.findPrecedingNode(key <int>) | Returns closest node preceding the key value among all the fingers of node.
| n.stabilize() | Verifies local node's successor. Runs periodically.
| n.notify(n' <node>) | Notifies n'  that n might be its predecessor. Runs periodically as part of stabilize().
| n.checkPredecessor() | Checks status of predecessor and reacts to a node failure.
| n.updateFingers() | Verifies the successor node stored in each entry in n 's finger table. Runs periodically
| n.moveKeys(n <node>) | Following node failure, method relocates replicated key-values to main data bank. 
