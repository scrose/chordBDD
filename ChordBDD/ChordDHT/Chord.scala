// -----------------------------------------------------------------
// Implements Distributed Hash Table using Chord Protocol
//  * Requires node cluster IP addresses configured in configuration
//  * file application.conf. Hash-keys are based on port numbers.
//  Parameter: m = Chord ring bit size
//             n = Number of Chord nodes to create
// -----------------------------------------------------------------

package ChordDHT

import ChordDHT.ChordNode._
import akka.actor.{ActorRef, ActorSystem, PoisonPill}
import akka.pattern.ask
import akka.util.Timeout
import com.typesafe.config._

import scala.concurrent.duration._
import scala.concurrent.{Await, ExecutionContextExecutor}

class Chord(m: Int = 12, n: Int = 2) {

    // Initialize a network of distributed nodes in range m
    val basePort = 2553 // Port address of initial node
    val config: Config = ConfigFactory.load()
      .withValue("akka.loglevel", ConfigValueFactory.fromAnyRef("OFF"))
      .withValue("akka.stdout-loglevel", ConfigValueFactory.fromAnyRef("OFF"))
    val system = ActorSystem("ClusterSystem", config)
    implicit val timeout: Timeout = 8000.seconds
    implicit val ec: ExecutionContextExecutor = system.dispatcher

    var msgCount = 0

    // Create initial Chord node
    val n0: ActorRef = system.actorOf(ChordNode.props(basePort.toString, m), "Node_0")
    val nodeList: List[Int] = List.range(basePort + 1, basePort + n)
    nodeList.foreach(n => system.actorOf(ChordNode.props(n.toString, m, Option(n0)), "Node_" + n.toString))

    // Simulate an involuntary departure of a node
    system.scheduler.schedule(10.seconds, 3000000.seconds)(
      n0 ! PoisonPill
    )(system.dispatcher)


    // API Operations
    // -----------------------------------------------------
    // Lookup value at <key>
    def lookup(key: Any): Option[Any] = {
      msgCount +=1
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

