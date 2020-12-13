package main

import "fmt"
import "strconv"

var L = false

var RP = "x"

const OS = ", "

type branch int

func (_ branch) aBranch() {}

type branchi interface {
	aBranch()
}

type reg int

type regi interface {
	aReg()
}

type regOrConst interface {
}

func ff(a reg) {}

func (r reg) aReg() {}

var rs []string = []string{"TR1", "TR2", "TR3", "TR4", "TR5", "TR6", "TR7", "TR8", "TR9", "TR10", "THP", "TMAIN", "TBP", "TSP", "TSS"}

var irs []string = []string{
	"ax", "bx", "cx", "dx", "si", "di", "bp", "8", "9", "10", "11", "12", "13", "14", "15"}
var ars []string = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "19", "20", "21", "22"}

var fmap = make(map[string]func(*emitter, *CallExpr))

const (
	TR1 reg = iota
	TR2
	TR3
	TR4
	TR5
	TR6
	TR7
	TR8
	TR9
	TR10
	THP
	TMAIN
	TBP
	TSP
	TSS

	RMAX
)

const (
	LR reg = 30
	SP reg = LR + 1 + iota
	XZR
)

const BP = ".br"
const FP = ".f"

var IR reg = -1

func moffOff(a int) int {
	return a * 8
}

func offSet(a, b string) string {
	return fmt.Sprintf("[%v%v%v]", a, OS, b)
}

func init() {
	fmap["assert"] = func(e *emitter, ce *CallExpr) {
		if len(ce.Params) != 2 {
			e.err("")
		}
		e.assignToReg(TR2, ce.Params[0])
		e.assignToReg(TR3, ce.Params[1])
		e.cmp(TR2, TR3)
		lab := e.clab()
		e.br(lab, "eq")
		ln := e.st.Gpos().Line
		e.mov(TR1, ln)
		e.emitExit()
		e.makeLabel(lab)
	}
	fmap["malloc"] = func(e *emitter, ce *CallExpr) {
		if len(ce.Params) != 1 {
			e.err("")
		}
		e.mov(TR1, THP)
		e.assignToReg(TR3, ce.Params[0])
		e.mov(TR2, TR3)
		e.lsl(TR3, 3)
		e.add(THP, TR3)
	}
	fmap["bad"] = func(e *emitter, ce *CallExpr) {
		if len(ce.Params) != 0 {
			e.err("")
		}
		ln := e.st.Gpos().Line
		e.mov(TR1, ln)
		if L {
			e.emitR("push", TMAIN)
		} else {
			e.mov(LR, TMAIN)
		}
		e.emit("ret")
	}

	fmap["len"] = func(e *emitter, ce *CallExpr) {

		if len(ce.Params) != 1 {
			e.err("")
		}
		v := ce.Params[0].(*VarExpr).Wl.Value
		ml := e.rMap[v]
		if ml.mlt == mlVoid {
			e.mov(TR9, -1)
			e.iLoad(TR1, TR9, ml)
			return
		}
		e.mov(TR1, ml.len)
	}
	fmap["exit"] = func(e *emitter, ce *CallExpr) {

		if len(ce.Params) != 1 {
			e.err("")
		}
		e.assignToReg(TR1, ce.Params[0])
		if L {
			e.emitR("push", TMAIN)
		} else {
			e.mov(LR, TMAIN)
		}
		e.emit("ret")

	}

}
func makeReg(i regi) string {
	if i.(reg) >= TR1 && i.(reg) <= TSS {
		return rs[i.(reg)]
	}

	if i == LR {
		return "lr"
	}
	if i == SP {
		if L {
			return RP + "sp"
		}
		return "sp"
	}
	if i == XZR {
		return "xzr"
	}

	return fmt.Sprintf("%v%v", RP, i)
}

func atoi(e errp, s string) int {
	x, err := strconv.ParseInt(s, 0, 0)
	if err != nil {
		e.err(err.Error())
	}
	return int(x)
}

func makeConst(i int, pref bool) string {
	if L {
		if pref {
			return "$" + fmt.Sprint(i)
		}
		return fmt.Sprint(i)
	} else {
		return fmt.Sprintf("#%v", i)
	}
}

func makeBranch(i branchi) string {
	return fmt.Sprintf("%v%v", BP, i)
}
func makeRC(a regOrConst, pref bool) string {
	if a2, ok := a.(reg); ok {
		return makeReg(a2)
	}
	return makeConst(a.(int), pref)
}

type atype int

const (
	ATinvalid atype = iota
	ATeq
	ATpre
	ATpost
)

func localCond(a string) string {
	rt := a
	if L {
		switch a {
		case "eq":
			rt = "e"
		case "gt":
			rt = "g"
		case "lt":
			rt = "l"

		}
	}
	return rt
}

var didPrint = false
