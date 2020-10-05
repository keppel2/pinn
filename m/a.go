package main

import "fmt"

func f() {}
func main() {
fmt.Println("mn")
f()()
}
