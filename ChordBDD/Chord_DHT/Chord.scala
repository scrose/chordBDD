// -----------------------------------------------------------------
// Implements Distributed Hash Table using Chord Protocol
//  * Requires node cluster IP addresses configured in configuration
//  * file application.conf. Hash-keys are based on port numbers.
//  Parameter: m = Chord ring bit size
//             n = Number of Chord nodes to create
// -----------------------------------------------------------------

package Chord_DHT

import Chord_DHT.ChordNode._
import akka.actor.{ActorRef, ActorSystem}
import akka.pattern.ask
import akka.util.Timeout
import com.typesafe.config.ConfigFactory
import scala.concurrent.duration._
import scala.concurrent.{Await, ExecutionContextExecutor}

class Chord(m: Int = 12, n: Int = 2) {
    // Initialize a network of distributed nodes in range m
    val basePort = 2553 // Port address of initial node
    val system = ActorSystem("ClusterSystem", ConfigFactory.load("application.conf"))
    implicit val timeout: Timeout = 8000.seconds
    implicit val ec: ExecutionContextExecutor = system.dispatcher

    // Create initial Chord node
    val n0: ActorRef = system.actorOf(ChordNode.props(basePort.toString, m), "Node_0")
    val nodeList: List[Int] = List.range(basePort + 1, basePort + n)
    nodeList.foreach(n => system.actorOf(ChordNode.props(n.toString, m, Option(n0)), "Node_" + n.toString))

    // API Operations
    // -----------------------------------------------------
    // Lookup value at <key>
    def lookup(key: Any): Option[Any] = {
      val f1 = n0 ? LookUp(key)
      Await.result(f1, timeout.duration).asInstanceOf[Option[Any]]
    }

    // Store value in DHT; returns associated key
    def insert(key: Any, data: Any): Int = {
      // Get key successor
      val f1 = n0 ? Insert(key)
      val successor = Await.result(f1, timeout.duration).asInstanceOf[Option[NodeData]]
      val f2 = successor.get.ref ? InsertAt(key, Option(data))
      Await.result(f2, timeout.duration).asInstanceOf[Int]
    }
}

