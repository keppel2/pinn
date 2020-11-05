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
	case "-", "+", "!", "@", "#", "range":
		ue := new(UnaryExpr)
		ue.Init(p.p)
		ue.op = p.tok
		p.next()
		if p.tok == "]" {
			if ue.op != "@" && ue.op != "#" {
				p.err("")
			}
			return ue
		}
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

func (p *parser) arrayExpr(d bool) *ArrayExpr {
	rt := ArrayExpr{}
	rt.EL = p.exprList()
	p.want("]")
	rt.Dots = d
	return &rt
}

func (p *parser) fileA() File {
	f := File{}
	f.Position = p.p
	p.next()
	f.SList = p.stmtList()

	return f
}

func (p *parser) loopStmt() *LoopStmt {
	p.want("loop")
	rt := new(LoopStmt)
	rt.Init(p.p)
	rt.B = p.blockStmt()
	return rt
}
func (p *parser) whileStmt() *WhileStmt {
	p.want("while")
	rt := new(WhileStmt)
	rt.Init(p.p)
	rt.Cond = p.uexpr()
	rt.B = p.blockStmt()
	return rt
}

func (p *parser) ifStmt() *IfStmt {
	p.want("if")
	rt := new(IfStmt)
	rt.Init(p.p)
	rt.Cond = p.uexpr()
	rt.Then = p.stmt()
	if p.got("else") {
		rt.Else = p.stmt()
	}
	return rt
}

func (p *parser) field() *Field {
	n := new(Field)
	n.node.Init(p.p)
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

func (p *parser) varStmt() *VarStmt {
	p.want("var")
	ds := new(VarStmt)
	ds.Init(p.p)
	ds.Position = p.p
	ds.List = append(ds.List, p.wLit())
	for p.got(",") {
		ds.List = append(ds.List, p.wLit())
	}

	ds.Kind = p.kind()

	p.want(";")
	return ds
}

func (p *parser) typeStmt() *TypeStmt {
	ds := new(TypeStmt)
	ds.Init(p.p)
	ds.Position = p.p
	p.want("type")

	ds.Wl = p.wLit()
	ds.Kind = p.kind()

	p.want(";")
	return ds

}

func (p *parser) funcStmt() *FuncStmt {
	rt := new(FuncStmt)
	rt.Init(p.p)
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
	rt.B = p.blockStmt()

	return rt
}

func (p *parser) returnStmt() *ReturnStmt {
	rt := new(ReturnStmt)
	rt.Init(p.p)
	p.want("return")
	if !p.got(";") {
		rt.E = p.uexpr()
		p.want(";")
	}
	return rt
}

func (p *parser) forrStmt() *ForrStmt {
	rt := new(ForrStmt)
	rt.Init(p.p)
	p.want("forr")
	rt.LH = p.exprList()

	if p.tok != "=" && p.tok != ":=" {
		p.err("")
	}
	rt.Op = p.tok
	p.next()
	rt.RH = p.uexpr()
	rt.B = p.blockStmt()
	return rt
}

func (p *parser) forStmt() *ForStmt {
	rt := new(ForStmt)
	rt.Init(p.p)
	p.want("for")
	rt.Inits = p.stmt()
	rt.E = p.uexpr()
	p.want(";")
	rt.Loop = p.stmt()
	rt.B = p.blockStmt()
	return rt
}

func (p *parser) assignOrExprStmt() Stmt {
	lhsa := p.exprList()
	var rt Stmt
	if p.tok == "=" || p.tok == ":=" || p.tok == "+=" || p.tok == "-=" || p.tok == "*=" || p.tok == "/=" || p.tok == "%=" || p.tok == "++" || p.tok == "--" {
		rt = p.assignStmt(lhsa)
	} else {
		if len(lhsa) != 1 {
			p.err("")
		}
		rt = p.exprStmt(lhsa[0])
	}
	return rt
}

func (p *parser) stmt() Stmt {
	var rt Stmt
	switch p.tok {
	case "for":
		rt = p.forStmt()
	case "forr":
		rt = p.forrStmt()
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

	case "literal", "name": //, "-", "+":
		rt = p.assignOrExprStmt()
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

func (p *parser) blockStmt() *BlockStmt {
	rt := new(BlockStmt)
	rt.Init(p.p)
	p.want("{")
	rt.SList = p.stmtList()
	p.want("}")
	return rt
}

func (p *parser) sKind() *SKind {
	rt := new(SKind)
	rt.Init(p.p)
	rt.Wl = p.wLit()
	return rt
}

func (p *parser) kind() Kind {
	switch p.tok {
	case "[":
		p.want("[")
		if p.got("]") {
			rt := new(SlKind)
			rt.Init(p.p)
			rt.K = p.kind()
			return rt
		}
		if p.got("map") {
			p.want("]")
			rt := new(MKind)
			rt.Init(p.p)
			rt.K = p.kind()
			return rt
		}
		rt := new(ArKind)
		rt.Init(p.p)
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

func (p *parser) assignStmt(LHSa []Expr) *AssignStmt {

	rt := new(AssignStmt)
	rt.Init(p.p)
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
	es.Expr = LHS
	p.want(";")
	return es
}

func (p *parser) pexpr(prec int) Expr {
	//	if p.tok == "+" || p.tok == "-" || p.tok == "/" || p.tok == "*" || p.tok == "%" || p.tok == "<" || p.tok == "<=" || p.tok == ">=" || p.tok == ">" || p.tok == "==" || p.tok == "!=" || p.tok == "&&" || p.tok == "||" || p.tok == ">>" || p.tok == "<<" || p.tok == "&" || p.tok == "|" || p.tok == "^" {
	rt := p.unaryExpr()
	//	fmt.Println(prec, p.tok, p.lit, tokenMap[p.tok])

	for tokenMap[p.tok] > prec {
		//		fmt.Println(p.tok, "in")

		if p.tok == "?" {
			return p.trinaryExpr(rt)
		}

		t := new(BinaryExpr)
		t.Init(p.p)
		t.op = p.tok
		t.LHS = rt
		prec := tokenMap[p.tok]
		p.next()
		if p.tok == "]" {
			return rt
		}
		t.RHS = p.pexpr(prec)
		rt = t
	}
	return rt

	/*



		}
		if p.tok == "?" {
			return p.trinaryExpr(LHS)
		}
		return LHS
	*/
}

func (p *parser) uexpr() Expr {
	return p.pexpr(0)
}

func (p *parser) trinaryExpr(lhs Expr) Expr {
	rt := new(TrinaryExpr)
	rt.Init(p.p)
	rt.LHS = lhs
	p.want("?")
	rt.MS = p.uexpr()
	p.want(":")
	rt.RHS = p.uexpr()
	return rt
}

/*
func (p *parser) binaryExpr(lhs Expr) Expr {
	op := p.tok
	p.next()
	rt := BinaryExpr{}
	rt.LHS = lhs
	rt.op = op
	rt.RHS = p.uexpr()

	return rt
}
*/

func (p *parser) indexExpr(lhs Expr) Expr {
	p.want("[")
	rt := new(IndexExpr)
	rt.Init(p.p)
	rt.X = lhs
	if p.got("]") {
		return rt
	}
	rt.E = p.uexpr()
	p.want("]")
	return rt
	/*

		if p.tok != ("#") && p.tok != ("@") {
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
	*/
}

func (p *parser) callExpr(lhs Expr) Expr {
	p.want("(")
	rt := new(CallExpr)
	rt.Init(p.p)
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

func (p *parser) iLit() *ILit {
	il := new(ILit)
	il.Init(p.p)
	il.Position = p.p
	if p.tok != "literal" {
		p.err("")
	}
	il.Value = p.lit
	p.next()
	return il

}
func (p *parser) wLit() *WLit {
	wl := new(WLit)
	wl.Init(p.p)
	if p.tok != "name" {
		p.err("")
	}
	wl.Value = p.lit
	p.next()
	return wl
}

func (p *parser) varExpr() Expr {
	rt := new(VarExpr)
	rt.Init(p.p)
	rt.Wl = p.wLit()
	return rt
}

func (p *parser) numberExpr() Expr {
	ne := new(NumberExpr)
	ne.Init(p.p)

	ne.Il = p.iLit()
	return ne

}
