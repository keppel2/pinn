package main

import "io/ioutil"
import "strings"
import "fmt"
import "os"

func main() {
	fmt.Println(os.Args)

	//f()
	src, _ := ioutil.ReadFile(os.Args[1] + ".pinn")
	ssrc := string(src)
	tok(ssrc)
	p := new(parser)
	p.init(strings.NewReader(ssrc))
	p.fileA()
	return

}
