// -----------------------------------------------------------------
// Implements Cluster Node for DHT Chord Protocol
//  * Requires node cluster IP addresses configured in configuration
//  * file application.conf. Hash-keys are based on port numbers.
//  * Parameters:
//  * ipAddress <string> = IP address of cluster node
//  * mBits <int> = bit size for keys in chord
//  * nodeConnect <ActorRef> = Proxy node for node join operation
// -----------------------------------------------------------------

package ChordDHT

import java.security.MessageDigest

import ChordDHT.ChordNode._
import akka.actor.{Actor, ActorRef, Cancellable, Props}
import akka.util.Timeout
import akka.pattern.ask

import scala.concurrent.{ExecutionContextExecutor, Future}
import scala.concurrent.duration._
import scala.util.{Failure, Success}

class ChordNode (ipAddress: String = "", mBits: Int, nodeConnect : Option[ActorRef] = None) extends Actor {

  val m: Int = mBits // bit size of the Chord ring
  val ip: String = ipAddress
  val id: Int = hashID(ip)
  val ref: ActorRef = self
  var keys: Option[List[Int]] = Option(List[Int]())
  var keysPred: Option[List[Int]] = Option(List[Int]())
  var data: Map[Int, Any] = Map()
  var finger: Array[fingerEntry] = new Array[fingerEntry](m)
  var predecessor: Option[NodeData] = None
  var successor: Option[NodeData] = Option(this.toNodeData)
  var next: Int = 0
  var msgCount: Int = 0

  implicit val ec: ExecutionContextExecutor = context.system.dispatcher
  implicit val timeout: Timeout = 5.seconds

  // ----------------------------------------------------
  // Schedule stabilization and finger fix jobs to run periodically
  val cancellable1: Cancellable = context.system.scheduler.schedule(40.milliseconds, 10.milliseconds)(
    if (successor.isDefined && successor.get.id != id) stabilize()
  )(context.system.dispatcher)
  val cancellable2: Cancellable = context.system.scheduler.schedule(40.milliseconds, 5.milliseconds)(
    updateFingers(next))(context.system.dispatcher)
  val cancellable3: Cancellable = context.system.scheduler.schedule(3.seconds, 3.seconds)(
    printTable("Status Update"))(context.system.dispatcher)
  val cancellable4: Cancellable = context.system.scheduler.schedule(5.seconds, 1.seconds)(
    checkPredecessor()
  )(context.system.dispatcher)

  // Finger Table entry
  class fingerEntry(k: Int, n: NodeData) {
    var start: Int = ((id + math.pow(2, k)) % math.pow(2, m)).toInt
    var interval: (Int, Int) = (start, ((id + math.pow(2, k + 1)) % math.pow(2, m)).toInt)
    var node: Option[NodeData] = Option(n)
  }

  override def preStart(): Unit = {
    println(f"Chord Node[$id] started")
    initFingers()
    join()
  }

  override def postStop(): Unit = {
    cancellable1.cancel()
    cancellable2.cancel()
    cancellable3.cancel()
    cancellable4.cancel()
    println(f"Chord Node[$id] stopped")
  }

  // Node communication
  // -----------------------------------------------------
  def receive: PartialFunction[Any, Unit] = {

    // Respond to initialization ping
    case Init(respondTo) => respondTo ! InitResponse(Option(this.toNodeData))

    // Join node to ring
    case Join(node) =>
      if (node.isDefined) {
        findSuccessor(node.get.id, node.get.ref, join = true)
      }

    // Send successor node to new node
    case JoinResponse(node) =>
      successor = node
      printTable(f"Node[$id]: Successful Join -> Successor Node: ${successor.get.id}.\n")
      // successor.get.moveKeys(this)
      successor.get.ref ! Notify(this.toNodeData)

    // Get successor node to key (in NodeData)
    case FindSuccessor(key, respondTo, retrieve, insert, join) =>
      findSuccessor(key: Int, respondTo, retrieve, insert, join)

    // Response to successor request
    case FoundSuccessor(key, node) =>
      if (node.isDefined) {
        //printf("Node %d [Port: %s]: Found successor to Key %d = %d.\n", id, ip, key, node.get.id)
        if (key == id) successor = node
        // Update finger table
        fixFingers(key, node)
      }

    // Request predecessor of node -> Response: send predecessor
    case Stabilize() => sender ! predecessor

    // Notification of possible predecessor
    case Notify(node) => notify(Option(node))

    // Check predecessor exists
    case CheckPredecessor() => sender ! true

    // API: Insert data and return key
    case Insert(keyValue) =>
      val key: Int = hashID(keyValue.toString)
      findSuccessor(key, sender, insert = true)

    // Insert data at key
    case InsertAt(keyValue, dataInsert) =>
      val key: Int = hashID(keyValue.toString)
      // println(f"Node $id [Port: $ip]: Inserting data for Key[$key].")
      sender ! insert(key, dataInsert)

    // Insert data at key
    case InsertReplica(keysInsert, dataReplica) =>
      //println(f"Node $id [Port: $ip]: Replicating data.")
      keysInsert.foreach(k => insert(k, dataReplica(k), replica = true))


    // API: Lookup key and return data
    case LookUp(keyValue) =>
      val key = hashID(keyValue.toString)
      // println(f"Node $id [Port: $ip]: Lookup Key[$key].")
      findSuccessor(key, sender, retrieve = true)

    // Get data at key
    case Retrieve(key, respondTo) =>
      if (data.contains(key)) {
        //println(f"Node $id [Port: $ip]: Retrieve data[${data(key).toString}] at Key[$key].")
        respondTo ! data(key)
      }
      else {
        // printTable(f"Node $id [Port: $ip]: Data at Key[$key] Not found.")
        respondTo ! None
      }

    // Default for missing case class
    case _ => println(f"Node $id [Port: $ip]: Received Unknown Message")
  }


  // API Operations
  // -----------------------------------------------------

  // Store value in DHT; returns associated key
  // Include replicated data in same data map / different key map
  private def insert(key: Int, value: Any, replica: Boolean = false): Int = {
    if (value != null && !replica) {
      if (!keys.get.contains(key)) keys = Option(keys.getOrElse(List[Int]()) :+ key)
      // Replicate data on successor node
      successor.get.ref ! InsertReplica(List(key), Map(key -> value))
    } else if (value != null && replica)  {
      if (!keysPred.get.contains(key)) keysPred = Option(keysPred.getOrElse(List[Int]()) :+ key)
    }
    else {return -1}
    // TODO: Create key-based check of data mapping
    if (!data.exists(_ == (key,value))) data += (key -> value)
    key
  }

  // Node Failure Handler:
  // Copies replica keys to main data
  private def moveKeys (){
    if (keysPred.isDefined){
      keysPred.get.foreach(k => keys = Option(keys.getOrElse(List[Int]()) :+ k))
    }
    keysPred = Option(List[Int]())
  }

  // Chord Operations
  // -----------------------------------------------------

  // Local node joins Chord system via a remote connected node
  private def join() {
    if (nodeConnect.isDefined) {
      nodeConnect.get ! Join(Option(this.toNodeData))
    } else {
      printTable(f"Node $id [Port: $ip]: First node to join chord.\n")
    }
  }

  // Find successor of key in (n, successor]; otherwise forward query around the ring
  private def findSuccessor(key: Int, respondTo: ActorRef, retrieve: Boolean = false, insert: Boolean = false, join: Boolean = false)  {
    // Check if key is in (n, successor]
    if (inInterval(key, id, successor.get.id, "right")) {
      if (insert) respondTo ! successor
      else if (retrieve) successor.get.ref ! Retrieve(key, respondTo)
      else if (join) respondTo ! JoinResponse(successor)
      else respondTo ! FoundSuccessor(key, successor)
    }
    // Successor not found locally:
    else {
      val n = findPrecedingNode(key)
      // Case 1: Forward to successor (first hop)
      if (respondTo == ref) n.get.ref ! FindSuccessor(key, ref, retrieve, insert, join)
      // Case 2: Forward around the ring (subsequent hops)
      else n.get.ref forward FindSuccessor(key, respondTo, retrieve, insert, join)
    }
  }

  // Search local table for highest predecessor of chordKey,
  // i.e. the closest node preceding chordKey among all the fingers of node
  private def findPrecedingNode (chordKey: Int): Option[NodeData] = {
    // Find finger in (n, chordKey)
    (m - 1 to 0 by -1)
        .find(i => inInterval(finger(i).node.get.id, id, chordKey))
        .map(finger(_).node)
        .getOrElse(Option(this.toNodeData))
  }

  // Periodically verify local node's successor and notify/update in turn
  private def stabilize() {
    val f: Future[Option[NodeData]] = ask(successor.get.ref, Stabilize()).mapTo[Option[NodeData]]
    f.onComplete {
      case Success(node) =>
        if (node.isDefined && inInterval(node.get.id, id, successor.get.id)) {
          if (node.get.id != id) successor = node
        }
      case Failure(e) =>
        //println(f"Node[$id]: <Not Found> Successor[${successor.get.id}].")
        val nextSuccessor = (m - 1 to 0 by -1)
          .find(i => inInterval(finger(i).node.get.id, successor.get.id, id))
          .map(finger(_).node)
          .getOrElse(predecessor)
        successor = nextSuccessor
        successor.get.ref ! InsertReplica(keys.get, data)
    }
    // Notify successor node that local node is possible predecessor
    successor.get.ref ! Notify(this.toNodeData)
  }

  // Notify local node that the remote node might be its predecessor
  // If node has failed, migrate replicated keys to the main data bank
  private def notify(node: Option[NodeData] = None){
    if (predecessor.isEmpty || inInterval(node.get.id, predecessor.get.id, id)) {
      predecessor = node
      // Update default base node successor
      if (successor.get.id == id) successor = node
    }
  }

  // Initialize finger table for single node cluster
  private def initFingers() {
    (m - 1 to 0 by -1).foreach(i => finger(i) = new fingerEntry(i, this.toNodeData))
    this.successor = Option(this.toNodeData)
  }
  // Fix fingers of node based on successor
  private def fixFingers (key: Int, node: Option[NodeData]): Unit = {
    (0 until m)
      .filter(i =>
        inInterval(node.get.id, finger(i).start, finger(i).node.get.id, "right") &&
        !inInterval(node.get.id, finger(i).start, finger(i).interval._2, "right"))
      .foreach(finger(_).node = node)
  }

  // Updates local finger table entries (called periodically).
  private def updateFingers(i: Int): Unit = {
    findSuccessor(finger(next).start, ref)
    next += 1
    if (next == m) next = 0
  }

  // Check status of predecessor node
  private def checkPredecessor(): Unit = {
    if (predecessor.isDefined) {
      val f: Future[Boolean] = ask(predecessor.get.ref, CheckPredecessor()).mapTo[Boolean]
      f.onComplete {
        case Success(result) =>
          //println(f"Node[$id]: <Found> Predecessor[${predecessor.get.id}].")
        case Failure(e) =>
          println(f"Node[$id]: <Not Found> Predecessor[${predecessor.get.id}].")
          predecessor = None
          moveKeys()
      }
    }
  }


  // ----------------------------------------------------
  // Utility functions

  // Check whether x is in [a,b]|[a,b)|(a,b]|(a,b)
  // closed: bracket to close interval
  private def inInterval(x: Int, a: Int, b: Int, closed: String = ""): Boolean = {
    val n = math.pow(2,m).toInt
    val shift = n - a
    if (a == b) return true
    else if (((closed == "left") || (closed == "both")) && (x == a)) return true
    else if (((closed == "right") || (closed == "both")) && (x == b)) return true
    else if (((closed == "") || (closed == "right")) && (x == a)) return false
    else if (((closed == "") || (closed == "left")) && (x == b)) return false
    else if ((x + shift)%n < (b + shift)%n) return true
    false
  }

  // Package node data in immutable data
  private def toNodeData: NodeData = {
    new NodeData(this)
  }

  // Get truncated SHA-1 hash-key
  private def hashID (strVal: String): Int = {
    val hashSize: BigInt = math.pow(2,m).toLong - 1
    // Chord identifier = SHA-1(IP address)
    val shaVal = MessageDigest.getInstance("SHA-1").digest(strVal.getBytes("UTF-8"))
    math.abs(BigInt(shaVal).%(hashSize).toInt)
  }


  // ----------------------------------------------------
  // Status methods

  // Print finger table to stdout
  def printTable(msg: String = ""): Unit = {
    val pred = predecessor.map(_.id).getOrElse("Unknown")
    val succ = successor.map(_.id).getOrElse("Unknown")
    val keyD = keys.map(_.toString()).get
    val keyR = keysPred.map(_.toString()).get
    val hr = "--------------------------------------------------\n"
    var output: String = "\n\n" + msg + f"\nNode[$id%d]: Finger Table\n" + hr
    output += "%1$-5s %2$-10s %3$-15s %4$-10s\n".format("i", "start", "interval", "successor") + hr
    for (i <- 0 until m) {
      val start = finger(i).start
      val interval = finger(i).interval
      val succ = finger(i).node.map(_.id).getOrElse("-")
      output += f"$i%-5s $start%-10s $interval%-15s $succ%-10s\n"
    }
    output += hr + f"Predecessor: $pred%s\n"
    output += f"Successor:$succ%s\n"
    output += f"Keys: $keyD%s\n"
    output += f"Replicated Keys: $keyR%s\n" + hr
    println(output)
  }
}

// Companion object defines message handlers by the ChordNode Actor
object ChordNode {
  sealed trait Command
  final case class Init(respondTo: ActorRef) extends Command
  final case class InitResponse(node: Option[NodeData]) extends Command
  final case class Join(node: Option[NodeData]) extends Command
  final case class JoinResponse(node: Option[NodeData]) extends Command
  final case class FindSuccessor(key: Int, respondTo: ActorRef, retrieve: Boolean, insert: Boolean, join: Boolean) extends Command
  final case class FoundSuccessor(key: Int, successor: Option[NodeData]) extends Command
  final case class Retrieve(key: Int, respondTo: ActorRef) extends Command
  final case class FindPredecessor() extends Command
  final case class Stabilize() extends Command
  final case class CheckPredecessor() extends Command
  final case class Notify(node: NodeData) extends Command
  final case class LookUp(key: Any) extends Command
  final case class Insert(key: Any) extends Command
  final case class InsertAt(key: Any, data: Any) extends Command
  final case class InsertReplica(keys: List[Int], data: Map[Int, Any]) extends Command

  def props(ipAddress: String, m: Int, nodeConnect: Option[ActorRef] = None):
  Props = Props(new ChordNode(ipAddress, m, nodeConnect))
  case object ChordNode
}

// Message object to package immutable Chord Node data
class NodeData (n: ChordNode) {
  var id: Int = n.id
  var ip: String = n.ip
  var ref: ActorRef = n.ref
  var m: Int = n.m
  var keys: List[Int] = n.keys.getOrElse(List[Int]())
}


