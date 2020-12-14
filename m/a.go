package main

import "fmt"



func main() {
str := `
sdfj
`
  for _, r := range str {
    if r != '\n' {
    fmt.Printf("printchar(\"%v\");", string(r))
    } else {
    fmt.Println("println();")
    }
  }
}
