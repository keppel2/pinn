package main

import "fmt"

const RP = "w"
const TR = RP + "29"
const TR2 = RP + "28"
const TRL = "x27"
const TMAIN = "x26"
const BP = ".br"
const FP = ".f"

type emitter struct {
	src     string
	rMap    map[string]int
	creg    int
	cbranch int
	lstack  []int
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
	e.src += "\n"
}

func (e *emitter) init() {
	e.rMap = make(map[string]int)
	e.creg = 8
	e.cbranch = 1
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

	panic(fmt.Sprintln(msg, e.rMap, e.creg, e.src))
}

func (e *emitter) regOrImm(ex Expr) string {
	rt := ""
	switch t := ex.(type) {
	case *NumberExpr:
		rt = "#" + t.Il.Value
	case *VarExpr:
		i, ok := e.rMap[t.Wl.Value]
		if !ok {
			e.err("")
		}
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
		e.emit("mov", TR, e.regOrImm(v))
		return TR
	}
	return e.regOrImm(ex)
}

func (e *emitter) binaryExpr(dest int, be *BinaryExpr) {
	if be.op == "==" || be.op == "!=" || be.op == "<" || be.op == ">" {
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
		e.emit("udiv", TR2, makeReg(dest), rh)
		e.emit("msub", makeReg(dest), TR2, rh, makeReg(dest))
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
	for k, v := range ce.Params {
		e.assignToReg(k+1, v)
	}
	e.emit("mov", TRL, "lr")
	e.emit("bl", FP+ID)
	e.emit("mov", "lr", TRL)

}

func (e *emitter) emitStmt(s Stmt) {
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
			e.emit("mov", "lr", TMAIN)
			e.emit("ret")
			e.makeLabel(lab)
		} else if ID == "bad" {
			e.emit("mov", RP+"0", "1")
			e.emit("mov", "lr", TMAIN)
			e.emit("ret")

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
		e.assignToReg(0, t.E)
	case *AssignStmt:
		lhi := e.rMap[t.LHSa[0].(*VarExpr).Wl.Value]
		e.assignToReg(lhi, t.RHSa[0])
	case *VarStmt:
		s := t.List[0].Value
		e.rMap[s] = e.creg
		e.creg++
	}

}

func (e *emitter) emitF(f *File) {
	e.src = `
.global main
main:
`
	e.emit("mov", TMAIN, "lr")
	for _, s := range f.SList {
		e.emitStmt(s)
	}
	e.emit("mov", RP+"0", RP+"zr")
	e.emit("ret")
	for _, s := range f.FList {
		e.emitFunc(s)
	}
}
