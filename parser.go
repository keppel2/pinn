package main

import (
	"fmt"
	"io"
	//	"os"
)

type parser struct {
	scan
}

func (p *parser) err(msg string) {
	panic(fmt.Sprintln(msg, p.p, p.tok, p.lit))
}

func (p *parser) init(r io.Reader) {
	p.scan.init(
		r)
}

func (p *parser) got(tok string) bool {
	if p.tok == tok {
		p.next()
		return true
	}
	return false
}

func (p *parser) want(tok string) {
	if !p.got(tok) {
		p.err("expecting " + tok)
	}
}

func contains(s []string, t string) bool {
	for _, v := range s {
		if v == t {
			return true
		}
	}
	return false
}

func (p *parser) unaryExpr() Expr {
	switch p.tok {
	case "-", "+", "!":
		ue := UnaryExpr{}
		ue.op = p.tok
		p.next()
		ue.E = p.unaryExpr()
		return ue

	case "literal", "name", "(", "[", "...":
		return p.primaryExpr()
	}
	p.err("")
	return nil

	//  return p.primaryExpr()
}

func (p *parser) primaryExpr() Expr {
	x := p.operand()
	for {
		switch p.tok {
		case "(":
			x = p.callExpr(x)
		case "[":
			x = p.indexExpr(x)
		default:
			return x
		}
	}
}

func (p *parser) operand() Expr {
	dots := false
	switch p.tok {
	case "name":
		return p.varExpr()
	case "literal":
		return p.numberExpr()
	case "(":
		p.want("(")
		e := p.uexpr()
		p.want(")")
		return e
	case "...":
		p.next()
		dots = true
		fallthrough
	case "[":
		p.want("[")
		return p.arrayExpr(dots)
	}
	p.err("")
	return nil
}

func (p *parser) arrayExpr(d bool) ArrayExpr {
	rt := ArrayExpr{}
	rt.EL = p.exprList()
	p.want("]")
	rt.Dots = d
	return rt
}

func (p *parser) fileA() File {
	f := File{}
	f.Position = p.p
	p.next()
	f.SList = p.stmtList()
	fmt.Println(f.SList)
	visitFile(f)

	return f
}

func (p *parser) loopStmt() LoopStmt {
	p.want("loop")
	rt := LoopStmt{}
	rt.B = p.blockStmt().(BlockStmt)
	return rt
}
func (p *parser) whileStmt() WhileStmt {
	p.want("while")
	rt := WhileStmt{}
	rt.Cond = p.uexpr()
	rt.B = p.blockStmt().(BlockStmt)
	return rt
}

func (p *parser) ifStmt() IfStmt {
	p.want("if")
	rt := IfStmt{}
	rt.Cond = p.uexpr()
	rt.Then = p.stmt()
	if p.got("else") {
		rt.Else = p.stmt()
	}
	return rt
}

func (p *parser) field() Field {
	n := Field{}
	n.List = append(n.List, p.wLit())
	for p.got(",") {
		n.List = append(n.List, p.wLit())
	}
	if p.got("...") {
		n.Dots = true
	}
	n.Kind = p.kind()
	return n
}

func (p *parser) varStmt() VarStmt {
	p.want("var")
	ds := VarStmt{}
	ds.Position = p.p
	ds.List = append(ds.List, p.wLit())
	for p.got(",") {
		ds.List = append(ds.List, p.wLit())
	}

	ds.Kind = p.kind()

	p.want(";")
	return ds
}

func (p *parser) typeStmt() TypeStmt {
	ds := TypeStmt{}
	ds.Position = p.p
	p.want("type")

	ds.Wl = p.wLit()
	ds.Kind = p.kind()

	p.want(";")
	return ds

}

func (p *parser) funcStmt() FuncStmt {
	rt := FuncStmt{}
	p.want("func")
	rt.Wl = p.wLit()
	p.want("(")
	if !p.got(")") {
		vd := p.field()
		rt.PList = append(rt.PList, vd)
		for p.got(",") {
			vd = p.field()
			rt.PList = append(rt.PList, vd)
		}
		p.want(")")
	}
	if p.tok != "{" {
		rt.K = p.kind()
	}
	rt.B = p.blockStmt().(BlockStmt)

	return rt
}

func (p *parser) returnStmt() ReturnStmt {
	rt := ReturnStmt{}
	p.want("return")
	if !p.got(";") {
		rt.E = p.uexpr()
		p.want(";")
	}
	return rt
}

func (p *parser) stmt() Stmt {
	var rt Stmt
	switch p.tok {
	case "return":
		rt = p.returnStmt()
	case "var":
		rt = p.varStmt()
	case "type":
		rt = p.typeStmt()
	case "func":
		rt = p.funcStmt()
	case "if":
		rt = p.ifStmt()
	case "while":
		rt = p.whileStmt()
	case "loop":
		rt = p.loopStmt()

	case "literal": //, "-", "+":
		lhs := p.unaryExpr()
		rt = p.exprStmt(lhs)
	case "name":
		lhsa := p.exprList()
		if p.tok == "=" || p.tok == ":=" || p.tok == "+=" || p.tok == "-=" || p.tok == "*=" || p.tok == "/=" || p.tok == "%=" || p.tok == "++" || p.tok == "--" {
			rt = p.assignStmt(lhsa)
		} else {
			rt = p.exprStmt(lhsa[0])
		}
	case "{":
		rt = p.blockStmt()
	case ";":
		p.next()
	default:
		p.err("")
	}
	return rt

}

func (p *parser) stmtList() []Stmt {
	rt := make([]Stmt, 0)
	for p.tok != "EOF" && p.tok != "}" {
		rt = append(rt, p.stmt())
	}
	return rt
}

func (p *parser) exprList() []Expr {
	rt := make([]Expr, 0)
	rt = append(rt, p.uexpr())
	for p.tok == "," {
		p.next()
		rt = append(rt, p.uexpr())
	}
	return rt
}

func (p *parser) blockStmt() Stmt {
	rt := BlockStmt{}
	p.want("{")
	rt.SList = p.stmtList()
	p.want("}")
	return rt
}

func (p *parser) sKind() SKind {
	rt := SKind{}
	rt.Wl = p.wLit()
	return rt
}

func (p *parser) kind() Kind {
	switch p.tok {
	case "[":
		p.want("[")
		if p.got("]") {
			rt := SlKind{}
			rt.K = p.kind()
			return rt
		}
		if p.got("map") {
			p.want("]")
			rt := MKind{}
			rt.K = p.kind()
			return rt
		}
		rt := ArKind{}
		rt.Len = p.uexpr()
		p.want("]")
		rt.K = p.kind()
		return rt

	case "name":
		return p.sKind()
	}
	p.err("")
	return nil
}

func (p *parser) assignStmt(LHSa []Expr) AssignStmt {

	rt := AssignStmt{}
	rt.Position = p.p
	rt.Op = p.tok
	p.next()
	rt.LHSa = LHSa
	if rt.Op == "++" || rt.Op == "--" {
	} else {
		rt.RHSa = p.exprList()
	}
	p.want(";")
	return rt

}

func (p *parser) exprStmt(LHS Expr) ExprStmt {
	es := ExprStmt{}
	es.Position = p.p
	rt := p.pexpr(LHS)
	es.Expr = rt
	p.want(";")
	return es
}

func (p *parser) pexpr(LHS Expr) Expr {
	if p.tok == "+" || p.tok == "-" || p.tok == "/" || p.tok == "*" || p.tok == "%" || p.tok == "<" || p.tok == "<=" || p.tok == ">=" || p.tok == ">" || p.tok == "==" || p.tok == "!=" || p.tok == "&&" || p.tok == "||" || p.tok == ">>" || p.tok == "<<" || p.tok == "&" || p.tok == "|" || p.tok == "^" {
		return p.binaryExpr(LHS)
	}
	if p.tok == "?" {
		return p.trinaryExpr(LHS)
	}
	return LHS
}

func (p *parser) uexpr() Expr {
	return p.pexpr(p.unaryExpr())
}

func (p *parser) trinaryExpr(lhs Expr) Expr {
	rt := TrinaryExpr{}
	rt.LHS = lhs
	p.want("?")
	rt.MS = p.uexpr()
	p.want(":")
	rt.RHS = p.uexpr()
	return lhs
}

func (p *parser) binaryExpr(lhs Expr) Expr {
	op := p.lit
	p.next()
	rt := BinaryExpr{}
	rt.LHS = lhs
	rt.op = op
	rt.RHS = p.uexpr()

	return rt
}

func (p *parser) indexExpr(lhs Expr) Expr {
	p.want("[")
	rt := IndexExpr{}
	rt.X = lhs

	if p.tok != (":") && p.tok != ("@") {
		rt.Start = p.uexpr()
		if p.got("]") {
			return rt
		}
	}
	if p.tok == "@" {
		rt.Inc = true
	}
	p.next()
	if p.got("]") {
		return rt
	}
	rt.End = p.uexpr()
	p.want("]")
	return rt
}

func (p *parser) callExpr(lhs Expr) Expr {
	p.want("(")
	rt := CallExpr{}
	rt.ID = lhs
	if p.got(")") {
		return rt
	}
	e := p.uexpr()
	rt.Params = append(rt.Params, e)
	for p.got(",") {
		rt.Params = append(rt.Params, p.uexpr())
	}
	p.want(")")
	return rt
}

func (p *parser) iLit() ILit {
	il := ILit{}
	il.Position = p.p
	if p.tok != "literal" {
		p.err("")
	}
	il.Value = p.lit
	p.next()
	return il

}
func (p *parser) wLit() WLit {
	wl := WLit{}
	wl.Position = p.p
	if p.tok != "name" {
		p.err("")
	}
	wl.Value = p.lit
	p.next()
	return wl
}

func (p *parser) varExpr() Expr {
	rt := VarExpr{}
	rt.Position = p.p
	rt.Wl = p.wLit()
	return rt
}

func (p *parser) numberExpr() Expr {
	ne := NumberExpr{}
	ne.Position = p.p

	ne.Il = p.iLit()
	return ne

}
