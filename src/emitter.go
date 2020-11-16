package main

import "math/rand"
import "fmt"
import "reflect"
import "strconv"

const RP = "x"

func bw() int {
	if RP == "w" {
		return 4
	}
	return 8
}

const (
	TR = 29 - iota
	TR2
	TR3
	TR4
	TR5
	TMAIN
	TBP
	TSP
	RMAX
)
const BP = ".br"
const FP = ".f"

var RB = 9

const (
	MLinvalid = iota
	MLreg
	MLstack
	MLheap
)

type mloc struct {
	Mlt int
	i   int
	len int
}

type emitter struct {
	src     string
	rMap    map[string]mloc
	rAlloc  [RMAX + 1]string
	cbranch int
	moff    int
	lstack  [][2]int
	lst     Node
	st      Node
}

func (m *mloc) init(mlt int, a int) {
	m.Mlt = mlt
	m.i = a
}

func (e *emitter) findReg() int {
	for k, v := range e.rAlloc[RB:] {
		if v == "" {
			return k + RB
		}
	}
	return e.freeReg()
}

func moffOff(a int) int {
	return a * bw()
}

func (e *emitter) newVar(s string, k Kind) {
	switch t := k.(type) {
	case *SKind:
		//    e.fillReg(s, false)
	case *ArKind:
		ml := mloc{}
		ml.init(MLheap, e.moff)
		x, _ := strconv.Atoi(t.Len.(*NumberExpr).Il.Value)
		ml.len = x
		e.moff += x
		e.rMap[s] = ml

	}
}

func (e *emitter) push(s string) {
	e.emit("str", s, "["+makeReg(TSP), "-8]!")
}

func (e *emitter) pop(s string) {
	e.emit("ldr", s, "["+makeReg(TSP)+"]", "8")
}
func (e *emitter) popP() {
	for i := RB - 1; i >= 0; i-- {
		e.pop(makeReg(i))
	}
}

func (e *emitter) pushP() {
	for i := 0; i < RB; i++ {
		e.push(makeReg(i))
	}
}

func (e *emitter) freeReg() int {
	k := rand.Intn(RMAX + 1 - RB)
	k += RB
	s := e.rAlloc[k]
	e.rAlloc[k] = ""
	ml := mloc{}
	ml.init(MLreg, e.moff)
	e.rMap[s] = ml
	e.emit("str", makeReg(k), "["+makeXReg(TBP), fmt.Sprintf("%v]", moffOff(e.moff)))
	e.moff++
	return k
}

func (e *emitter) emit(i string, ops ...string) {
	const ind = "  "
	const OS = ", "
	const AM = " "
	e.src += ind + i + AM
	if ops != nil {
		e.src += ops[0]
		for _, s := range ops[1:] {
			e.src += OS + s
		}
	}
	e.src += "//" + fmt.Sprint(e.rMap, e.rAlloc, e.st, reflect.TypeOf(e.st)) + "\n"
}

func (e *emitter) init() {
	rand.Seed(42)
	e.rMap = make(map[string]mloc)
	e.cbranch = 1
	//	e.moff = -1
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

func (e *emitter) clab() int {
	rt := e.cbranch
	e.cbranch++
	return rt
}

func makeReg(i int) string {
	return fmt.Sprintf("%v%v", RP, i)
}

func makeXReg(i int) string {
	return fmt.Sprintf("x%v", i)
}

func makeBranch(i int) string {
	return fmt.Sprintf("%v%v", BP, i)
}

func (e *emitter) makeLabel(i int) {
	e.src += fmt.Sprintf("%v%v:\n", BP, i)
}

func (e *emitter) pushloop(a, b int) {
	e.lstack = append(e.lstack, [2]int{a, b})
}

func (e *emitter) poploop() [2]int {
	rt := e.lstack[len(e.lstack)-1]
	e.lstack = e.lstack[0 : len(e.lstack)-1]
	return rt
}

func (e *emitter) peekloop() [2]int {
	return e.lstack[len(e.lstack)-1]
}

func (e *emitter) err(msg string) {

	panic(fmt.Sprintln(msg, e.rMap, e.rAlloc, e.src))
}
func (e *emitter) print(a int) {
	e.pushP()
	e.emit("mov", makeReg(0), "1")
	e.emit("mov", makeReg(1), "0")
	e.emit("mov", makeReg(TR2), makeReg(a))
	e.emit("sub", makeReg(TSP), makeReg(TSP), "16")
	for i := 0; i < 16; i++ {
		e.emit("and", makeReg(TR), makeReg(TR2), "0xf")
		e.emit("cmp", makeReg(TR), "10")
		lab := e.clab()
		e.emit("b.lt", makeBranch(lab))
		e.emit("add", makeReg(TR), makeReg(TR), "('a' - ':')")
		e.makeLabel(lab)
		e.emit("lsr", makeReg(TR2), makeReg(TR2), "4")
		e.emit("add", makeReg(TR), makeReg(TR), "'0'")
		e.emit("lsl", makeReg(1), makeReg(1), "8")
		e.emit("add", makeReg(1), makeReg(1), makeReg(TR))
		if i == 7 {
			e.emit("str", makeReg(1), fmt.Sprintf("[%v, 8]", makeReg(TSP)))

			e.emit("mov", makeReg(1), "0")
		}
	}
	e.emit("str", makeReg(1), fmt.Sprintf("[%v]", makeReg(TSP)))

	e.emit("mov", makeReg(1), makeReg(TSP))
	e.emit("mov", makeReg(2), "16")
	e.emit("mov", makeReg(8), "64")
	e.emit("svc", "0")
	e.emit("add", makeReg(TSP), makeReg(TSP), "16")
	e.popP()

}

func (e *emitter) fillReg(s string, load bool) int {
	ml, ok := e.rMap[s]
	if ok && ml.Mlt == MLreg {
		return ml.i
	}

	k := e.findReg()
	ml = mloc{}
	ml.init(MLreg, k)
	ml.i = k
	e.rMap[s] = ml
	e.rAlloc[ml.i] = s
	if ok && load {
		e.emit("ldr", makeReg(k), "["+makeXReg(TBP), fmt.Sprintf("%v]", moffOff(ml.i)))
	}

	return k

}

func (e *emitter) regOrImm(ex Expr) string {
	rt := ""
	switch t := ex.(type) {
	case *NumberExpr:
		rt = "#" + t.Il.Value
	case *VarExpr:
		i := e.fillReg(t.Wl.Value, true)
		rt = RP + fmt.Sprint(i)
	default:
		e.err("")
	}
	return rt

}

func (e *emitter) operand(ex Expr) string {
	rt := ""
	switch t := ex.(type) {
	case *NumberExpr, *VarExpr:
		rt += e.regOrImm(t)
	default:
		e.err("")
	}
	return rt
}

func (e *emitter) moveToTr(ex Expr) string {
	e.st = ex
	if v, ok := ex.(*NumberExpr); ok {
		e.emit("mov", makeReg(TR), e.regOrImm(v))
		return makeReg(TR)
	}
	return e.regOrImm(ex)
}

func (e *emitter) binaryExpr(dest int, be *BinaryExpr) {
	e.lst = e.st
	e.st = be
	defer func() { e.st = e.lst }()
	if be.op == "==" || be.op == "!=" || be.op == "<" || be.op == "<=" || be.op == ">" || be.op == ">=" {
		lh := e.moveToTr(be.LHS)
		rh := e.regOrImm(be.RHS)
		e.emit("cmp", lh, rh)
		bi := ""
		switch be.op {
		case "==":
			bi = "NE"
		case "!=":
			bi = "EQ"
		case "<":
			bi = "GE"
		case "<=":
			bi = "GT"
		case ">":
			bi = "LE"
		case ">=":
			bi = "LT"
		}
		e.emit("b."+bi, makeBranch(dest))
		return
	}
	switch t := be.LHS.(type) {
	case *NumberExpr, *VarExpr:
		rh := e.operand(t)
		e.emit("mov", makeReg(dest), rh)
	case *BinaryExpr:
		e.binaryExpr(dest, t)
	case *CallExpr:
		e.emitCall(t)
		e.emit("mov", makeReg(dest), RP+"0")
	}

	op := ""
	rh := ""
	if t, ok := be.RHS.(*CallExpr); ok {
		e.emitCall(t)
		rh = RP + "0"
	}
	switch be.op {
	case "+":
		op = "add"
		fallthrough
	case "-":
		if op == "" {
			op = "sub"
		}
		if rh == "" {
			rh = e.regOrImm(be.RHS)
		}
	case "*", "/":
		if be.op == "*" {
			op = "mul"
		} else {
			op = "udiv"
		}
		if rh == "" {
			rh = e.moveToTr(be.RHS)
		}
	case "%":
		if rh == "" {
			rh = e.moveToTr(be.RHS)
		}
		e.emit("udiv", makeReg(TR2), makeReg(dest), rh)
		e.emit("msub", makeReg(dest), makeReg(TR2), rh, makeReg(dest))
		return
	}
	e.emit(op, makeReg(dest), makeReg(dest), rh)
}

func (e *emitter) emitFunc(f *FuncDecl) {
	e.src += FP + f.Wl.Value + ":\n"
	reg := 1
	for _, vd := range f.PList {
		for _, vd2 := range vd.List {
			ml := mloc{}
			ml.init(MLreg, reg)
			e.rMap[vd2.Value] = ml
			e.rAlloc[reg] = vd2.Value
			reg++
		}
	}
	e.emitStmt(f.B)
	e.emit("ret")
}

func (e *emitter) assignToReg(r int, ex Expr) {
	switch t2 := ex.(type) {
	case *NumberExpr, *VarExpr:
		rh := e.operand(t2)
		e.emit("mov", makeReg(r), rh)
	case *BinaryExpr:
		e.binaryExpr(r, t2)
	case *CallExpr:
		e.emitCall(t2)
		e.emit("mov", makeReg(r), makeReg(0))
	case *IndexExpr:
		v := t2.X.(*VarExpr).Wl.Value
		ml := e.rMap[v]
		e.assignToReg(TR4, t2.E)
		e.emit("add", makeReg(TR4), makeReg(TR4), fmt.Sprint(ml.i))
		e.emit("lsl", makeReg(TR4), makeReg(TR4), "#3")

		//		x, _ := strconv.Atoi(t2.E.(*NumberExpr).Il.Value)
		e.emit("ldr", makeReg(r), "["+makeXReg(TBP), fmt.Sprintf("%v]", makeReg(TR4)))

	default:
		e.err("")
	}

}

func (e *emitter) emitCall(ce *CallExpr) {
	e.st = ce
	ID := ce.ID.(*VarExpr).Wl.Value
	// e.pushP()
	e.push(makeReg(TR3))
	for k, v := range ce.Params {
		e.push(makeReg(k + 1))
		e.assignToReg(k+1, v)
	}
	e.push("lr")
	e.emit("bl", FP+ID)
	e.pop("lr")
	for k, _ := range ce.Params {
		e.pop(makeReg(k + 1))
	}
	e.pop(makeReg(TR3))
	//  e.popP()

}

func (e *emitter) emitStmt(s Stmt) {
	e.st = s
	switch t := s.(type) {
	case *ExprStmt:
		ce := t.Expr.(*CallExpr)
		ID := ce.ID.(*VarExpr).Wl.Value
		if ID == "assert" {
			e.assignToReg(TR4, ce.Params[0])
			e.assignToReg(TR5, ce.Params[1])
			e.emit("cmp", makeReg(TR4), makeReg(TR5))
			lab := e.clab()
			e.emit("b.eq", makeBranch(lab))
			e.emit("mov", RP+"0", "1")
			e.emit("mov", "lr", makeXReg(TMAIN))
			e.emit("ret")
			e.makeLabel(lab)
		} else if ID == "bad" {
			e.emit("mov", RP+"0", "1")
			e.emit("mov", "lr", makeXReg(TMAIN))
			e.emit("ret")
		} else if ID == "print" {
			e.assignToReg(TR2, ce.Params[0])
			e.print(TR2)
		} else {
			e.emitCall(ce)
		}

	case *BlockStmt:
		for _, s := range t.SList {
			e.emitStmt(s)
		}
	case *ContinueStmt:
		e.emit("b", makeBranch(e.peekloop()[0]))
	case *BreakStmt:
		e.emit("b", makeBranch(e.peekloop()[1]))
	case *LoopStmt:
		lab := e.clab()
		e.makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab, lab2)
		e.emitStmt(t.B)
		e.emit("b", makeBranch(lab))
		e.makeLabel(lab2)
		e.poploop()
	case *WhileStmt:
		lab := e.clab()
		e.makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab, lab2)
		e.binaryExpr(lab2, t.Cond.(*BinaryExpr))
		e.emitStmt(t.B)
		e.emit("b", makeBranch(lab))
		e.makeLabel(lab2)

	case *IfStmt:
		lab := e.clab()
		if t.Else == nil {
			e.binaryExpr(lab, t.Cond.(*BinaryExpr))
			e.emitStmt(t.Then)
		} else {
			lab2 := e.clab()
			e.binaryExpr(lab2, t.Cond.(*BinaryExpr))
			e.emitStmt(t.Then)
			e.emit("b", makeBranch(lab))
			e.makeLabel(lab2)
			e.emitStmt(t.Else)
		}
		e.makeLabel(lab)

	case *ReturnStmt:
		if t.E != nil {
			e.assignToReg(TR3, t.E)
		}
		e.emit("mov", makeReg(0), makeReg(TR3))
		e.emit("ret")
	case *AssignStmt:
		lh := t.LHSa[0]
		switch t2 := lh.(type) {
		case *VarExpr:
			lhi := e.fillReg(t2.Wl.Value, false)
			e.assignToReg(lhi, t.RHSa[0])
		case *IndexExpr:
			e.assignToReg(TR5, t.RHSa[0])
			v := t2.X.(*VarExpr).Wl.Value
			ml := e.rMap[v]
			e.assignToReg(TR4, t2.E)
			e.emit("add", makeReg(TR4), makeReg(TR4), fmt.Sprint(ml.i))
			e.emit("lsl", makeReg(TR4), makeReg(TR4), "#3")
			//			x, _ := strconv.Atoi(t2.E.(*NumberExpr).Il.Value)
			e.emit("str", makeReg(TR5), "["+makeXReg(TBP), fmt.Sprintf("%v]", makeReg(TR4)))

			// v2 := t2.E.
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
		e.emit("b", makeBranch(lab3))
		e.makeLabel(lab)
		if t.Loop != nil {
			e.emitStmt(t.Loop)
		}
		e.makeLabel(lab3)

		if t.E != nil {
			e.binaryExpr(lab2, t.E.(*BinaryExpr))
		}
		e.emitStmt(t.B)
		e.emit("b", makeBranch(lab))

		e.makeLabel(lab2)

	default:
		e.err("")

	}

}

func (e *emitter) emitF(f *File) {
	e.src = `
.global main
main:
`
	e.emit("mov", makeXReg(TMAIN), "lr")
	e.emit("sub", "sp", "sp", "0x100")
	e.emit("mov", makeXReg(TSP), "sp")
	e.emit("sub", "sp", "sp", "0x10000")
	e.emit("mov", makeXReg(TBP), "sp")
	for _, s := range f.SList {
		e.emitStmt(s)
	}
	e.emit("mov", RP+"0", RP+"zr")
	e.emit("ret")
	for _, s := range f.FList {
		e.emitFunc(s)
	}
}
