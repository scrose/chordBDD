import ChordDHT.Chord
import org.apache.log4j.{Level, Logger}

object Main extends App {

  // Disable logging messages
  Logger.getLogger("org").setLevel(Level.OFF)
  Logger.getLogger("akka").setLevel(Level.OFF)

  // Create new Chord DHT
  var dhtable = new Chord(3)

  // Insert value into Chord Table
  val value = List(0, 3, 6, 18, 27)
  val key = "ExampleList"
  val hashkey = dhtable.insert(key, value)
  println(f"Value Stored at Key[$hashkey]")

  // Lookup value at key
  println(f"Value Retrieved: ${dhtable.lookup(key).getOrElse("Not Found")}")

}
