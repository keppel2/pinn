package main

//a
import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strings"
)

var fFlag = flag.String("f", "a", "Filename")
var sFlag = flag.Bool("s", false, "Run scanner")
var pFlag = flag.Bool("p", false, "Parse")
var vFlag = flag.Bool("v", false, "Run visitor")
var oFlag = flag.String("o", "a.S", "Output")
var aFlag = flag.String("a", "x64", "x64 or arm64")
var eFlag = flag.String("e", "darwin", "darwin or linux")

func g() {
	//	ts := TypeStmt{}
	//	pnode(ts)
}

func f2() {
	// rd, _ := ioutil.ReadDir(os.Args[1])
	//  for _, ofi := range rd {

}

func main() {
	flag.Parse()
	if len(os.Args) <= 1 {
		os.Exit(1)
	}

	src, err := ioutil.ReadFile(*fFlag + ".pinn")
	if err != nil {
		panic(err)
	}
	ssrc := string(src)
	if *sFlag {
		s := new(scan)
		s.init(strings.NewReader(ssrc))
		fmt.Println(s.tokenize())
		return

	}
	p := new(parser)
	p.init(strings.NewReader(ssrc))

	f := p.fileA()
	if *pFlag {
		return
	}
	if *vFlag {
		v := new(visitor)
		v.init()
		v.visitFile(f)
		fmt.Println(v.s)
		return
	}

	e := emitter{}
	e.init(f)
	if *aFlag == "x64" {
		e.a = acx
	} else if *aFlag == "arm64" {
		e.a = acArm
	} else {
		e.err("")
	}
	if *eFlag == "darwin" {
		e.e = enDarwin
	} else if *aFlag == "linux" {
		e.e = enLinux
	} else {
		e.err("")
	}

	_ = debug.Stack
	defer func() {
		err := recover()
		if err != nil {
			fmt.Fprintln(os.Stderr, e.p.sb.String(), "STK", string(debug.Stack()))
			fmt.Fprintln(os.Stderr, "MSG", err, "EDS", e.ds)
			os.Exit(1)
		}
	}()
	e.emitF()

	ioutil.WriteFile(*oFlag, []byte(e.p.sb.String()), 0666)

	//fmt.Println(e.p.sb.String())
	return
}
