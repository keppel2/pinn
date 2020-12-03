package main

import "math/rand"
import "fmt"
import "reflect"
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

/*
func (e *emitter) _f() {
	e.mov(5, TR1)
}
*/

func (r reg) aReg() {}

var rs []string = []string{"TR1", "TR2", "TR3", "TR4", "TR5", "TR6", "TR7", "TR8", "TR9", "TR10", "TR11", "TMAIN", "TBP", "TSP", "TSS"}

var irs []string = []string{
	"ax", "bx", "cx", "dx", "si", "di", "bp", "8", "9", "10", "11", "12", "13", "14", "15"}
var ars []string = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "19", "20", "21", "22"}

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
	TR11
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

/*
const (
	R0 reg = iota
	R1
	R2
	R3
	R4
	R5
	R6
	R7
	R8
)
*/

const BP = ".br"
const FP = ".f"

var RB reg = TSS + 1
var IR reg = -1

/*
const (
	MLinvalid = iota
	MLreg
	MLstack
	MLheap
)
*/

type mloc struct {
	fc  bool
	i   int
	len int
}

type emitter struct {
	src     string
	rMap    map[string]*mloc
	cbranch branch
	ebranch branch
	moff    int
	soff    int
	lstack  [][2]branch
	fc      bool
	fexitm  map[string]branch
	fexit   branch
	lst     Node
	st      Node
	file    *File
}

func (m *mloc) String() string {
	return fmt.Sprintf("%#v", m)
}

func (m *mloc) init(fc bool) {
	m.fc = fc
}

func moffOff(a int) int {
	return a * 8
}

func (e *emitter) clearL() {
	for k, v := range e.rMap {
		if v.fc {
			delete(e.rMap, k)
		}
	}
}

func (e *emitter) newVar(s string, k Kind) {
	switch t := k.(type) {
	case *SKind:
		if _, ok := e.rMap[s]; ok {
			e.err(s)
		}
		e.mov(TR2, 0)
		e.storeId(s, TR2)

	case *ArKind:
		if _, ok := e.rMap[s]; ok {
			e.err(s)
		}
		ml := new(mloc)
		ml.init(e.fc)
		ml.len = e.atoi(t.Len.(*NumberExpr).Il.Value)
		if e.fc {
			e.soff += ml.len
			ml.i = e.soff
			for i := 0; i < ml.len; i++ {
				e.mov(TR2, 0)
				e.push(TR2)
			}
		} else {
			ml.i = e.moff
			e.moff += ml.len
		}
		e.rMap[s] = ml
	default:
		e.err(s)
	}
}

func (e *emitter) push(r reg) {
	e.str(ATpre, r, TSP, -8)
}

func (e *emitter) pop(r reg) {
	e.ldr(ATpost, r, TSP, 8)
}

func (e *emitter) popx() {
	e.add(TSP, 8)
}
func (e *emitter) pushAll() {

	for i := TR2; i <= TR9; i++ {
		if i != TSP {
			e.push(i)
		}
	}

}
func (e *emitter) popAll() {
	for i := TR9; i >= TR2; i-- {
		if i != TSP {
			e.pop(i)
		}
	}
}
func (e *emitter) setIndex(index reg, m *mloc) {
	e.lsl(index, 3)
	if m.fc {
		e.sub(index, moffOff(m.i))
	} else {
		e.add(index, moffOff(m.i))
	}
}

func (e *emitter) iStore(dest reg, index reg, m *mloc) {
	if m.fc {
		if L {
			e.emit("mov", makeReg(dest), fmt.Sprintf("%v(%v,%v,8)", -moffOff(m.i), makeReg(TSS), makeReg(index)))
		} else {
			e.setIndex(index, m)
			e.str(ATeq, dest, TSS, index)
		}
	} else {
		if L {
			e.emit("mov", makeReg(dest), fmt.Sprintf("%v(%v,%v,8)", moffOff(m.i), makeReg(TBP), makeReg(index)))
		} else {
			e.setIndex(index, m)
			e.str(ATeq, dest, TBP, index)
		}
	}
}
func (e *emitter) iLoad(dest reg, index reg, m *mloc) {
	if m.fc {
		if L {
			e.emit("mov", fmt.Sprintf("%v(%v,%v,8)", -moffOff(m.i), makeReg(TSS), makeReg(index)), makeReg(dest))
		} else {
			e.setIndex(index, m)
			e.ldr(ATeq, dest, TSS, index)
		}
	} else {
		if L {
			e.emit("mov", fmt.Sprintf("%v(%v,%v,8)", moffOff(m.i), makeReg(TBP), makeReg(index)), makeReg(dest))
		} else {
			e.setIndex(index, m)
			e.ldr(ATeq, dest, TBP, index)
		}
	}
}

func offSet(a, b string) string {
	return fmt.Sprintf("[%v%v%v]", a, OS, b)
}

func (e emitter) dString() string {
	return fmt.Sprint(e.st, reflect.TypeOf(e.st), e.rMap)
}

func (e *emitter) emit(i string, ops ...string) {
	const ind = "  "
	const AM = " "
	e.src += ind + i + AM
	if ops != nil {
		e.src += ops[0]
		for _, s := range ops[1:] {
			e.src += OS + s
		}
	}
	e.src += "//" + e.dString() + "\n"
}

func (e *emitter) emitR(i string, ops ...regOrConst) {
	sa := []string{}
	for _, s := range ops {

		sa = append(sa, makeRC(s, true))
	}
	e.emit(i, sa...)
}

func (e *emitter) init(f *File) {
	if L {
		RP = "%r"
	}
	rand.Seed(42)
	e.rMap = make(map[string]*mloc)
	e.fexitm = make(map[string]branch)
	e.cbranch = 1
	e.file = f
}

/*
func (e *emitter) varAt(i int) string {
	for k, v := range e.rMap {
		if v == i {
			return k
		}
	}
	return ""
}
*/

func (e *emitter) clab() branch {
	rt := e.cbranch
	e.cbranch++
	return rt
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

func (e *emitter) atoi(s string) int {
	x, err := strconv.Atoi(s)
	if err != nil {
		e.err(err.Error())
	}
	return x

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

func (e *emitter) label(s string) {
	e.src += s + ":\n"
}

func (e *emitter) makeLabel(i branchi) {
	e.label(fmt.Sprintf("%v%v", BP, i))
}

func (e *emitter) pushloop(a, b branch) {
	e.lstack = append(e.lstack, [2]branch{a, b})
}

func (e *emitter) poploop() [2]branch {
	rt := e.lstack[len(e.lstack)-1]
	e.lstack = e.lstack[0 : len(e.lstack)-1]
	return rt
}

func (e *emitter) peekloop() [2]branch {
	return e.lstack[len(e.lstack)-1]
}

func (e *emitter) err(msg string) {

	panic(fmt.Sprintln(msg, e.dString(), e.src))
}

func (e *emitter) emitPrint() {
	e.label("println")
	if L {
		e.mov(TR8, int('\n'))
		e.push(TR8)
		e.mov(TR1, 0x2000004) //SYSCALL 1 on linux
		e.mov(TR6, 1)         //STDOUT
		e.mov(TR4, 1)         //1 byte
		e.mov(TR5, TSP)
		e.emit("syscall")
		e.add(TSP, 8)
		e.emit("ret")
	} else {
		e.mov(TR1, int('\n'))
		e.push(TR1)
		e.mov(TR1, 1)
		e.mov(TR2, TSP)
		e.mov(TR3, 1)
		e.mov(TR9, 64)
		e.emitR("svc", 0)
		e.add(TSP, 8)
		e.emit("ret")
	}

	e.label("print")
	e.mov(TSS, TSP)
	e.ldr(ATeq, TR5, TSS)

	e.sub(TSP, 17)
	e.mov(TR3, int(','))
	e.str(ATeq, TR3, TSP)
	e.mov(TR2, 0)
	e.mov(TR3, 0)

	lab := e.clab()
	lab2 := e.clab()
	lab3 := e.clab()
	e.makeLabel(lab)
	e.mov(TR4, TR5)
	e.and(TR4, 0xf)
	e.cmp(TR4, 10)
	e.br(lab2, "lt")
	e.add(TR4, int('a'-':'))
	e.makeLabel(lab2)
	e.lsr(TR5, 4)
	e.add(TR4, int('0'))
	e.lsl(TR2, 8)
	e.add(TR2, TR4)
	e.cmp(TR3, 7)
	e.br(lab3, "ne")
	e.str(ATeq, TR2, TSP, 9)
	e.mov(TR2, 0)
	e.makeLabel(lab3)
	e.add(TR3, 1)
	e.cmp(TR3, 16)
	e.br(lab, "ne")
	e.str(ATeq, TR2, TSP, 1)
	if L {
		e.mov(TR1, 0x2000004)
		e.mov(TR6, 1)
		e.mov(TR4, 17)
		e.mov(TR5, TSP)
		e.emit("syscall")

	} else {

		e.mov(TR1, 1)
		e.mov(TR2, TSP)
		e.mov(TR3, 17)
		e.mov(TR9, 64)
		e.emitR("svc", 0)
	}
	e.add(TSP, 17)
	e.emit("ret")
}

func (e *emitter) loadId(v string, r reg) {
	ml, ok := e.rMap[v]
	if !ok {
		e.err(v)
	}
	if ml.fc {
		e.ldr(ATeq, r, TSS, -moffOff(ml.i))
	} else {
		e.ldr(ATeq, r, TBP, moffOff(ml.i))
	}
}

func (e *emitter) storeId(v string, r reg) {
	ml, ok := e.rMap[v]
	if ok {
		if ml.fc {
			e.str(ATeq, r, TSS, -moffOff(ml.i))
		} else {
			e.str(ATeq, r, TBP, moffOff(ml.i))
		}
	} else {
		ml := new(mloc)
		ml.init(e.fc)
		if ml.fc {
			e.push(r)
			e.soff++
			ml.i = e.soff
		} else {
			ml.i = e.moff
			e.str(ATeq, r, TBP, moffOff(ml.i))
			e.moff++
		}
		e.rMap[v] = ml
	}

}

func makeRC(a regOrConst, pref bool) string {
	if a2, ok := a.(reg); ok {
		return makeReg(a2)
	}
	return makeConst(a.(int), pref)
}

func (e *emitter) br(b branchi, s ...string) {
	if L {
		br := "jmp"
		if len(s) == 1 {
			br = "j" + localCond(s[0])
		}
		e.emit(br, makeBranch(b.(branch)))
		return
	}
	br := "b"
	if len(s) == 1 {
		br += "." + s[0]
	}
	e.emit(br, makeBranch(b.(branch)))
}

type atype int

const (
	ATinvalid atype = iota
	ATeq
	ATpre
	ATpost
)

func (e *emitter) str(t atype, d regi, base regi, offset ...regOrConst) {
	if len(offset) == 1 {
		switch t {
		case ATeq:
			if L {
				e.emit("mov", makeReg(d), fmt.Sprintf("%v(%v)", makeRC(offset[0], false), makeReg(base)))
			} else {
				e.emit("str", makeReg(d), offSet(makeReg(base), makeRC(offset[0], true)))
			}

		case ATpre:
			if L {
				e.add(base, offset[0])
				e.emit("mov", makeReg(d), fmt.Sprintf("(%v)", makeReg(base)))
			} else {
				e.emit("str", makeReg(d), fmt.Sprintf("[%v%v%v]!", makeReg(base), OS, makeRC(offset[0], true)))
			}
		case ATpost:
			if L {
				e.emit("mov", makeReg(d), fmt.Sprintf("(%v)", makeReg(base)))
				e.add(base, offset[0])
			} else {
				e.emit("str", makeReg(d), fmt.Sprintf("[%v]%v%v", makeReg(base), OS, makeRC(offset[0], true)))
			}
		}
	} else {
		if L {
			e.emit("mov", makeReg(d), fmt.Sprintf("(%v)", makeReg(base)))
		} else {
			e.emit("str", makeReg(d), fmt.Sprintf("[%v]", makeReg(base)))
		}
	}
}

func (e *emitter) ldr(t atype, d regi, base regi, offset ...regOrConst) {
	if len(offset) == 1 {
		switch t {
		case ATeq:
			if L {
				e.emit("mov", fmt.Sprintf("%v(%v)", makeRC(offset[0], false), makeReg(base)), makeReg(d))
			} else {
				e.emit("ldr", makeReg(d), offSet(makeReg(base), makeRC(offset[0], true)))
			}
		case ATpre:
			if L {
				e.add(base, offset[0])
				e.emit("mov", fmt.Sprintf("(%v)", makeReg(base)), makeReg(d))
			} else {
				e.emit("ldr", makeReg(d), fmt.Sprintf("[%v%v%v]!", makeReg(base), OS, makeRC(offset[0], true)))
			}
		case ATpost:
			if L {
				e.emit("mov", fmt.Sprintf("(%v)", makeReg(base)), makeReg(d))
				e.add(base, offset[0])
			} else {
				e.emit("ldr", makeReg(d), fmt.Sprintf("[%v]%v%v", makeReg(base), OS, makeRC(offset[0], true)))
			}
		}
	} else {
		if L {
			e.emit("mov", fmt.Sprintf("(%v)", makeReg(base)), makeReg(d))
		} else {
			e.emit("ldr", makeReg(d), fmt.Sprintf("[%v]", makeReg(base)))
		}
	}
}

func (e *emitter) nativeOp(op string, a regi, b regOrConst) {
	if L {
		e.emitR(op, b, a)
	} else {
		e.emitR(op, a, a, b)
	}
}

func (e *emitter) cmp(a regi, b regOrConst) {
	if L {
		e.emitR("cmpq", b, a)
	} else {
		e.emitR("cmp", a, b)
	}
}

func (e *emitter) sub(a regi, b regOrConst) {
	e.nativeOp("sub", a, b)
}
func (e *emitter) add(a regi, b regOrConst) {
	e.nativeOp("add", a, b)
}
func (e *emitter) mul(a regi, b regOrConst) {
	if L {
		e.nativeOp("imul", a, b)
	} else {
		e.nativeOp("mul", a, b)
	}

}
func (e *emitter) rem(a regi, b regOrConst) {
	if L {
		e.mov(TR1, a)
		e.mov(TR4, 0)
		e.emitR("div", b)
		e.mov(a, TR4)
	} else {
		e.mov(TR5, a)
		e.emitR("udiv", a, TR5, b)
		e.emitR("msub", a, a, b, TR5)
	}
}
func (e *emitter) div(a regi, b regOrConst) {
	if L {
		e.mov(TR1, a)
		e.mov(TR4, 0)
		e.emitR("div", b)
		e.mov(a, TR1)
	} else {
		e.nativeOp("udiv", a, b)
	}
}
func (e *emitter) and(a regi, b regOrConst) {
	e.nativeOp("and", a, b)
}
func (e *emitter) lsl(a regi, b regOrConst) {
	if L {
		e.nativeOp("sal", a, b)
	} else {
		e.nativeOp("lsl", a, b)
	}
}
func (e *emitter) lsr(a regi, b regOrConst) {
	if L {
		e.nativeOp("shr", a, b)
	} else {
		e.nativeOp("lsr", a, b)
	}
}
func (e *emitter) mov(a regi, b regOrConst) {
	if L {
		e.emitR("mov", b, a)
	} else {
		e.emitR("mov", a, b)
	}
	/*
		a2, ok := a.(reg)
		if !ok {
			e.err("")
		}
	*/

	//	sb := makeRC(b)
	//	e.emit("mov", makeReg(a.(reg)), sb)//.(reg)), sb)
}

func (e *emitter) doOp(dest, b reg, op string) {
	switch op {
	case "+":
		e.add(dest, b)
		return
	case "-":
		e.sub(dest, b)
		return
	case "*":
		e.mul(dest, b)
		return
	case "/":
		e.div(dest, b)
		return
	case "%":
		e.rem(dest, b)
		return
	default:
		e.err(op)
	}
}

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

func (e *emitter) condExpr(dest branch, be *BinaryExpr) {
	if be.op == "||" {
		lab := e.clab()
		lab2 := e.clab()
		e.condExpr(lab, be.LHS.(*BinaryExpr))
		e.br(lab2)
		e.makeLabel(lab)
		e.condExpr(dest, be.RHS.(*BinaryExpr))
		e.makeLabel(lab2)
	} else if be.op == "&&" {
		e.condExpr(dest, be.LHS.(*BinaryExpr))
		e.condExpr(dest, be.RHS.(*BinaryExpr))
	} else if be.op == "==" || be.op == "!=" || be.op == "<" || be.op == "<=" || be.op == ">" || be.op == ">=" {
		e.assignToReg(TR4, be.LHS)
		e.assignToReg(TR2, be.RHS)
		e.cmp(TR4, TR2)
		bi := ""
		switch be.op {
		case "==":
			bi = "ne"
		case "!=":
			bi = "eq"
		case "<":
			bi = "ge"
		case "<=":
			bi = "gt"
		case ">":
			bi = "le"
		case ">=":
			bi = "lt"
		default:
			e.err(be.op)
		}
		e.br(branch(dest), bi)
		return
	} else {
		e.err(be.op)
	}

}

func (e *emitter) binaryExpr(dest reg, be *BinaryExpr) {
	_, okL := be.LHS.(*BinaryExpr)
	_, okR := be.RHS.(*BinaryExpr)
	var first, second Expr
	if okR && !okL {
		first = be.RHS
		second = be.LHS
	} else {
		first = be.LHS
		second = be.RHS
	}

	e.assignToReg(dest, first)
	e.assignToReg(dest+1, second)
	e.doOp(dest, dest+1, be.op)
}

func (e *emitter) emitFunc(f *FuncDecl) {
	e.label(FP + f.Wl.Value)
	e.soff = 0
	e.mov(TSS, TSP)
	for _, field := range f.PList {
		if ark, ok := field.Kind.(*ArKind); ok {
			for _, vd2 := range field.List {

				if _, ok := e.rMap[vd2.Value]; ok {
					e.err(vd2.Value)
				}
				ml := new(mloc)
				ml.init(e.fc)
				plen := e.atoi(ark.Len.(*NumberExpr).Il.Value)
				e.soff += plen
				ml.len = plen
				ml.i = -(f.PSize - e.soff)
				e.rMap[vd2.Value] = ml
			}
			continue

		}
		for _, vd2 := range field.List {

			if _, ok := e.rMap[vd2.Value]; ok {
				e.err(vd2.Value)
			}
			ml := new(mloc)
			ml.init(e.fc)
			e.soff++
			ml.i = -(f.PSize - e.soff)
			e.rMap[vd2.Value] = ml
		}
	}
	e.soff = 0
	lab := e.clab()
	e.ebranch = lab
	e.emitStmt(f.B)
	e.makeLabel(lab)

	e.mov(TSP, TSS)
	e.emit("ret")
	e.clearL()
}

func (e *emitter) assignToReg(r reg, ex Expr) {
	e.lst = e.st
	e.st = ex
	defer func() { e.st = e.lst }()
	switch t2 := ex.(type) {
	case *NumberExpr:
		e.mov(r, e.atoi(t2.Il.Value))
	case *VarExpr:
		e.loadId(t2.Wl.Value, r)
	case *BinaryExpr:
		e.binaryExpr(r, t2)
	case *UnaryExpr:
		if t2.op == "-" {
			e.assignToReg(r, t2.E)
			e.mov(TR11, -1)
			e.mul(r, TR11)
		} else if t2.op == "&" {
			switch t3 := t2.E.(type) {
			case *VarExpr:
				v := t3.Wl.Value
				ml := e.rMap[v]
				e.mov(r, 0)
				e.setIndex(r, ml)
				if ml.fc {
					e.add(r, TSS)
				} else {
					e.add(r, TBP)
				}
			case *IndexExpr:
				v := t3.X.(*VarExpr).Wl.Value
				ml := e.rMap[v]
				e.assignToReg(r, t3.E)
				e.setIndex(r, ml)
				if ml.fc {
					e.add(r, TSS)
				} else {
					e.add(r, TBP)
				}
			}
		} else if t2.op == "*" {
			e.assignToReg(r, t2.E)
			e.ldr(ATeq, r, r)
		}
	case *TrinaryExpr:
		lab := e.clab()
		lab2 := e.clab()
		e.condExpr(lab, t2.LHS.(*BinaryExpr))
		e.assignToReg(r, t2.MS)
		e.br(lab2)
		e.makeLabel(lab)
		e.assignToReg(r, t2.RHS)
		e.makeLabel(lab2)

	case *CallExpr:
		e.emitCall(t2)
		e.mov(r, TR1)
	case *IndexExpr:
		v := t2.X.(*VarExpr).Wl.Value
		ml := e.rMap[v]
		e.assignToReg(r, t2.E)
		e.iLoad(r, r, ml)
	default:
		e.err("")
	}

}

func (e *emitter) emitCall(ce *CallExpr) {
	e.st = ce
	ID := ce.ID.(*VarExpr).Wl.Value
	fun := e.file.getFunc(ID)
	if fun == nil {
		e.err(ID)
	}
	if len(ce.Params) != fun.PCount {
		e.err(ID)
	}

	fn := FP + ID
	if ID == "assert" {
		e.assignToReg(TR2, ce.Params[0])
		e.assignToReg(TR3, ce.Params[1])
		e.cmp(TR2, TR3)
		lab := e.clab()
		e.br(lab, "eq")
		ln := e.st.Gpos().Line
		e.mov(TR1, ln)
		if L {
			e.emitR("push", TMAIN)
		} else {
			e.mov(LR, TMAIN)
		}
		e.emit("ret")
		e.makeLabel(lab)
		return
	} else if ID == "bad" {
		ln := e.st.Gpos().Line
		e.mov(TR1, ln)
		if L {
			e.emitR("push", TMAIN)
		} else {
			e.mov(LR, TMAIN)
		}
		e.emit("ret")
		return
	} else if ID == "exit" {
		e.assignToReg(TR1, ce.Params[0])
		if L {
			e.emitR("push", TMAIN)
		} else {
			e.mov(LR, TMAIN)
		}
		e.emit("ret")
		return
	} else if ID == "print" {
		fn = ID
		didPrint = true
	} else if ID == "println" {
		didPrint = true
		fn = ID
	}

	// e.pushP()
	e.pushAll()
	e.push(TSS)
	if !L {
		e.push(LR)
	}

	for k, v := range ce.Params {
		//		e.push(1 + reg(k))
		kind := fun.getKind(k)
		if ie, ok := v.(*VarExpr); ok && e.rMap[ie.Wl.Value].len > 0 {
			if e.atoi(kind.(*ArKind).Len.(*NumberExpr).Il.Value) != e.rMap[ie.Wl.Value].len {
				e.err(ID)
			}
			ml := e.rMap[ie.Wl.Value]
			for i := ml.len - 1; i >= 0; i-- {
				e.mov(TR2, i)
				e.iLoad(TR2, TR2, ml)
				e.push(TR2)
			}
		} else {
			if kind != nil {
				if _, ok := kind.(*SKind); !ok {
					e.err(ID)
				}
			}

			e.assignToReg(TR2, v)
			e.push(TR2)
		}
	}

	if L {
		e.emit("call", fn)
	} else {
		e.emit("bl", fn)
	}
	e.add(TSP, moffOff(fun.PSize))
	if !L {
		e.pop(LR)
	}
	e.pop(TSS)

	e.popAll()

}

func (e *emitter) emitStmt(s Stmt) {
	e.st = s
	e.emit("//")
	switch t := s.(type) {
	case *ExprStmt:
		e.assignToReg(TR2, t.Expr)
	case *BlockStmt:
		for _, s := range t.SList {
			e.emitStmt(s)
		}
	case *ContinueStmt:
		e.br(e.peekloop()[0])
	case *BreakStmt:
		e.br(e.peekloop()[1])
	case *LoopStmt:
		lab := e.clab()
		e.makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab, lab2)
		e.emitStmt(t.B)
		e.br(lab)
		e.makeLabel(lab2)
		e.poploop()
	case *WhileStmt:
		lab := e.clab()
		e.makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab, lab2)
		e.condExpr(lab2, t.Cond.(*BinaryExpr))
		e.emitStmt(t.B)
		e.br(lab)
		e.makeLabel(lab2)
		e.poploop()

	case *IfStmt:
		lab := e.clab()
		if t.Else == nil {
			e.condExpr(lab, t.Cond.(*BinaryExpr))
			e.emitStmt(t.Then)
		} else {
			lab2 := e.clab()
			e.condExpr(lab2, t.Cond.(*BinaryExpr))
			e.emitStmt(t.Then)
			e.br(lab)
			e.makeLabel(lab2)
			e.emitStmt(t.Else)
		}
		e.makeLabel(lab)

	case *ReturnStmt:
		if t.E != nil {
			e.assignToReg(TR1, t.E)
		} else {
			e.mov(TR1, 5)
		}
		if L {
		} else {
		}
		e.br(e.ebranch)
	case *AssignStmt:
		lh := t.LHSa[0]
		switch lh2 := lh.(type) {
		case *UnaryExpr:
			if lh2.op != "*" {
				e.err(lh2.op)
			}
			e.assignToReg(TR3, lh2.E)
			e.assignToReg(TR2, t.RHSa[0])
			e.str(ATeq, TR2, TR3)
		case *VarExpr:
			id := lh2.Wl.Value
			if t.Op == ":=" && e.rMap[id] != nil {
				e.err(id)
			}
			if t.Op == "=" && e.rMap[id] == nil {
				e.err(id)
			}
			if t.Op == "+=" || t.Op == "-=" || t.Op == "/=" || t.Op == "*=" || t.Op == "%=" || t.Op == "++" || t.Op == "--" {
				e.loadId(id, TR2)
				if t.Op[1:2] == "=" {
					e.assignToReg(TR3, t.RHSa[0])
				} else {
					e.mov(TR3, 1)
				}
				e.doOp(TR2, TR3, t.Op[0:1])
				e.storeId(id, TR2)
				return
			}
			if ae, ok := t.RHSa[0].(*ArrayExpr); ok {
				k := new(ArKind)
				k.Init(e.st.Gpos())
				aLen := new(NumberExpr)
				aLen.Il = new(ILit)
				aLen.Il.Value = fmt.Sprint(len(ae.EL))
				k.Len = aLen
				e.newVar(id, k)
				for key, expr := range ae.EL {
					e.assignToReg(TR2, expr)
					e.mov(TR3, key)
					e.iStore(TR2, TR3, e.rMap[id])
				}
				return

			}

			e.assignToReg(TR2, t.RHSa[0])
			e.storeId(id, TR2)

		case *IndexExpr:
			if t.Op == "+=" || t.Op == "-=" || t.Op == "/=" || t.Op == "*=" || t.Op == "%=" || t.Op == "++" || t.Op == "--" {
				e.assignToReg(TR2, lh2)
				if t.Op[1:2] == "=" {
					e.assignToReg(TR3, t.RHSa[0])
				} else {
					e.mov(TR3, 1)
				}
				e.doOp(TR2, TR3, t.Op[0:1])
			} else {
				e.assignToReg(TR2, t.RHSa[0])
			}

			v := lh2.X.(*VarExpr).Wl.Value
			ml := e.rMap[v]
			e.assignToReg(TR3, lh2.E)
			e.iStore(TR2, TR3, ml)
		default:
			e.err("")
		}

	case *VarStmt:
		for _, v := range t.List {
			e.newVar(v.Value, t.Kind)
		}
	case *ForStmt:
		if t.Inits != nil {
			e.emitStmt(t.Inits)
		}

		lab := e.clab()
		lab2 := e.clab()
		lab3 := e.clab()
		e.pushloop(lab, lab2)
		e.br(lab3)
		e.makeLabel(lab)
		if t.Loop != nil {
			e.emitStmt(t.Loop)
		}
		e.makeLabel(lab3)

		if t.E != nil {
			e.condExpr(lab2, t.E.(*BinaryExpr))
		}
		e.emitStmt(t.B)
		e.br(lab)

		e.makeLabel(lab2)
		e.poploop()

	default:
		e.err("")

	}

}

func (e *emitter) emitDefines() {
	if L {
		for r := TR1; r <= TSS; r++ {
			e.src += "#define " + rs[r] + " " + fmt.Sprintf("%v%v", RP, irs[r]) + "\n"
		}
	} else {
		for r := TR1; r <= TSS; r++ {
			e.src += "#define " + rs[r] + " " + fmt.Sprintf("%v%v", RP, ars[r]) + "\n"
		}
	}
}

var didPrint = false

func (e *emitter) emitF() {
	e.emitDefines()
	e.src += ".global _main\n"
	e.label("_main")
	if L {
		e.emitR("pop", TMAIN)
		e.emitR("push", TMAIN)
	} else {
		e.mov(TMAIN, LR)
	}
	e.mov(TSP, SP)
	e.sub(TSP, 0x100)
	e.mov(TSS, TSP)
	e.mov(TBP, TSP)
	e.sub(TBP, 0x10000)
	lab := e.clab()
	e.ebranch = lab
	for _, s := range e.file.SList {
		e.emitStmt(s)
	}
	e.mov(TR1, 0)
	e.makeLabel(lab)
	e.emit("ret")
	e.clearL()
	e.fc = true
	for _, s := range e.file.FList {
		if s.B != nil {
			e.emitFunc(s)
		}
	}
	if didPrint {
		e.emitPrint()
	}
}
