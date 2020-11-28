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
	if len(os.Args) <= 1 {
		os.Exit(1)
	}

	//	g()
	n := new(File)
	p2 := new(parser)
	n.Init(p2.p)
	src, err := ioutil.ReadFile(os.Args[1] + ".pinn")
	if err != nil {
		panic(err)
	}
	ssrc := string(src)
	//	tok(ssrc)
	p := new(parser)
	p.init(strings.NewReader(ssrc))
	f := p.fileA()
	if len(os.Args) > 2 {
		if os.Args[2] == "x86_64" {
			L = true
		} else if os.Args[2] == "parse" {
			return
		} else {

			visitFile(f)
		}
	}
	e := emitter{}
	e.init()
	e.emitF(f)
	fmt.Println(e.src)
	return

}
