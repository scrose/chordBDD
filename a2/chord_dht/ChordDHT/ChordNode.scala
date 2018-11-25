package ChordDHT

import java.security.MessageDigest

import ChordDHT.ChordNode._
import akka.actor.{Actor, ActorRef, Props}
import akka.pattern.ask
import akka.util.Timeout

import scala.concurrent.duration._
import scala.concurrent.{ExecutionContextExecutor, Future}
import scala.util.{Failure, Success}


// ******************************************************
// Implements Cluster Node for DHT Chord Protocol
// ******************************************************
class ChordNode (ipAddress: String = "", mBits: Int, nodeConnect : Option[ActorRef] = None) extends Actor {

  val m: Int = mBits // bit size of the Chord ring
  val ip: String = ipAddress
  val id: Int = hashID(ip)
  val ref: ActorRef = self
  var keys: Option[List[Int]] = None
  var finger: Array[fingerEntry] = new Array[fingerEntry](m)
  var predecessor: Option[NodeData] = None
  var successor: Option[NodeData] = Option(this.toNodeData)
  var next: Int = 0
  var data: Map[Int, List[Int]] = Map()

  implicit val ec: ExecutionContextExecutor = context.system.dispatcher
  implicit val timeout: Timeout = 3.seconds


  // Finger Table entry
  class fingerEntry (k: Int, n: NodeData) {
    var start: Int = ((id + math.pow(2, k))%math.pow(2,m)).toInt
    var interval: (Int, Int) = (start, ((id + math.pow(2, k + 1))%math.pow(2,m)).toInt)
    var node: Option[NodeData] = Option(n)
  }

  // Get truncated SHA-1 key hashx
  def hashID (strVal: String): Int = {
    val hashSize: BigInt = math.pow(2,m).toLong - 1
    // Chord identifier = SHA-1(IP address)
    val shaVal = MessageDigest.getInstance("SHA-1").digest(strVal.getBytes("UTF-8"))
    math.abs(BigInt(shaVal).%(hashSize).toInt)
  }

  // Initialize Finger Table
  initFingers()

  // Join Chord
  join()

  // ----------------------------------------------------
  // Schedule stabilization and finger fix jobs to run every 5 seconds
  context.system.scheduler.schedule(2.seconds, 500.milliseconds)(
    if (successor.isDefined && successor.get.id != id) stabilize()
  )(context.system.dispatcher)
  context.system.scheduler.schedule(2.seconds, 50.milliseconds)(updateFingers(next))(context.system.dispatcher)
  context.system.scheduler.schedule(5.seconds, 5.seconds)(printFinger())(context.system.dispatcher)

  // Node communication
  // -----------------------------------------------------
  def receive: PartialFunction[Any, Unit] = {

    // Respond to initialization ping
    case Init(respondTo) => respondTo ! true

    // Get request to join cluster -> Response: send successor
    case Join(newNode) =>
      sender ! findSuccessor(newNode.id).getOrElse(None)
      // Set the default base node successor to the first join node
      if (successor.get.id == this.id) successor = Option(newNode)

    // Get successor node to key (in NodeData)
    case FindSuccessor(key) =>
        printf("Node %d [Port: %s]: Find successor to Key %d.\n", id, ip, key)
        sender ! findSuccessor(key).getOrElse(None)

    // Request predecessor of node -> Response: send predecessor
    case FindPredecessor() => sender ! predecessor

    // Notification of possible predecessor
    case Notify(refNode) =>
        notify(Option(refNode))

    // API: Insert data and return key
    // TODO: create wrapper for local method Insert()
    case Insert(keyInsert, dataInsert, respondTo) =>
      val key: Int = hashID(keyInsert.toString)
      val n = findSuccessor(key).get
      val f: Future[Int] = ask(n.ref, InsertAt(dataInsert, key)).mapTo[Int]
      f.onComplete {
        case Success(result)=> respondTo ! result
        case Failure(e)=>
          println(" in failure")
          e.printStackTrace()
      }

    // API: Lookup key and return data
    // TODO: create wrapper for local method LookUp()
    case LookUp(reqkey, respondTo) =>
      val key = hashID(reqkey)
      printFinger(f"Node $id [Port: $ip]: Lookup Key[$key].\n")
      val n = findSuccessor(key).get
      val f: Future[List[Int]] = ask(n.ref, Retrieve(key)).mapTo[List[Int]]
      f.onComplete {
        case Success(result)=>
          respondTo ! result
          printf("Found: Node %d [Port: %s] is successor for Key[%d].\n", n.id, n.ip, key)
        case Failure(e)=>
          println(" in failure")
          e.printStackTrace()
      }

    // Insert data at key
    case InsertAt(dataInsert, key) =>
      printf("Node %d [Port: %s]: Inserting data for Key[%d].\n", id, ip, key)
      sender ! insert(dataInsert, key)

    // Get data at key
    case Retrieve(key) =>
      //printFinger(f"Node $id [Port: $ip]: Retrieve data[${data(key).toString}] at Key[$key].\n")
      sender ! data(key)

    // Default for missing case class
    case _ => println(f"Node $id [Port: $ip]: Received Unknown Message")
  }

  // Package node data in immutable data
  def toNodeData: NodeData = {
    new NodeData(this)
  }

  // API Operations
  // -----------------------------------------------------
  // Lookup value at <key>
  /*def lookup(key: Int): Option[List[Int]] = {

  }*/
  // Store value in DHT; returns associated key
  def insert(value: List[Int], key: Int): Int = {
    keys = Option(keys.getOrElse(List[Int]()) :+ key)
    data += (key -> value)
    printf("Node %d [Port: %s]: Inserted data [%s] at Key[%d].\n", id, ip, data(key), key)
    key
  }

  // Chord Operations
  // -----------------------------------------------------

  // Local node joins Chord system via a remote connected node
  def join(): Unit ={
    if (nodeConnect.isDefined) {
      val f: Future[NodeData] = ask(nodeConnect.get, FindSuccessor(id)).mapTo[NodeData]
      f.onComplete {
        case Success(result)=>
          successor = Option(result)
          buildFingers(successor)
          printFinger(f"Node[$id]: Successful Join -> Successor Node: ${successor.get.id}.\n")
          // successor.get.moveKeys(this)
          successor.get.ref ! Notify(this.toNodeData)
        case Failure(e)=>
          println(" in failure")
          e.printStackTrace()
      }
    } else {
      printFinger(f"Node $id [Port: $ip]: First node to join chord.\n")
    }
  }

  // Find successor of key in (n, successor]; otherwise forward query around the circle
  def findSuccessor(chordKey: Int): Option[NodeData] = {
    if (inInterval(chordKey, id, successor.get.id, "right")) successor
    else {
      val n = findPrecedingNode(chordKey)
      if (n.get.id == id) return n
      val f: Future[NodeData] = ask(n.get.ref, FindSuccessor(chordKey)).mapTo[NodeData]
      f.onComplete {
        case Success(result)=> Option(result)
        case Failure(e)=>
          println(" in failure")
          e.printStackTrace()
      }
      n
    }
  }

  // Check if ith finger of node n comes between local node and chordKey,
  // This is the closest node preceding id among all the fingers of node
  def findPrecedingNode (chordKey: Int): Option[NodeData] = {
    // Find finger in (n, chordKey)
    (m - 1 to 0 by -1)
        .find(i => inInterval(finger(i).node.get.id, id, chordKey))
        .map(finger(_).node)
        .getOrElse(Option(this.toNodeData))
  }

  // Periodically verify local node's successor and notify/update in turn
  def stabilize() {
    val f: Future[Option[NodeData]] = ask(successor.get.ref, FindPredecessor()).mapTo[Option[NodeData]]
    f.onComplete {
      case Success(result)=>
        if (result.isDefined && inInterval(result.get.id, id, successor.get.id)) {
          if (result.get.id != id) successor = result
          printFinger(f"Node $id [Port: $ip]: Stabilize ~ Successor: ${successor.get.id}, Succ.Pred: ${result.get.id}\n")
        }
      case Failure(e)=>
        println(" in failure")
        e.printStackTrace()
    }
    // Notify successor node that local node is possible predecessor
    successor.get.ref ! Notify(this.toNodeData)
  }

  // Notify local node that the remote node might be its predecessor
  def notify(notifyNode: Option[NodeData] = None){
    if (predecessor.isEmpty || inInterval(notifyNode.get.id, predecessor.get.id, id)) {
      predecessor = notifyNode
      // Update default base node successor
      if (successor.get.id == id) successor = notifyNode
      printf(f"Node $id [Port: $ip] Update: Predecessor: ${predecessor.get.id}, Successor: ${successor.get.id}.\n")
    }
  }

  // Initialize finger table for single node cluster
  def initFingers() {
    (m - 1 to 0 by -1).foreach(i => finger(i) = new fingerEntry(i, this.toNodeData))
    this.successor = Option(this.toNodeData)
  }
  // Build fingers of node based on successor
  def buildFingers (refNode: Option[NodeData]): Unit = {
    if (successor.isDefined) {
      // First non-trivial finger
      val i0 = (math.floor(math.log(successor.get.id - id)) + 1).toInt
      // Find further successors
      (0 until i0).foreach(i => finger(i).node = findSuccessor(i + math.pow(2, id - 1).toInt))
    }
  }
  // Updates local finger table entries (called periodically).
  def updateFingers(i: Int): Unit = {
    finger(next).node = findSuccessor(id + math.pow(2, next).toInt)
    next += 1
    if (next == m) next = 0
  }

  // TODO: Node removal (and recovery)
  def leave (n: NodeData){
    println("Node %d leaves.", id, n.id )
  }
  // TODO: Moves keys from local node to successor node n
  def moveKeys (n: NodeData){
    println("Move Keys %d -> %d.", id, n.id )
  }

  // ----------------------------------------------------
  // Utility functions

  // Check whether x is in [a,b]|[a,b)|(a,b]|(a,b)
  // closed: bracket to close interval
  def inInterval(x: Int, a: Int, b: Int, closed: String = ""): Boolean = {
    val n = math.pow(2,m).toInt
    val shift = n - a
    if (a == b) return true
    else if (((closed == "left")||(closed == "both")) && (x == a)) return true
    else if (((closed == "right")||(closed == "both")) && (x == b)) return true
    else if ((x + shift)%n < (b + shift)%n) return true
    false
  }

  // ----------------------------------------------------
  // Status methods

  // Print finger table to stdout
  def printFinger(msg: String = ""): Unit = {
    val pred = predecessor.map(_.id).getOrElse("Unknown")
    val succ = successor.map(_.id).getOrElse("Unknown")
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
    output += f"Keys: ${keys.map(_.toString()).getOrElse("Empty")}%s\n" + hr
    println(output)
  }
}

// Companion object defines message handlers by the ChordNode Actor
object ChordNode {
  sealed trait Command
  final case class Init(respondTo: ActorRef) extends Command
  final case class Join(node: NodeData) extends Command
  final case class FindSuccessor(key: Int) extends Command
  final case class Retrieve(key: Int) extends Command
  final case class FindPredecessor() extends Command
  final case class Notify(node: NodeData) extends Command
  final case class LookUp(key: String, respondTo: ActorRef) extends Command
  final case class Insert(key: String, data: List[Int], respondTo: ActorRef) extends Command
  final case class InsertAt(data: List[Int], key: Int) extends Command

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


