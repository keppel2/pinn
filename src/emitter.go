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
}

func (e *emitter) init() {
	e.rMap = make(map[string]int)
	e.creg = 1
	e.cbranch = 1
}

func (e emitter) clab() string {
	return fmt.Sprintf("ifb%v", e.cbranch)
}

func (e *emitter) err(msg string) {

	panic(fmt.Sprintln(msg, e.rMap, e.creg))
}

func (e *emitter) regOrImm(ex Expr) string {
	rt := ""
	switch t := ex.(type) {
	case NumberExpr:
		rt = "#" + t.Il.Value
	case VarExpr:
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
	case NumberExpr, VarExpr:
		rt = e.regOrImm(t)
	default:
		e.err("")
	}

	return rt
}

func (e *emitter) moveToTr(ex Expr) (string, string) {
	rt := ""
	if v, ok := ex.(NumberExpr); ok {
		rt += ind + "mov" + AM + TR + OS + e.regOrImm(v) + "\n"
		return rt, TR
	}
	return rt, e.regOrImm(ex)
}

func (e *emitter) binaryExpr(dest string, be BinaryExpr) string {
	rt := ""
	if be.op == "==" || be.op == "!=" || be.op == "<" || be.op == ">" {
		mtr, lh := e.moveToTr(be.LHS)
		rt += mtr
		rh := e.regOrImm(be.RHS)
		rt += ind + "cmp" + AM + lh + OS + rh + "\n"
		lab := e.clab()
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
		rt += ind + "b." + bi + AM + lab + "\n"

		return rt
	}
	switch t := be.LHS.(type) {
	case NumberExpr, VarExpr:
		rt += ind + "mov" + AM + dest + OS + e.regOrImm(t) + "\n"
	case BinaryExpr:
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
		rt += ind + "udiv" + AM + TR2 + OS + dest + OS + rh + "\n"
		rt += ind + "msub" + AM + dest + OS + TR2 + OS + rh + OS + dest + "\n"
		return rt
	}
	rt += ind + op + AM + dest + OS + dest + OS + rh + "\n"
	return rt
}
func (e *emitter) emitExpr(dest string, ex Expr) string {

	rt := ""
	return rt

	//switch t := e.(type) {
	//	}
}

func (e *emitter) emitStmt(s Stmt) string {
	rt := ""
	switch t := s.(type) {
	case BlockStmt:
		for _, s := range t.SList {
			rt += e.emitStmt(s)
		}
	case IfStmt:
		rt += e.binaryExpr("", t.Cond.(BinaryExpr))
		lab := e.clab()
		e.cbranch++
		rt += e.emitStmt(t.Then)
		rt += lab + ":\n"
		e.cbranch--

	case ReturnStmt:
		rt += ind + "mov" + AM + "w0" + OS
		rt += e.regOrImm(t.E) + "\n"
	case AssignStmt:
		lh := e.operand(t.LHSa[0].(VarExpr))
		rh := ""
		switch t2 := t.RHSa[0].(type) {
		case NumberExpr, VarExpr:
			rh += e.operand(t2)
			rt += ind + "mov" + AM + lh + OS + rh + "\n"
			return rt
		case BinaryExpr:

			rt += e.binaryExpr(lh, t2)
			//			rt = "  add " + lh + ", " + e.emitExpr(lh, t2.LHS) + ", " + e.emitExpr(lh, t2.RHS) + "\n"
			return rt
		}
		//rh := e.emitExpr(t.RHSa[0])

	case VarStmt:
		s := t.List[0].Value
		e.rMap[s] = e.creg
		e.creg++
	}
	return rt

}

func (e *emitter) emit(f File) string {
	rt := `
.global main
main:
`
	for _, s := range f.SList {
		rt += e.emitStmt(s)
	}
	rt += ind + "ret\n"
	return rt
}
