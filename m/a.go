package main

import "fmt"
//import "runtime/debug"
//import "strings"


var s []int

func main() {
  s := []int{2, 5};
  e := 0;
  s[e], e = 7, 1;
  fmt.Println(s)
}
