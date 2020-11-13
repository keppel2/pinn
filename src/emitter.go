package main

import "math/rand"
import "fmt"
import "reflect"

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
	TMAIN
	TBP
	TSP
	RMAX
)
const BP = ".br"
const FP = ".f"

var RB = 9

type emitter struct {
	src     string
	rMap    map[string]int
	rAlloc  [RMAX + 1]string
	cbranch int
	moff    int
	lstack  []int
	st      Stmt
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
	return (-1 - a) * bw()
}

func (e *emitter) newVar(s string) int {
	i := e.findReg()
	e.rMap[s] = i
	e.rAlloc[i] = s
	return i
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
	e.rMap[s] = e.moff
	e.emit("str", makeReg(k), "["+makeXReg(TBP), fmt.Sprintf("%v]", moffOff(e.moff)))
	e.moff--
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
	e.rMap = make(map[string]int)
	e.cbranch = 1
	e.moff = -1
}

func (e *emitter) varAt(i int) string {
	for k, v := range e.rMap {
		if v == i {
			return k
		}
	}
	return ""
}

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

func (e *emitter) pushloop(s int) {
	e.lstack = append(e.lstack, s)
}

func (e *emitter) poploop() int {
	rt := e.lstack[len(e.lstack)-1]
	e.lstack = e.lstack[0 : len(e.lstack)-1]
	return rt
}

func (e *emitter) peekloop() int {
	return e.lstack[len(e.lstack)-1]
}

func (e *emitter) err(msg string) {

	panic(fmt.Sprintln(msg, e.rMap, e.rAlloc, e.src))
}

func (e *emitter) fillReg(s string, load bool) int {
	moff, ok := e.rMap[s]
	if ok && moff >= 0 {
		return moff
	}

	k := e.findReg()
	e.rMap[s] = k
	e.rAlloc[k] = s
	if ok && load {
		e.emit("ldr", makeReg(k), "["+makeXReg(TBP), fmt.Sprintf("%v]", moffOff(moff)))
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
	if v, ok := ex.(*NumberExpr); ok {
		e.emit("mov", makeReg(TR), e.regOrImm(v))
		return makeReg(TR)
	}
	return e.regOrImm(ex)
}

func (e *emitter) binaryExpr(dest int, be *BinaryExpr) {
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
			e.rMap[vd2.Value] = reg
			e.rAlloc[reg] = vd2.Value
			reg++
		}
	}
	e.emitStmt(f.B)
	e.emit("ret")
}

/*
func (e *emitter) emitExpr(dest string, ex Expr){

	rt := ""
	return rt

	switch t := e.(type) {
		}
}
*/

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

	default:
		e.err("")
	}

}

func (e *emitter) emitCall(ce *CallExpr) {
	ID := ce.ID.(*VarExpr).Wl.Value
	// e.pushP()
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
	//  e.popP()

}

func (e *emitter) emitStmt(s Stmt) {
	e.st = s
	switch t := s.(type) {
	case *ExprStmt:
		ce := t.Expr.(*CallExpr)
		ID := ce.ID.(*VarExpr).Wl.Value
		if ID == "assert" {
			lh := e.moveToTr(ce.Params[0])
			rh := e.regOrImm(ce.Params[1])
			e.emit("cmp", lh, rh)
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
			e.pushP()
			e.emit("mov", makeReg(0), "1")
			e.emit("mov", makeReg(1), "0")
			e.emit("sub", makeReg(TSP), makeReg(TSP), "16")
			for i := 0; i < 16; i++ {
				e.emit("and", makeReg(TR), makeReg(TR2), "0xf")
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
		} else {
			e.emitCall(ce)
		}

	case *BlockStmt:
		for _, s := range t.SList {
			e.emitStmt(s)
		}
	case *ContinueStmt:
		e.emit("b", makeBranch(e.peekloop()-1))
	case *BreakStmt:
		e.emit("b", makeBranch(e.peekloop()))
	case *LoopStmt:
		lab := e.clab()
		e.makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab2)
		e.emitStmt(t.B)
		e.emit("b", makeBranch(lab))
		e.makeLabel(lab2)
		e.poploop()
	case *WhileStmt:
		lab := e.clab()
		e.makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab2)
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
			e.assignToReg(0, t.E)
		}
		e.emit("ret")
	case *AssignStmt:
		lhi := e.fillReg(t.LHSa[0].(*VarExpr).Wl.Value, false)
		e.assignToReg(lhi, t.RHSa[0])
	case *VarStmt:
		//		s := t.List[0].Value
		//		e.newVar(s)
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
