package main

import "io/ioutil"
import "strings"
import "fmt"
import "os"

func g() {
	ts := TypeStmt{}
	pnode(ts)
}

func main() {
	fmt.Println(os.Args)

	//	g()
	src, _ := ioutil.ReadFile(os.Args[1] + ".pinn")
	ssrc := string(src)
	tok(ssrc)
	p := new(parser)
	p.init(strings.NewReader(ssrc))
	p.fileA()
	return

}
