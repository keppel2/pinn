package main

import "fmt"

const ind = "  "
const OS = ", "
const AM = " "
const TR = "w29"
const TR2 = "w28"

type emitter struct {
	rMap    map[string]int
	creg    int
	cbranch int
	lstack  []int
}

func emit(i string, ops ...string) string {
	rt := ""
	rt += ind + i + AM
	if ops != nil {
		rt += ops[0]
		for _, s := range ops[1:] {
			rt += OS + s
		}
	}
	rt += "\n"
	return rt
}

func (e *emitter) init() {
	e.rMap = make(map[string]int)
	e.creg = 1
	e.cbranch = 1
}

func (e *emitter) clab() int {
	rt := e.cbranch
	e.cbranch++
	return rt
}

func makeBranch(i int) string {
	return fmt.Sprintf("gb%v", i)
}

func makeLabel(i int) string {
	return fmt.Sprintf("gb%v:\n", i)
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

	panic(fmt.Sprintln(msg, e.rMap, e.creg))
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
		rt = "w" + fmt.Sprint(i)
	default:
		e.err("")
	}
	return rt

}

func (e *emitter) operand(ex Expr) string {
	rt := ""
	switch t := ex.(type) {
	case *NumberExpr, *VarExpr:
		rt = e.regOrImm(t)
	default:
		e.err("")
	}

	return rt
}

func (e *emitter) moveToTr(ex Expr) (string, string) {
	rt := ""
	if v, ok := ex.(*NumberExpr); ok {
		rt += emit("mov", TR, e.regOrImm(v))
		return rt, TR
	}
	return rt, e.regOrImm(ex)
}

func (e *emitter) binaryExpr(dest string, be *BinaryExpr) string {
	rt := ""
	if be.op == "==" || be.op == "!=" || be.op == "<" || be.op == ">" {
		mtr, lh := e.moveToTr(be.LHS)
		rt += mtr
		rh := e.regOrImm(be.RHS)
		rt += emit("cmp", lh, rh)
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
		rt += emit("b."+bi, dest)

		return rt
	}
	switch t := be.LHS.(type) {
	case *NumberExpr, *VarExpr:
		rt += emit("mov", dest, e.regOrImm(t))
	case *BinaryExpr:
		rt += e.binaryExpr(dest, t)
	}
	op := ""
	rh := ""
	switch be.op {
	case "+":
		op = "add"
		fallthrough
	case "-":
		if op == "" {
			op = "sub"
		}
		rh = e.regOrImm(be.RHS)
	case "*", "/":
		if be.op == "*" {
			op = "mul"
		} else {
			op = "udiv"
		}
		mtr := ""
		mtr, rh = e.moveToTr(be.RHS)
		rt += mtr
	case "%":
		mtr := ""
		mtr, rh = e.moveToTr(be.RHS)
		rt += mtr
		rt += emit("udiv", TR2, dest, rh)
		rt += emit("msub", dest, TR2, rh, dest)
		return rt
	}
	rt += emit(op, dest, dest, rh)
	return rt
}

func (e *emitter) emitFunc(f *FuncDecl) string {
	rt := ""
	rt += f.Wl.Value + ":\n"
	rt += e.emitStmt(f.B)
	rt += emit("ret")
	return rt
}

/*
func (e *emitter) emitExpr(dest string, ex Expr) string {

	rt := ""
	return rt

	switch t := e.(type) {
		}
}
*/

func (e *emitter) emitStmt(s Stmt) string {
	rt := ""
	switch t := s.(type) {
	case *ExprStmt:
		ce := t.Expr.(*CallExpr)
		if ce.ID.(*VarExpr).Wl.Value == "assert" {
			mtr, lh := e.moveToTr(ce.Params[0])
			rt += mtr
			rh := e.regOrImm(ce.Params[1])
			rt += emit("cmp", lh, rh)
			lab := e.clab()
			rt += emit("b.eq", makeBranch(lab))
			rt += emit("mov", "w0", "1")
			rt += ind + "ret" + "\n"
			rt += makeLabel(lab)

		} else {
			rt += emit("mov", "x23", "lr")
			rt += emit("bl", ce.ID.(*VarExpr).Wl.Value)
			rt += emit("mov", "lr", "x23")
		}

	case *BlockStmt:
		for _, s := range t.SList {
			rt += e.emitStmt(s)
		}
	case *ContinueStmt:
		rt += emit("b", makeBranch(e.peekloop()-1))
	case *BreakStmt:
		rt += emit("b", makeBranch(e.peekloop()))
	case *LoopStmt:
		lab := e.clab()
		rt += makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab2)
		rt += e.emitStmt(t.B)
		rt += emit("b", makeBranch(lab))
		rt += makeLabel(lab2)
		e.poploop()
	case *WhileStmt:
		lab := e.clab()
		rt += makeLabel(lab)
		lab2 := e.clab()
		e.pushloop(lab2)
		rt += e.binaryExpr(makeBranch(lab2), t.Cond.(*BinaryExpr))
		rt += e.emitStmt(t.B)
		rt += emit("b", makeBranch(lab))
		rt += makeLabel(lab2)

	case *IfStmt:
		lab := e.clab()
		if t.Else == nil {
			rt += e.binaryExpr(makeBranch(lab), t.Cond.(*BinaryExpr))
			rt += e.emitStmt(t.Then)
		} else {
			lab2 := e.clab()
			rt += e.binaryExpr(makeBranch(lab2), t.Cond.(*BinaryExpr))
			rt += e.emitStmt(t.Then)
			rt += emit("b", makeBranch(lab))
			rt += makeLabel(lab2)
			rt += e.emitStmt(t.Else)
		}
		rt += makeLabel(lab)

	case *ReturnStmt:
		rt += emit("mov", "w0", e.regOrImm(t.E))
	case *AssignStmt:
		lh := e.operand(t.LHSa[0].(*VarExpr))
		rh := ""
		switch t2 := t.RHSa[0].(type) {
		case *NumberExpr, *VarExpr:
			rh += e.operand(t2)
			rt += emit("mov", lh, rh)

			return rt
		case *BinaryExpr:

			rt += e.binaryExpr(lh, t2)
			return rt
		}

	case *VarStmt:
		s := t.List[0].Value
		e.rMap[s] = e.creg
		e.creg++
	}
	return rt

}

func (e *emitter) emit(f *File) string {
	rt := `
.global main
main:
`
	for _, s := range f.SList {
		rt += e.emitStmt(s)
	}
	rt += ind + "ret\n"
	for _, s := range f.FList {
		rt += e.emitFunc(s)
	}
	return rt
}
