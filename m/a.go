package main

import "fmt"
var _ = fmt.Print
//import "runtime/debug"
//import "strings"


var s []int

func f() (int, int) { return 2, 5}

func main() {


  a, b, c := f(), 7
}
