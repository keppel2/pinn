package main

import "fmt"
//import "strings"


var s []int

func main() {
  fmt.Println(len(s))
  fmt.Println(s == nil)
  s = make([]int, 0)
  fmt.Println(len(s))
  fmt.Println(s == nil)
  s = s[0:0]
  fmt.Println(len(s))
  fmt.Println(s == nil)
}
