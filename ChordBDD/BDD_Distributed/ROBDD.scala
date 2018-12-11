// -----------------------------------------------------------------
// Reduced Ordered Binary Decision Diagram (ROBDD) - Distributed Package
//  * Object for representing and manipulating BDDs
// Parameters:
//  * expr <string> = Boolean expression in standard Java eval notation
//  * vars <array[<string>]> = List of ordered variable string identifiers
//  * DHT <Chord> = Chord distributed hash table singleton
// -----------------------------------------------------------------

package BDD_Distributed

import BDD_Structure._
import Chord_DHT._
import net.sourceforge.jeval.Evaluator
import org.apache.spark.SparkContext

import scala.collection.mutable

class ROBDD(expr: String, vars: Array[String], DHT: Chord) {
  val n: Int = vars.length
  var nodeCount: Int = 0
  var exp: String = expr.replaceAll("""[\n\r\s]+""", "")
  val bdd_index: Int = expr.hashCode()

  //  Initialize Distributed Hash Table (DHT) represents nodes in ROBDD
  initDHT()
  private def initDHT() {
    // Create terminal nodes [0,1]
    DHT.insert((bdd_index, 0), new BDDNode(n, 0, 0))
    DHT.insert((bdd_index, 1), new BDDNode(n, 0, 0))
    nodeCount += 2
  }

  // -----------------------------------------------------------------
  // Return root node index of ROBDD
  private def root: Int = {
    nodeCount - 1
  }

  // -----------------------------------------------------------------
  // Initialize H (Inverse of T): maps (variable,low,high) triples to node index in T

  // Add node to lookup table H
  private def insert(node: Int, i: Int, l: Int, h: Int): Boolean = {
    DHT.insert((bdd_index, i, l, h), node)
    true
  }
  // Test if node is hashed
  private def isMember(i: Int, l: Int, h: Int): Boolean = {
    DHT.lookup((bdd_index, i, l, h)).isDefined
  }
  // Look up BDD node in unique table
  private def lookup(i: Int, l: Int, h: Int): Int =  {
    DHT.lookup((bdd_index, i, l, h)).getOrElse(-1).asInstanceOf[Int]
  }

  // -----------------------------------------------------------------
  // Create & add BDD node to ROBDD
  private def mkNode(index: Int, low: Int, high: Int): Int = {
    if (low == high) low
    else if (isMember(index, low, high)) lookup(index, low, high)
    else {
      val node: Int = add(index, low, high)
      if (insert(node, index, low, high)) node else -1
    }
  }

  // Add unique node to computed table
  private def add(i: Int, l: Int, h: Int): Int = {
    val index = nodeCount
    val node = new BDDNode(i, l, h)
    DHT.insert((bdd_index, index), node)
    nodeCount += 1
    index
  }

  // -----------------------------------------------------------------
  // Construct ROBDD for boolean expression exp <String>
  def build(): Int = {
    printf("\n\nDistributed build of ROBDD for Exp = \n%s\n\n", exp)
    def bld(exp: String, i: Int): Int = {
      if (i > n) {
        if (eval(exp)) 1 else 0
      }
      // Shannon co-factors
      else {
        val l = bld(newExp(exp,i,0), i + 1)
        val h = bld(newExp(exp,i,1), i + 1)
        mkNode(i, l, h)
      }
    }
    bld(exp, 0)
  }

  // -----------------------------------------------------------------
  // Parallelized ROBDD constructor for boolean expression exp <String>
  def pbuild(sc: SparkContext): Int = {
    printf("\n\nParallel build of ROBDD for Exp = \n%s\n\n", exp)

    def bld(exp: String, i: Int): Int = {
      var l = 0
      var h = 0
      if (i > n) {
        //if (eval(exp)) 1 else 0
        0
      } else if (i == 0) {
        val rdd = sc.parallelize(List(bld(newExp(exp, i, 0), i + 1), bld(newExp(exp, i, 1), i + 1)))
        val results = rdd.take(2).toVector
        l = results(0)
        h = results(1)
        mkNode(i, l, h)
      } else {
        l = bld(newExp(exp, i, 0), i + 1)
        h = bld(newExp(exp, i, 1), i + 1)
        mkNode(i, l, h)
      }
    }
    bld(exp, 0)
  }


  // -----------------------------------------------------------------
  // ITE (If-Then-Else) Operator: computes a logical operation on two ROBDDs
  // Parameters:
  //  * bdd1, bdd2 <ROBDD> = ROBDD objects to apply ITE operator
  //  * op <enum> = Operator value

  def ite(bdd1: ROBDD, bdd2: ROBDD, op: Operators){
    printf("\n\nApply ROBDD operator %s to BE: \nExp[1] = %s\nExp[2] = %s\n\n", op, bdd1.exp, bdd2.exp)
    val u1 = bdd1.root
    val u2 = bdd2.root
    val c = 0

    def app(u1: Int, u2: Int, k: Int): Int = {
      var u = 0
      val c = k + 1
      // Lookup nodes in DHT
      val bdd1_u1 = DHT.lookup((bdd1.bdd_index, u1)).get.asInstanceOf[BDDNode]
      val bdd2_u2 = DHT.lookup((bdd2.bdd_index, u2)).get.asInstanceOf[BDDNode]
      // Memoize lookups
      val memo = DHT.lookup(("memo", bdd_index, u1, u2))
      if (memo.isDefined) return memo.get.asInstanceOf[Int]
      // Apply operation to terminals
      else if ( (u1 == 0 || u1 == 1) && (u2 == 0 || u2 == 1 ) ) {
        if  (op == AND) u = and(u1, u2)
        else if (op == OR) u = or(u1, u2)
      }
      else if (bdd1_u1.i == bdd2_u2.i) {
        u = mkNode(bdd1_u1.i, app(bdd1_u1.low, bdd2_u2.low, c), app(bdd1_u1.high, bdd2_u2.high, c))
      }
      else if (bdd1_u1.i < bdd2_u2.i) {
        u = mkNode(bdd1_u1.i, app(bdd1_u1.low, u2, c), app(bdd1_u1.high, u2, c))
      }
      else { // varb(u1) > varb(u2)
        u = mkNode(bdd2_u2.i, app(u1, bdd2_u2.low, c), app(u1, bdd2_u2.high, c))
      }
      DHT.insert(("memo", bdd_index, u1, u2), u)
      u
    }
    app(u1, u2, c)
  }

  // -----------------------------------------------------------------
  // ITE (If-Then-Else) Operator: computes a logical operation on two ROBDDs
  def pite(bdd1: ROBDD, bdd2: ROBDD, op: Operators, sc: SparkContext): Int = {
    printf("\n\nApply ROBDD operator %s to BE: \nExp[1] = %s\nExp[2] = %s\n\n", op, bdd1.exp, bdd2.exp)
    val u1 = bdd1.root
    val u2 = bdd2.root

    // Initialize G: operations memoization
    val G = new mutable.HashMap[(Int, Int), Int]()
    var rdd = sc.emptyRDD[Int]

    def app(u1: Int, u2: Int): Int = {
      var u = 0
      var i = 0
      // Lookup nodes in hash map
      val bdd1_u1 = DHT.lookup((bdd1.bdd_index, u1)).get.asInstanceOf[BDDNode]
      val bdd2_u2 = DHT.lookup((bdd2.bdd_index, u2)).get.asInstanceOf[BDDNode]
      // Memoized result
      if (G.contains((u1, u2))) G((u1, u2))
      // Apply operation to terminals
      else if ( (u1 == 0 || u1 == 1) && (u2 == 0 || u2 == 1 ) ) {
        if  (op == AND) u = and(u1, u2)
        else if (op == OR) u = or(u1, u2)
      } else {

        if (bdd1_u1.i == bdd2_u2.i) {
          rdd = sc.parallelize(List(app(bdd1_u1.low, bdd2_u2.low), app(bdd1_u1.high, bdd2_u2.high)))
          i = bdd1_u1.i
        }
        else if (bdd1_u1.i < bdd2_u2.i) {
          rdd = sc.parallelize(List(app(bdd1_u1.low, u2), app(bdd1_u1.high, u2)))
          i = bdd1_u1.i
        }
        else { // varb(u1) > varb(u2)
          rdd = sc.parallelize(List(app(u1, bdd2_u2.low), app(u1, bdd2_u2.high)))
          i = bdd2_u2.i
        }
        val results = rdd.take(2).toVector
        val l = results(0)
        val h = results(1)
        u = mkNode(i,l,h)
      }
      G((u1, u2)) = u
      u
    }
    app(u1, u2)

  }


  // AND operator
  private def and(u1: Int, u2: Int): Int = {
    u1 * u2
  }
  // OR operator
  private def or(u1: Int, u2: Int): Int = {
    if (u1 + u2 > 0) 1 else 0
  }


  // -----------------------------------------------------------------
  // Utilities

  // [JEval module] Evaluates boolean expressions passed in as strings
  private def eval(exp: String): Boolean = {
    val engine = new Evaluator()
    engine.getBooleanResult(exp)
  }

  // Implements Shannon expansion by replacing var with boolean value [0,1]
  private def newExp(exp: String, v: Int, value: Int): String = {
    val varStr = "x" + v
    val valStr = value.toString
    val newExpStr = exp.replaceAll(varStr, valStr)
    newExpStr
  }

  // Prints a tabular representation of the ROBDD.
  def printTable(): Unit = {
    val output = StringBuilder.newBuilder
    val labels = List("index", "var", "low", "high")
    //output.append(f"\n\nROBDD Node Table \n Expression = $exp\n")
    output.append("------------------------------------------------------\n")
    output.append(f"${labels.head}%-10s ${labels(1)}%-10s ${labels(2)}%-10s ${labels(3)}%-10s\n")
    for (u <- 0 until nodeCount) {
      val v = DHT.lookup((bdd_index, u)).get.asInstanceOf[BDDNode].i
      val l = DHT.lookup((bdd_index, u)).get.asInstanceOf[BDDNode].low
      val h = DHT.lookup((bdd_index, u)).get.asInstanceOf[BDDNode].high
      var l_str = l.toString
      var h_str = h.toString
      if (u == 0 || u ==1) {
        h_str = "-"
        l_str = "-"
      }
      output.append(f"$u%-10s ${v.toString}%-10s $l_str%-10s $h_str%-10s\n")
    }
    println(output)
  }
}


