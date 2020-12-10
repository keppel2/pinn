package main

import "fmt"
type t int



func main() {
  f := func(tt t, a int)(i int) {
    fmt.Println(i)
    return 42
  }
  var x t
  x = 101
  rt := f(x)
  fmt.Println(rt)
}
