package main

import "math/rand"
import "fmt"
import "reflect"
import "strconv"

const RP = "x"

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

const (
	LR reg = 30 - iota
	TR1
	TR2
	TR3
	TRV
	TMAIN
	TBP
	TSP
	RMAX
)

const (
	SP reg = LR + 1 + iota
	XZR
)

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

const BP = ".br"
const FP = ".f"

var RB reg = R8 + 1

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
	r   reg
}

type emitter struct {
	src     string
	rMap    map[string]*mloc
	rAlloc  [LR + 1]string
	cbranch branch
	moff    int
	lstack  [][2]branch
	lst     Node
	st      Node
}

func (m *mloc) String() string {
	return fmt.Sprintf("%#v", m)
}

func (m *mloc) init(mlt int) {
	m.Mlt = mlt
}

func (e *emitter) findReg() reg {
	for k, v := range e.rAlloc[RB:] {
		if v == "" {
			return reg(k) + RB
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
		ml.init(MLheap)
		ml.i = e.moff
		ml.len = e.atoi(t.Len.(*NumberExpr).Il.Value)
		e.moff += ml.len
		e.rMap[s] = ml

	}
}

func (e *emitter) push(r reg) {
	e.emit("str", makeReg(r), "["+makeReg(TSP), "-8]!")
}

func (e *emitter) pop(r reg) {
	e.emit("ldr", makeReg(r), "["+makeReg(TSP)+"]", "8")
}
func (e *emitter) popP() {
	for i := RB - 1; i >= 0; i-- {
		e.pop(i)
	}
}

func (e *emitter) pushP() {
	for i := R0; i <= R8; i++ {
		e.push(i)
	}
}

func offSet(a, b string) string {
	return fmt.Sprintf("[%v%v%v]", a, OS, b)
}

func (e *emitter) freeReg() reg {
	k := rand.Intn(int(RMAX + 1 - RB))
	k += int(RB)
	s := e.rAlloc[k]
	e.rAlloc[k] = ""
	ml, ok := e.rMap[s]
	if !ok {
		e.err("")
	}
	ml.Mlt = MLheap
	ml.i = e.moff
	e.moff++
	e.emit("str", makeReg(reg(k)), offSet(makeReg(TBP), makeConst(moffOff(ml.i))))
	return reg(k)
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

func (e *emitter) emitR(i string, ops ...regOrConst) {
	sa := []string{}
	for _, s := range ops {

		sa = append(sa, makeRC(s))
	}
	e.emit(i, sa...)
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

func (e *emitter) clab() branch {
	rt := e.cbranch
	e.cbranch++
	return rt
}

func makeReg(i reg) string {

	if i == LR {
		return "lr"
	}
	if i == SP {
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

func makeConst(i int) string {
	return fmt.Sprintf("#%v", i)
}

func makeBranch(i branch) string {
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
	e.label("print")
	//	e.pushP()
	//	e.mov(TR3, a)

	e.emitR("sub", TSP, TSP, 24)
	e.mov(TR2, 0)
	for i := 0; i < 16; i++ {
		e.emitR("and", TR1, R1, 0xf)
		e.emitR("cmp", TR1, 10)
		lab := e.clab()
		e.br(".lt", lab)
		e.emitR("add", TR1, TR1, int('a'-':'))
		e.makeLabel(lab)
		e.emitR("lsr", R1, R1, 4)
		e.emitR("add", TR1, TR1, int('0'))
		e.emitR("lsl", TR2, TR2, 8)
		e.emitR("add", TR2, TR2, TR1)
		if i == 7 {
			e.emit("str", makeReg(TR2), offSet(makeReg(TSP), makeConst(16)))
			e.mov(TR2, 0)
		}
	}
	e.emit("str", makeReg(TR2), offSet(makeReg(TSP), makeConst(8)))
	e.mov(TR2, int(','))
	e.emit("str", makeReg(TR2), fmt.Sprintf("[%v]", makeReg(TSP)))

	e.mov(R0, 1)
	e.mov(R1, TSP)
	e.mov(R2, 24)
	e.mov(R8, 64)
	e.emitR("svc", 0)
	e.emitR("add", TSP, TSP, 24)
	e.emit("ret")
}

func (e *emitter) fillReg(s string) reg {
	ml, ok := e.rMap[s]
	if ok && ml.Mlt == MLreg {
		return ml.r
	}
	k := e.findReg()
	if !ok {
		ml = new(mloc)
		ml.init(MLreg)
		ml.r = k
		e.rMap[s] = ml
	} else {
		e.emit("ldr", makeReg(k), offSet(makeReg(TBP), makeConst(moffOff(ml.i))))
		ml.Mlt = MLreg
		ml.r = k
	}
	e.rAlloc[ml.r] = s
	return k
}

func makeRC(a regOrConst) string {
	if a2, ok := a.(reg); ok {
		return makeReg(a2)
	}
	return makeConst(a.(int))
}

func (e *emitter) br(t string, b branchi) {
	e.emit("b"+t, makeBranch(b.(branch)))
}

func (e *emitter) mov(a regi, b regOrConst) {
	e.emitR("mov", a, b)
	/*
		a2, ok := a.(reg)
		if !ok {
			e.err("")
		}
	*/

	//	sb := makeRC(b)
	//	e.emit("mov", makeReg(a.(reg)), sb)//.(reg)), sb)
}

func (e *emitter) regLoad(ex Expr) reg {
	var rt reg
	switch t := ex.(type) {
	case *NumberExpr:
		e.mov(TR1, e.atoi(t.Il.Value))
		rt = TR1
	case *VarExpr:
		i := e.fillReg(t.Wl.Value)
		rt = i
	default:
		e.err("")
	}
	return rt

}

func (e *emitter) doOp(dest, a, b reg, op string) {
	mn := ""
	switch op {
	case "+":
		mn = "add"
	case "-":
		mn = "sub"
	case "*":
		mn = "mul"
	case "/":
		mn = "udiv"
	}
	if mn != "" {
		e.emitR(mn, dest, a, b)
		return
	}
	switch op {
	case "%":
		e.emitR("udiv", dest, a, b)
		e.emitR("msub", dest, dest, b, a)
		return
	default:
		e.err(op)
	}

}

func (e *emitter) binaryExpr(dest reg, be *BinaryExpr) {
	e.lst = e.st
	e.st = be
	defer func() { e.st = e.lst }()
	if be.op == "==" || be.op == "!=" || be.op == "<" || be.op == "<=" || be.op == ">" || be.op == ">=" {
		e.assignToReg(TR3, be.LHS)
		e.assignToReg(TR2, be.RHS)
		e.emitR("cmp", TR3, TR2)
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
		}
		e.br("."+bi, branch(dest))
		return
	}
	switch t := be.LHS.(type) {
	case *NumberExpr, *VarExpr:
		e.assignToReg(TR2, t)
	case *BinaryExpr:
		e.binaryExpr(TR2, t)
	case *CallExpr:
		e.emitCall(t)
		e.mov(TR2, R0)
	}
	//	op := ""
	if t, ok := be.RHS.(*CallExpr); ok {
		e.emitCall(t)
		e.mov(TR3, 0)
	} else {
		e.assignToReg(TR3, be.RHS)
	}
	e.doOp(dest, TR2, TR3, be.op)
}

func (e *emitter) emitFunc(f *FuncDecl) {
	e.src += FP + f.Wl.Value + ":\n"
	reg := R1
	for _, vd := range f.PList {
		for _, vd2 := range vd.List {
			ml := new(mloc)
			ml.init(MLreg)
			ml.r = reg
			e.rMap[vd2.Value] = ml
			e.rAlloc[reg] = vd2.Value
			reg++
		}
	}
	e.emitStmt(f.B)
	e.emit("ret")
}

func (e *emitter) assignToReg(r reg, ex Expr) {
	switch t2 := ex.(type) {
	case *NumberExpr, *VarExpr:
		rh := e.regLoad(t2)
		e.mov(r, rh)
	case *BinaryExpr:
		e.binaryExpr(r, t2)
	case *CallExpr:
		e.emitCall(t2)
		e.mov(r, R0)
	case *IndexExpr:
		v := t2.X.(*VarExpr).Wl.Value
		ml := e.rMap[v]
		e.assignToReg(TR2, t2.E)
		e.emitR("lsl", TR2, TR2, 3)
		e.emitR("add", TR2, TR2, moffOff(ml.i))

		e.emit("ldr", makeReg(r), offSet(makeReg(TBP), makeReg(TR2)))

	default:
		e.err("")
	}

}

func (e *emitter) emitCall(ce *CallExpr) {
	e.st = ce
	ID := ce.ID.(*VarExpr).Wl.Value
	// e.pushP()
	e.push(TRV)
	for k, v := range ce.Params {
		e.push(1 + reg(k))
		e.assignToReg(reg(k)+1, v)
	}
	e.push(LR)
	e.emit("bl", FP+ID)
	e.pop(LR)
	for k, _ := range ce.Params {
		e.pop(reg(k) + 1)
	}
	e.pop(TRV)
	//  e.popP()

}

func (e *emitter) emitStmt(s Stmt) {
	e.st = s
	switch t := s.(type) {
	case *ExprStmt:
		ce := t.Expr.(*CallExpr)
		ID := ce.ID.(*VarExpr).Wl.Value
		if ID == "assert" {
			e.assignToReg(TR2, ce.Params[0])
			e.assignToReg(TR3, ce.Params[1])
			e.emitR("cmp", TR2, TR3)
			lab := e.clab()
			e.br(".eq", lab)
			ln := e.st.Gpos().Line
			e.mov(R0, ln)
			e.mov(LR, TMAIN)
			e.emit("ret")
			e.makeLabel(lab)
		} else if ID == "bad" {
			e.mov(R0, 1)
			e.mov(LR, TMAIN)
			e.emit("ret")
		} else if ID == "print" {

			e.assignToReg(R1, ce.Params[0])
			e.push(LR)
			e.emit("bl", "print")
			e.pop(LR)
		} else {
			e.emitCall(ce)
		}

	case *BlockStmt:
		for _, s := range t.SList {
			e.emitStmt(s)
		}
	case *ContinueStmt:
		e.br("", e.peekloop()[0])
	case *BreakStmt:
		e.br("", e.peekloop()[1])
	case *LoopStmt:
		lab := e.clab()
		e.makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab, lab2)
		e.emitStmt(t.B)
		e.br("", lab)
		e.makeLabel(lab2)
		e.poploop()
	case *WhileStmt:
		lab := e.clab()
		e.makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab, lab2)
		e.binaryExpr(reg(lab2), t.Cond.(*BinaryExpr))
		e.emitStmt(t.B)
		e.br("", lab)
		e.makeLabel(lab2)

	case *IfStmt:
		lab := e.clab()
		if t.Else == nil {
			e.binaryExpr(reg(lab), t.Cond.(*BinaryExpr))
			e.emitStmt(t.Then)
		} else {
			lab2 := e.clab()
			e.binaryExpr(reg(lab2), t.Cond.(*BinaryExpr))
			e.emitStmt(t.Then)
			e.br("", lab)
			e.makeLabel(lab2)
			e.emitStmt(t.Else)
		}
		e.makeLabel(lab)

	case *ReturnStmt:
		if t.E != nil {
			e.assignToReg(TRV, t.E)
		}
		e.mov(R0, TRV)
		e.emit("ret")
	case *AssignStmt:
		lh := t.LHSa[0]
		switch t2 := lh.(type) {
		case *VarExpr:
			if t.Op == "+=" || t.Op == "-=" || t.Op == "/=" || t.Op == "*=" || t.Op == "%=" {
				lhi := e.fillReg(t2.Wl.Value)
				e.assignToReg(TR2, lh)
				e.assignToReg(TR3, t.RHSa[0])
				e.doOp(lhi, TR2, TR3, t.Op[0:1])

				return
			}
			lhi := e.fillReg(t2.Wl.Value)
			if t.Op == "++" {
				e.mov(TR1, 1)
				e.doOp(lhi, lhi, TR1, "+")
				return
			} else if t.Op == "--" {
				e.mov(TR1, 1)
				e.doOp(lhi, lhi, TR1, "-")
				return
			}
			e.assignToReg(lhi, t.RHSa[0])
		case *IndexExpr:
			if t.Op == "+=" || t.Op == "-=" || t.Op == "/=" || t.Op == "*=" || t.Op == "%=" {
				e.assignToReg(TR3, t2)
				e.mov(TR2, TR3)
				e.assignToReg(TR1, t.RHSa[0])
				e.doOp(TR3, TR2, TR1, t.Op[0:1])
			} else if t.Op == "++" || t.Op == "--" {
				e.assignToReg(TR3, t2)
				if t.Op == "++" {
					e.mov(TR1, 1)
					e.doOp(TR3, TR3, TR1, "+")
				} else {
					e.mov(TR1, 1)
					e.doOp(TR3, TR3, TR1, "-")
				}
			} else {
				e.assignToReg(TR3, t.RHSa[0])
			}

			v := t2.X.(*VarExpr).Wl.Value
			ml := e.rMap[v]
			e.assignToReg(TR2, t2.E)
			e.emitR("lsl", TR2, TR2, 3)
			e.emitR("add", TR2, TR2, moffOff(ml.i))
			e.emit("str", makeReg(TR3), offSet(makeReg(TBP), makeReg(TR2)))
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
		e.br("", lab3)
		e.makeLabel(lab)
		if t.Loop != nil {
			e.emitStmt(t.Loop)
		}
		e.makeLabel(lab3)

		if t.E != nil {
			e.binaryExpr(reg(lab2), t.E.(*BinaryExpr))
		}
		e.emitStmt(t.B)
		e.br("", lab)

		e.makeLabel(lab2)

	default:
		e.err("")

	}

}

func (e *emitter) emitF(f *File) {
	e.src = ".global main\n"
	e.label("main")
	e.mov(TMAIN, LR)
	e.emitR("sub", SP, SP, 0x100)
	e.mov(TSP, SP)
	e.emitR("sub", SP, SP, 0x10000)
	e.mov(TBP, SP)
	for _, s := range f.SList {
		e.emitStmt(s)
	}
	e.mov(R0, XZR)
	e.emit("ret")
	for _, s := range f.FList {
		e.emitFunc(s)
	}
	e.emitPrint()
}
