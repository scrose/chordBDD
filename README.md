# Binary Decision Diagram Package using a Chord-Based Distributed Hash Table

## Abstract
Binary Decision Diagrams (BDDs) are the method of choice for symbolic verification and manipulation of Boolean functions. For large and complex functions, distributed (and parallelized) BDDs provide a compact, scalable and canonical data structure supported by a distributed hash-table (DHT) for node lookups and insertions. This paper describes a BDD package implementation that uses the Chord distributed lookup protocol for partitioned BDD node data across a cluster of multi-core machines. This package offers a scalable, load-balanced memory storage and O(logN) lookup time.

## Keywords
Symbolic model checking, distributed hash tables, binary decision diagrams, Chord protocol

