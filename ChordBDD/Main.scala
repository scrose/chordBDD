// -----------------------------------------------------------------
// Parallel ROBDD: Application Main
// -----------------------------------------------------------------

import BDD_Structure.AND
import Chord_DHT.Chord
import org.apache.log4j.{Level, Logger}
object Main extends App {

  // Method profiler (timer)
  def time[R](block: => R): R = {
    val t0 = System.nanoTime()
    val result = block    // call-by-name
    val t1 = System.nanoTime()
    println("Elapsed time: " + (t1 - t0) + "ns")
    result
  }

  // Disable logging messages
  Logger.getLogger("org").setLevel(Level.OFF)
  Logger.getLogger("akka").setLevel(Level.OFF)

  // Create Distributed Hash Table singleton
  var DHT = new Chord(12, 10)
  // Transient delay to wait for Chord stabilization (5 seconds)
  Thread.sleep(10000)

  // Create ordered variable list (important for BDD canonicity)
  var vars: Array[String] = Array("x1", "x2", "x3", "x4")

  // Build BDDs on DHT
  var bdd1 = new BDD_Distributed.ROBDD("(x0 && x2) || (!x0 && !x2)", vars, DHT)
  time{bdd1.build()}
  bdd1.printTable()

  var bdd2 = new BDD_Distributed.ROBDD("(x1 && x3) || (!x1 && !x3)", vars, DHT)
  time{bdd2.build()}
  bdd2.printTable()

  var bdd3 = new BDD_Distributed.ROBDD("None", vars, DHT)
  time{bdd3.ite(bdd1, bdd2, AND)}
  bdd3.printTable()

  // Test against Sequential (single-machine) BDDs
  var bdd4 = new BDD_Single.ROBDD("(x0 && x2) || (!x0 && !x2)", vars)
  time{bdd4.build()}
  bdd4.printTable()

  var bdd5 = new BDD_Single.ROBDD("(x1 && x3) || (!x1 && !x3)", vars)
  time{bdd5.build()}
  bdd5.printTable()

  var bdd6 = new BDD_Single.ROBDD("None", vars)
  time{bdd6.ite(bdd4, bdd5, AND)}
  bdd6.printTable()

}