package main

import "io/ioutil"
import "strings"
import "fmt"
import "os"

func g() {
	ts := TypeStmt{}
	pnode(ts)
}

func f2() {
	// rd, _ := ioutil.ReadDir(os.Args[1])
	//  for _, ofi := range rd {

}

func main() {
	//	fmt.Println(os.Args)

	//	g()
	src, _ := ioutil.ReadFile(os.Args[1] + ".pinn")
	ssrc := string(src)
//	tok(ssrc)
	p := new(parser)
	p.init(strings.NewReader(ssrc))
	f := p.fileA()
	if len(os.Args) > 2 {

	visitFile(f)
}
	e := emitter{}
	e.init()
	s := e.emit(f)
	fmt.Println(s)
	return

}
