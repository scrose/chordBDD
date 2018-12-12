package BDDStructure

// BDD Node definition
class BDDNode(variable: Int, l: Int, h: Int) extends Serializable  {
  var i: Int = variable
  var high: Int = h
  var low: Int = l
}
