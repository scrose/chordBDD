// -----------------------------------------------------------------
// Reduced Ordered Binary Decision Diagram (ROBDD) - Non-Distributed Package
//  * Object for representing and manipulating BDDs
// Parameters:
//  * expr <string> = Boolean expression in standard Java eval notation
//  * vars <array[<string>]> = List of ordered variable string identifiers
// -----------------------------------------------------------------

package BDDSingle

import BDDStructure._
import net.sourceforge.jeval.Evaluator

import scala.collection.mutable

// -----------------------------------------------------------------
class BDD(val expr: String = "", vars: Array[String]) extends Serializable {
  val n: Int = vars.length
  var nodeCount: Int = 0
  var exp: String = expr.replaceAll("""[\n\r\s]+""", "")

  // -----------------------------------------------------------------
  //  Initialize T (Unique BDD Node Table): represents nodes in ROBDD
  var nodeTable: mutable.HashMap[Int, BDDNode] = initT()

  private def initT(): mutable.HashMap[Int, BDDNode] = {
    nodeTable = new mutable.HashMap[Int, BDDNode]()
    // Create terminal nodes [0,1]
    nodeTable(0) = new BDDNode(n, 0, 0)
    nodeTable(1) = new BDDNode(n, 0, 0)
    nodeCount += 2
    nodeTable
  }

  // -----------------------------------------------------------------
  // Initialize H (Inverse of T): maps (variable,low,high) triples to node index in T
  var H = new mutable.HashMap[(Int, Int, Int), Int]()
  // Add node to lookup table H
  private def insert(node: Int, i: Int, low: Int, high: Int): Boolean = {
    H((i, low, high)) = node
    true
  }
  // Test if node is hashed
  private def isMember(i: Int, low: Int, high: Int): Boolean = {
    H.contains((i, low, high))
  }
  // Look up BDD node in unique table
  private def lookup(i: Int, low: Int, high: Int): Int =  {
    H.getOrElse((i, low, high), -1)
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
  // Add unique node to T
  private def add(i: Int, low: Int, h: Int): Int = {
    val index = nodeCount
    val node = new BDDNode(i, low, h)
    nodeTable(index) = node
    nodeCount += 1
    index
  }

  // -----------------------------------------------------------------
  // Construct ROBDD for boolean expression exp <String>
  def build(): Int = {
    printf("\n\nSequential build of ROBDD for Exp = \n%s\n\n", exp)
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
  // ITE (If-Then-Else) Operator: computes a logical operation on two ROBDDs
  // Parameters:
  //  * bdd1, bdd2 <ROBDD> = ROBDD objects to apply ITE operator
  //  * op <enum> = Operator value

  def ite(bdd1: BDD, bdd2: BDD, op: Operators) {
    printf("\n\nApply ITE operator %s to expressions: \nExp[1] = %s\nExp[2] = %s (Sequential)\n", op, bdd1.exp, bdd2.exp)
    val u1 = bdd1.root
    val u2 = bdd2.root

    // Initialize G: operations memoization
    val G = new mutable.HashMap[(Int, Int), Int]()
    val c = 0

    def app(u1: Int, u2: Int, k: Int): Int = {
      var u = 0
      val c = k + 1
      // Lookup nodes in hash map
      val bdd1_u1 = bdd1.nodeTable(u1)
      val bdd2_u2 = bdd2.nodeTable(u2)
      // Memoized result
      if (G.contains((u1, u2))) return G((u1, u2))
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
      G((u1, u2)) = u
      u
    }
    app(u1, u2, c)
  }


  // -----------------------------------------------------------------
  // Utilities

  // Return root node index of ROBDD
  private def root: Int = {
    nodeTable.size - 1
  }

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

  private def and(u1: Int, u2: Int): Int = {
    u1 * u2
  }

  private def or(u1: Int, u2: Int): Int = {
    if (u1 + u2 > 0) 1 else 0
  }

  // Prints a tabular representation of the ROBDD.
  def printTable(): Unit = {
    val output = StringBuilder.newBuilder
    val labels = List("index", "var", "low", "high")
    //output.append(f"\n\nROBDD Node Table\nExpression = $exp\n")
    output.append("------------------------------------------------------\n")
    output.append(f"${labels.head}%-10s ${labels(1)}%-10s ${labels(2)}%-10s ${labels(3)}%-10s\n")
    for (u <- 0 until nodeTable.size) {
      val v = nodeTable(u).i
      val l = nodeTable(u).low
      val h = nodeTable(u).high
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