package ChordDHT

import ChordDHT.ChordNode.{Init, Insert, LookUp}
import akka.actor.{ActorRef, ActorSystem, Inbox}
import akka.util.Timeout
import com.typesafe.config.ConfigFactory

import scala.concurrent.duration._
// ******************************************************
// Implements Distributed Hash Table using Chord Protocol
// ******************************************************
// Parameter: n = number of nodes in network
class Chord(n: Int) {

  // Initialize a network of distributed nodes in range m
  val m = 7 // Chord ring bit size
  val nodeList: List[Int] = List(2554, 2555, 2556)
  val system = ActorSystem("ClusterSystem", ConfigFactory.load("application.conf"))
  val n0: ActorRef = system.actorOf(ChordNode.props("2553", m), "Node_1")
  implicit val timeout: Timeout = 60.seconds
  implicit val inbox: Inbox = Inbox.create(system)

  // Initialize Chord nodes
  // -----------------------------------------------------
  nodeList.foreach(n => initNode(system.actorOf(ChordNode.props(n.toString, m, Option(n0)), "Node_" + n.toString)))

  def initNode (n: ActorRef)  {
    inbox.send(n, Init(inbox.getRef()))
    inbox.receive(10.seconds).asInstanceOf[Boolean]
    println(f"Node ${n.toString()} initialized.\n")
    Thread.sleep(3000)
  }

  // API Operations
  // TODO: Check if base node has failed prior to lookup
  // -----------------------------------------------------
  // Lookup value at <key>
  def lookup(key: String): Option[List[Int]] = {
    inbox.send(n0, LookUp(key, inbox.getRef()))
    Option(inbox.receive(10.seconds).asInstanceOf[List[Int]])
  }

  // Store value in DHT; returns associated key
  def insert(key: String, data: List[Int]): Int = {
    inbox.send(n0, Insert(key, data, inbox.getRef()))
    inbox.receive(10.seconds).asInstanceOf[Int]
  }
}

