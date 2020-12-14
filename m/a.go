package main

import "fmt"
import "strings"



func main() {
  var b strings.Builder
  b.WriteString("ignit")
  fmt.Printf("%v", b)
}
