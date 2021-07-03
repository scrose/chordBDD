package BDDStructure

import scala.io.Source.fromFile

class ExpParse(inputFilePath: String) extends Serializable {
  val exp: String = readFile(inputFilePath)

  private def readFile(inputFilePath: String): String = {
    val boolExpr: String = fromFile(inputFilePath).getLines.mkString
    boolExpr
  }

  private def printUsage(): Unit = {
    val usage =
      """BDD Parallel
        |Usage: <boolean function>
        |Expression - (String) Boolean function using ordered <variables> and <operators>
        |Variables - Single characters (e.g. A, B, C)
        |Operators - ^ = AND, + = OR, ~ = NOT""".stripMargin
    println(usage)
  }

  def getVector(n: Int): Seq[Boolean] = {
    println(n.toByte)
    0 until n.toByte map isBitSet(n.toByte)
  }
  def isBitSet(n: Byte)(bit: Int): Boolean =
    ((n >> bit) & 1) == 1
}
