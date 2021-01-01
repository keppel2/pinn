package main

import "fmt"
import "strconv"
import "os"

var _ = os.Stderr

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
	"ax", "10", "cx", "dx", "si", "di", "11", "8", "9", "bx", "bp", "12", "13", "14", "15"}
var ars []string = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "19", "20", "21", "22", "23", "24"}

var fmap = make(map[string]func(*emitter, *CallExpr) *mloc)

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
	NR
)

const BP = ".br"
const FP = ".f"

var IR reg = -1

func fmake(s string) string {
	return FP + s
}

func push(sa []string, b string) []string {
	rt := append(sa, b)
	return rt
}

func pop(sa []string) ([]string, string) {
	rt := sa[len(sa)-1]
	sa = sa[0 : len(sa)-1]
	return sa, rt
}

func rpn(sa []string) string {
	stack := make([]string, 0)
	for _, s := range sa {
		if s == "+" || s == "-" || s == "*" || s == "/" || s == "%" || s == "^" {
			var smb, sma string
			stack, smb = pop(stack)
			mb := atoi(nil, smb)
			stack, sma = pop(stack)
			ma := atoi(nil, sma)
			var m int
			switch s {
			case "+":
				m = ma + mb
			case "-":
				m = ma - mb
			case "*":
				m = ma * mb
			case "/":
				m = ma / mb
			case "%":
				m = ma % mb
			case "^":
				m = 1
				for i := 0; i < mb; i++ {
					m *= ma
				}
			}
			stack = push(stack, fmt.Sprint(m))
		} else {
			stack = push(stack, s)
		}
	}
	stack, rt := pop(stack)
	return rt
}

func moffOff(a int) int {
	return a * 8
}

func revString(a string) string {
	rt := ""
	for _, ch := range a {
		rt = string(ch) + rt
	}
	return rt
}
func offSet(a, b string) string {
	return fmt.Sprintf("[%v%v%v]", a, OS, b)
}

func init() {
	fmap["dbg"] = func(e *emitter, ce *CallExpr) *mloc {
		if len(ce.Params) != 0 {
			e.err("")
		}
		e.p.mov(TR4, TSP)
		return newSent(rsInt)
	}
	fmap["assert"] = func(e *emitter, ce *CallExpr) *mloc {
		if len(ce.Params) != 2 {
			e.err("")
		}
		e.assignToReg(ce.Params[0])
		e.p.push(TR2)
		e.assignToReg(ce.Params[1])
		e.p.pop(TR3)
		e.p.cmp(TR3, TR2)
		lab := e.clab()
		e.p.br(lab, "eq")
		e.p.mov(TR1, TR2)
		e.p.mov(TR2, TR3)
		e.p.emit2Print()
		e.p.mov(TR2, TR1)
		e.p.emit2Print()
		e.p.emit2Prints("--assert,")
		e.p.emitLC()
		e.p.mov(TR1, 5)
		e.p.emitExit()
		e.p.makeLabel(lab)
		return nil
	}
	fmap["malloc"] = func(e *emitter, ce *CallExpr) *mloc {
		if len(ce.Params) != 1 {
			e.err("")
		}
		e.assignToReg(ce.Params[0])
		e.p.mov(TR4, THP)
		e.p.lsl(TR2, 3)
		e.p.add(THP, TR2)
		return newSent(rsMloc)
	}
	fmap["bad"] = func(e *emitter, ce *CallExpr) *mloc {
		if len(ce.Params) != 0 {
			e.err("")
		}
		ln := e.st.Gpos().Line
		e.p.mov(TR1, ln)
		e.p.emitExit()
		return nil
	}

	fmap["len"] = func(e *emitter, ce *CallExpr) *mloc {
		rt := newSent(rsInt)
		if len(ce.Params) != 1 {
			e.err("")
		}
		v := ce.Params[0].(*VarExpr).Wl.Value
		ml := e.rMap[v]
		if ml.mlt == mlVoid {
			e.err(v)
		} else if ml.mlt == mlSlice {
			e.p.mov(TR5, 0)
			e.iLoad(TR4, TR5, ml)
			return rt
		}
		e.p.mov(TR4, ml.len)
		return rt
	}
	fmap["exit"] = func(e *emitter, ce *CallExpr) *mloc {

		if len(ce.Params) != 1 {
			e.err("")
		}
		e.assignToReg(ce.Params[0])
		e.p.mov(TR1, TR2)
		e.p.emitExit()
		return nil
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
		e.err(err.Error() + "," + s)
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

func makeVar(e Expr) string {
	return e.(*VarExpr).Wl.Value
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

var didPrint = true
