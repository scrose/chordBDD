package BDD_Structure

sealed trait Operators
  case object AND extends Operators
  case object OR extends Operators
  case object NOT extends Operators
