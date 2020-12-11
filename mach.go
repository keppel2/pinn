package main

import "debug/macho"
import "fmt"

func main() {
  f, _ := macho.Open("m.out")
  ss, _ := f.ImportedLibraries()
  fmt.Println(ss)
  fmt.Printf("%+v\n", f)
}
