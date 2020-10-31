package main

import "fmt"

type emitter struct {
	rMap map[string]int
	creg int
}

func (e *emitter) init() {
	e.rMap = make(map[string]int)
	e.creg = 1
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
		i, ok := e.rMap[t.Wl.Value]; if !ok {
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


func (e *emitter) binaryExpr(dest string, be BinaryExpr) string {
	rt := ""
	switch t := be.LHS.(type) {
	case NumberExpr, VarExpr:
		rt += "  mov " + dest + ", " + e.regOrImm(t) + "\n"
	case BinaryExpr:
		rt += e.binaryExpr(dest, t)
	}
	op := ""
	switch be.op {
	case "+":
		op = "add"
	case "-":
		op = "sub"
	}

	rt += "  " + op + " " + dest + "," + dest + "," + e.regOrImm(be.RHS) + "\n"
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
	case ReturnStmt:
		rt += "  mov w0, "
		rt += e.regOrImm(t.E) + "\n"
		rt += "  ret\n"
	case AssignStmt:
		lh := e.operand(t.LHSa[0].(VarExpr))
		rh := ""
		switch t2 := t.RHSa[0].(type) {
		case NumberExpr, VarExpr:
			rh += e.operand(t2)
			rt += "  mov " + lh + ", " + rh + "\n"
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
	//	rt += "ret\n"
	return rt
}
