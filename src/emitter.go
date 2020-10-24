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

func (e *emitter) emitExpr(ex Expr) string {
	rt := ""
	switch t := ex.(type) {
	case NumberExpr, VarExpr:
		rt = e.regOrImm(t)
	case BinaryExpr:
		e.err("")
	}

	return rt
}

func (e *emitter) emitStmt(s Stmt) string {
	rt := ""
	switch t := s.(type) {
	case ReturnStmt:
		/*
		rt += "  mov w0, "
		rt += e.emitExpr(t.E) + "\n"
		rt += "  ret\n"
		*/
	case AssignStmt:
		lh := e.emitExpr(t.LHSa[0].(VarExpr))
		rh := ""
		switch t2 := t.RHSa[0].(type) {
		case NumberExpr, VarExpr:
			rh += e.emitExpr(t2)
			rt += "  mov " + lh + ", " + rh + "\n"
			return rt
		case BinaryExpr:
			rt = "  add " + lh + ", " + e.emitExpr(t2.LHS) + ", " + e.emitExpr(t2.RHS) + "\n"
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
