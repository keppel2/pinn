package main

import "math/rand"
import "fmt"
import "reflect"
import "strconv"

const RP = "x"

const OS = ", "

const (
	TR1 = 29 - iota
	TR2
	TR3
	TR4
	TR5
	TRV
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
	rMap    map[string]*mloc
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
	return a * 8
}

func (e *emitter) newVar(s string, k Kind) {
	switch t := k.(type) {
	case *SKind:
	case *ArKind:
		ml := new(mloc)
		ml.init(MLheap, e.moff)
		ml.len = e.atoi(t.Len.(*NumberExpr).Il.Value)
		e.moff += ml.len
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

func offSet(a, b string) string {
	return fmt.Sprintf("[%v%v%v]", a, OS, b)
}

func (e *emitter) freeReg() int {
	k := rand.Intn(RMAX + 1 - RB)
	k += RB
	s := e.rAlloc[k]
	e.rAlloc[k] = ""
	ml, ok := e.rMap[s]
	if !ok {
		e.err("")
	}
	ml.Mlt = MLheap
	ml.i = moffOff(e.moff)
	e.moff++
	e.emit("str", makeReg(k), offSet(makeReg(TBP), makeConst(ml.i)))
	return k
}

func (e emitter) dString() string {
	return fmt.Sprint(e.rMap, e.rAlloc, e.st, reflect.TypeOf(e.st))
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

func (e *emitter) init() {
	rand.Seed(42)
	e.rMap = make(map[string]*mloc)
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

func (e *emitter) atoi(s string) int {
	x, err := strconv.Atoi(s)
	if err != nil {
		e.err(err.Error())
	}
	return x

}

func makeConst(i int) string {
	return fmt.Sprintf("#%v", i)
}

func makeBranch(i int) string {
	return fmt.Sprintf("%v%v", BP, i)
}

func (e *emitter) label(s string) {
	e.src += s + ":\n"
}

func (e *emitter) makeLabel(i int) {
	e.label(fmt.Sprintf("%v%v", BP, i))
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

	panic(fmt.Sprintln(msg, e.dString(), e.src))
}
func (e *emitter) print(a int) {
	e.pushP()
	e.emit("mov", makeReg(0), makeConst(1))
	e.emit("mov", makeReg(1), makeConst(0))
	e.emit("mov", makeReg(TR2), makeReg(a))
	e.emit("sub", makeReg(TSP), makeReg(TSP), makeConst(24))
	for i := 0; i < 16; i++ {
		e.emit("and", makeReg(TR1), makeReg(TR2), makeConst(0xf))
		e.emit("cmp", makeReg(TR1), makeConst(10))
		lab := e.clab()
		e.emit("b.lt", makeBranch(lab))
		e.emit("add", makeReg(TR1), makeReg(TR1), "('a' - ':')")
		e.makeLabel(lab)
		e.emit("lsr", makeReg(TR2), makeReg(TR2), makeConst(4))
		e.emit("add", makeReg(TR1), makeReg(TR1), "'0'")
		e.emit("lsl", makeReg(1), makeReg(1), makeConst(8))
		e.emit("add", makeReg(1), makeReg(1), makeReg(TR1))
		if i == 7 {
			e.emit("str", makeReg(1), offSet(makeReg(TSP), makeConst(16)))

			e.emit("mov", makeReg(1), makeConst(0))
		}
	}
	e.emit("str", makeReg(1), offSet(makeReg(TSP), makeConst(8)))
	e.emit("mov", makeReg(TR1), "','")
	e.emit("str", makeReg(TR1), fmt.Sprintf("[%v]", makeReg(TSP)))

	e.emit("mov", makeReg(1), makeReg(TSP))
	e.emit("mov", makeReg(2), makeConst(24))
	e.emit("mov", makeReg(8), makeConst(64))
	e.emit("svc", makeConst(0))
	e.emit("add", makeReg(TSP), makeReg(TSP), makeConst(24))
	e.popP()

}

func (e *emitter) fillReg(s string, load bool) int {
	ml, ok := e.rMap[s]
	if ok && ml.Mlt == MLreg {
		return ml.i
	}
	if !ok && load {
		e.err("")
	}
	k := e.findReg()
	off := -1
	if !ok {
		ml = new(mloc)
		ml.init(MLreg, k)
	} else {
		ml.Mlt = MLreg
		off = ml.i
		ml.i = k
	}
	e.rMap[s] = ml
	e.rAlloc[ml.i] = s
	if load {
		e.emit("ldr", makeReg(k), offSet(makeReg(TBP), makeConst(off)))
	}

	return k

}

func (e *emitter) regLoad(ex Expr) int {
	rt := -1
	switch t := ex.(type) {
	case *NumberExpr:
		e.emit("mov", makeReg(TR1), makeConst(e.atoi(t.Il.Value)))
		rt = TR1
	case *VarExpr:
		i := e.fillReg(t.Wl.Value, true)
		rt = i
	default:
		e.err("")
	}
	return rt

}

func (e *emitter) binaryExpr(dest int, be *BinaryExpr) {
	e.lst = e.st
	e.st = be
	defer func() { e.st = e.lst }()
	if be.op == "==" || be.op == "!=" || be.op == "<" || be.op == "<=" || be.op == ">" || be.op == ">=" {
		e.assignToReg(TR3, be.LHS)
		e.assignToReg(TR4, be.RHS)
		e.emit("cmp", makeReg(TR3), makeReg(TR4))
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
		movr := e.regLoad(t)
		e.emit("mov", makeReg(dest), makeReg(movr))
	case *BinaryExpr:
		e.binaryExpr(dest, t)
	case *CallExpr:
		e.emitCall(t)
		e.emit("mov", makeReg(dest), makeReg(0))
	}
	op := ""
	rh := -1
	if t, ok := be.RHS.(*CallExpr); ok {
		e.emitCall(t)
		rh = 0
	} else {
		rh = e.regLoad(be.RHS)
	}
	switch be.op {
	case "+":
		op = "add"
		fallthrough
	case "-":
		if op == "" {
			op = "sub"
		}
	case "*", "/":
		if be.op == "*" {
			op = "mul"
		} else {
			op = "udiv"
		}
	case "%":
		e.emit("udiv", makeReg(TR2), makeReg(dest), makeReg(rh))
		e.emit("msub", makeReg(dest), makeReg(TR2), makeReg(rh), makeReg(dest))
		return
	}
	e.emit(op, makeReg(dest), makeReg(dest), makeReg(rh))
}

func (e *emitter) emitFunc(f *FuncDecl) {
	e.src += FP + f.Wl.Value + ":\n"
	reg := 1
	for _, vd := range f.PList {
		for _, vd2 := range vd.List {
			ml := new(mloc)
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
		rh := e.regLoad(t2)
		e.emit("mov", makeReg(r), makeReg(rh))
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
		e.emit("lsl", makeReg(TR4), makeReg(TR4), makeConst(3))

		e.emit("ldr", makeReg(r), "["+makeReg(TBP), fmt.Sprintf("%v]", makeReg(TR4)))

	default:
		e.err("")
	}

}

func (e *emitter) emitCall(ce *CallExpr) {
	e.st = ce
	ID := ce.ID.(*VarExpr).Wl.Value
	// e.pushP()
	e.push(makeReg(TRV))
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
	e.pop(makeReg(TRV))
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
			e.emit("mov", makeReg(0), makeConst(1))
			e.emit("mov", "lr", makeReg(TMAIN))
			e.emit("ret")
			e.makeLabel(lab)
		} else if ID == "bad" {
			e.emit("mov", makeReg(0), makeConst(1))
			e.emit("mov", "lr", makeReg(TMAIN))
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
			e.assignToReg(TRV, t.E)
		}
		e.emit("mov", makeReg(0), makeReg(TRV))
		e.emit("ret")
	case *AssignStmt:
		lh := t.LHSa[0]
		switch t2 := lh.(type) {
		case *VarExpr:
			lhi := e.fillReg(t2.Wl.Value, false)
			if t.Op == "++" {
				e.emit("add", makeReg(lhi), makeReg(lhi), makeConst(1))
				return
			} else if t.Op == "--" {
				e.emit("sub", makeReg(lhi), makeReg(lhi), makeConst(1))
				return
			}
			e.assignToReg(lhi, t.RHSa[0])
		case *IndexExpr:
			if t.Op == "++" || t.Op == "--" {
				e.assignToReg(TR5, t2)
				if t.Op == "++" {
					e.emit("add", makeReg(TR5), makeReg(TR5), makeConst(1))
				} else {
					e.emit("sub", makeReg(TR5), makeReg(TR5), makeConst(1))
				}
			} else {
				e.assignToReg(TR5, t.RHSa[0])
			}

			v := t2.X.(*VarExpr).Wl.Value
			ml := e.rMap[v]
			e.assignToReg(TR4, t2.E)
			e.emit("add", makeReg(TR4), makeReg(TR4), fmt.Sprint(ml.i))
			e.emit("lsl", makeReg(TR4), makeReg(TR4), makeConst(3))
			e.emit("str", makeReg(TR5), offSet(makeReg(TBP), makeReg(TR4)))
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
	e.src = ".global main\n"
	e.label("main")
	e.emit("mov", makeReg(TMAIN), "lr")
	e.emit("sub", "sp", "sp", makeConst(0x100))
	e.emit("mov", makeReg(TSP), "sp")
	e.emit("sub", "sp", "sp", makeConst(0x10000))
	e.emit("mov", makeReg(TBP), "sp")
	for _, s := range f.SList {
		e.emitStmt(s)
	}
	e.emit("mov", makeReg(0), "xzr")
	e.emit("ret")
	for _, s := range f.FList {
		e.emitFunc(s)
	}
}
