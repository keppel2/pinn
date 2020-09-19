package main

import "fmt"
import "text/scanner"
import "strings"

func main() {
src := `var x int;
y := 10abc "dkj ;`

var s scanner.Scanner
s.Init(strings.NewReader(src))
for {
tok := s.Scan()
fmt.Printf("%v: %v\n", s.Position, s.TokenText())
if tok == scanner.EOF {
  break
}
}

}
