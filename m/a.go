package main

import "fmt"
import "runtime/debug"
//import "strings"


var s []int

func main() {
  defer func() { err := recover(); fmt.Println(err,"errrrrrrrr",string(debug.Stack())) } ()
  var p *int
  _ = p
  *p = 4

}
