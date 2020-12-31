package main

import (
	"fmt"
	"io"
	//	"strconv"
	//	"os"
)

type errp interface {
	err(string)
}

type parser struct {
	s       *scan
	dm      map[string]string
	qcounts []int
}

func (p *parser) next() {
	p.s.cursor++
}

func (p *parser) err(msg string) {
	panic(fmt.Sprintln(msg, p.s.ct().p, p.s.ct().tok, p.s.ct().lit))
}

func (p *parser) init(r io.Reader) {
	p.dm = make(map[string]string)
	p.s = new(scan)
	p.s.init(r)
	p.s.qmarks = make([]tlt, 1)
}

func (p *parser) got(tok string) bool {
	if p.s.ct().tok == tok {
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

func (p *parser) unaryExpr() Expr {
	switch p.s.ct().tok {
	case "-", "+", "!", "@", "*", "&", ":":
		ue := new(UnaryExpr)
		ue.Init(p.s.ct().p)
		ue.op = p.s.ct().tok
		p.next()
		if p.s.ct().tok == "]" {
			if ue.op != ":" {
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

	//  return p.s.ct().primaryExpr()
}

func (p *parser) primaryExpr() Expr {
	x := p.operand()
	for {
		switch p.s.ct().tok {
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
	switch p.s.ct().tok {
	case "name":
		return p.varExpr()
	case "literal":
		if p.s.ct().lk == StringLit {
			return p.stringExpr()
		} else if p.s.ct().lk == IntLit {
			return p.numberExpr()
		} else {
			p.err(p.s.ct().tok)
		}
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
	rt := new(ArrayExpr)
	rt.Init(p.s.ct().p)
	rt.EL = p.exprList()
	p.want("]")
	rt.Dots = d
	return rt
}

func (p *parser) pseudoF(ID string, count int) *FuncDecl {
	fd := p.newFuncDecl()
	wl := p.newWLit()
	wl.Value = ID
	fd.Wl = wl
	fd.PCount = count
	fd.PSize = count
	return fd
}

func (p *parser) fileA() *File {
	p.next()
	f := new(File)
	f.Init(p.s.ct().p)
	ss := make([]string, 0)
	for p.got("#") {
		p.want("define")
		str := p.s.ct().lit
		for p.s.ct().tok != ";" {
			t := p.s.ct().tok
			if t == "name" || t == "literal" {
				t = p.s.ct().lit
			}
			if p.dm[t] != "" {
				t = p.dm[t]
			}
			ss = append(ss, p.s.ct().tok)
			p.next()
		}
		p.next()
		if len(ss) == 3 {
			if ss[1] == "*" {
				na := atoi(p, ss[0])
				nb := atoi(p, ss[2])
				nc := na * nb
				p.dm[str] = fmt.Sprint(nc)
			}
		} else {
			p.dm[str] = ss[0]
		}
	}
	f.FList = append(f.FList, p.pseudoF("printdec", 1), p.pseudoF("print", 1), p.pseudoF("println", 0), p.pseudoF("printchar", 1), p.pseudoF("printch", -1))

	for p.s.ct().tok != "EOF" {
		if p.s.ct().tok == "func" {
			fd := p.funcDecl()
			if f.getFunc(fd.Wl.Value) != nil {
				p.err("")
			}
			f.FList = append(f.FList, fd)
		} else {
			f.SList = append(f.SList, p.stmt())
		}
	}
	return f
}

func (p *parser) loopStmt() *LoopStmt {
	p.want("loop")
	rt := new(LoopStmt)
	rt.Init(p.s.ct().p)
	rt.B = p.blockStmt()
	return rt
}
func (p *parser) whileStmt() *WhileStmt {
	p.want("while")
	rt := new(WhileStmt)
	rt.Init(p.s.ct().p)
	rt.Cond = p.uexpr()
	rt.B = p.blockStmt()
	return rt
}

func (p *parser) ifStmt() *IfStmt {
	p.want("if")
	rt := new(IfStmt)
	rt.Init(p.s.ct().p)
	rt.Cond = p.uexpr()
	rt.Then = p.stmt()
	if p.got("else") {
		rt.Else = p.stmt()
	}
	return rt
}

func (p *parser) field() *Field {
	n := new(Field)
	n.node.Init(p.s.ct().p)
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
	ds.node.Init(p.s.ct().p)
	ds.Position = p.s.ct().p
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
	ds.node.Init(p.s.ct().p)
	ds.Position = p.s.ct().p
	p.want("type")

	ds.Wl = p.wLit()
	ds.Kind = p.kind()

	p.want(";")
	return ds

}

func (p *parser) newFuncDecl() *FuncDecl {
	rt := new(FuncDecl)
	rt.Init(p.s.ct().p)
	return rt
}

func (p *parser) funcDecl() *FuncDecl {
	rt := new(FuncDecl)
	rt.Init(p.s.ct().p)
	p.want("func")
	rt.Wl = p.wLit()
	if fmap[rt.Wl.Value] != nil {
		p.err(rt.Wl.Value)
	}

	p.want("(")
	if !p.got(")") {
		for {
			vd := p.field()
			rt.PList = append(rt.PList, vd)
			if !p.got(",") {
				break
			}
		}
		p.want(")")
	}

	if p.s.ct().tok != "{" {
		rt.K = p.kind()
	}
	rt.B = p.blockStmt()
	/*
	  if rt.K != nil {
	    if ls, ok := rt.B.SList[len(rt.B.SList - 1)].(*ReturnStmt); ok {
	      if !ok {p.err(rt.Wl.Value)}
	    }
	  }
	*/

	rt.transform()

	return rt
}

func (p *parser) breakStmt() *BreakStmt {
	rt := new(BreakStmt)
	rt.Init(p.s.ct().p)
	p.want("break")
	p.want(";")
	return rt
}

func (p *parser) continueStmt() *ContinueStmt {
	rt := new(ContinueStmt)
	rt.Init(p.s.ct().p)
	p.want("continue")
	p.want(";")
	return rt
}

func (p *parser) returnStmt() *ReturnStmt {
	rt := new(ReturnStmt)
	rt.Init(p.s.ct().p)
	p.want("return")
	if !p.got(";") {
		rt.E = p.uexpr()
		p.want(";")
	}
	return rt
}

func (p *parser) forStmt() *ForStmt {
	rt := new(ForStmt)
	rt.Init(p.s.ct().p)
	p.want("for")
	if !p.got(";") {
		rt.Inits = p.assignOrExprStmt()
		if p.s.ct().tok == "{" {
			rt.B = p.blockStmt()
			return rt
		}
		p.want(";")
	}
	if !p.got(";") {
		rt.E = p.uexpr()
		p.want(";")
	}
	if p.s.ct().tok != "{" {
		rt.Loop = p.assignOrExprStmt()
	}
	rt.B = p.blockStmt()
	return rt
}

func (p *parser) assignOrExprStmt() Stmt {
	lhsa := p.exprList()
	var rt Stmt
	if p.s.ct().tok == "=" || p.s.ct().tok == ":=" || p.s.ct().tok == "+=" || p.s.ct().tok == "-=" || p.s.ct().tok == "*=" || p.s.ct().tok == "/=" || p.s.ct().tok == "%=" || p.s.ct().tok == "++" || p.s.ct().tok == "--" {
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
	switch p.s.ct().tok {
	case "for":
		rt = p.forStmt()
	case "return":
		rt = p.returnStmt()
	case "break":
		rt = p.breakStmt()
	case "continue":
		rt = p.continueStmt()
	case "var":
		rt = p.varStmt()
	case "type":
		rt = p.typeStmt()
	case "if":
		rt = p.ifStmt()
	case "while":
		rt = p.whileStmt()
	case "loop":
		rt = p.loopStmt()

	case "literal", "name", "(", "-", "*", "&", "[", "+":
		rt = p.assignOrExprStmt()
		p.want(";")
	case "{":
		rt = p.blockStmt()
	case ";":
		p.next()
	default:
		p.err("")
	}
	if p.s.qmark() != nil && p.s.qmark().colons < 1 {
		p.err("sq")
	}
	p.s.qmarks = make([]tlt, 1)
	return rt

}

func (p *parser) stmtList() []Stmt {
	rt := make([]Stmt, 0)
	for p.s.ct().tok != "}" {
		rt = append(rt, p.stmt())
	}
	return rt
}

func (p *parser) exprList() []Expr {
	rt := make([]Expr, 0)
	rt = append(rt, p.uexpr())
	for p.s.ct().tok == "," {
		p.next()
		rt = append(rt, p.uexpr())
	}
	return rt
}

func (p *parser) blockStmt() *BlockStmt {
	rt := new(BlockStmt)
	rt.Init(p.s.ct().p)
	p.want("{")
	rt.SList = p.stmtList()
	p.want("}")
	return rt
}

func (p *parser) sKind() *SKind {
	rt := new(SKind)
	rt.Init(p.s.ct().p)
	rt.Wl = p.wLit()
	return rt
}

func (p *parser) kind() Kind {
	switch p.s.ct().tok {
	case "[":
		p.want("[")
		if p.got("]") {
			rt := new(SlKind)
			rt.Init(p.s.ct().p)
			rt.K = p.kind()
			return rt
		}
		if p.got("map") {
			p.want("]")
			rt := new(MKind)
			rt.Init(p.s.ct().p)
			rt.K = p.kind()
			return rt
		}
		rt := new(ArKind)
		rt.Init(p.s.ct().p)
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
	rt.Init(p.s.ct().p)
	rt.Op = p.s.ct().tok
	p.next()
	if p.got("range") {
		rt.irange = true
	}
	rt.LHSa = LHSa
	if rt.Op == "++" || rt.Op == "--" {
	} else {
		rt.RHSa = p.exprList()
	}
	return rt

}

func (p *parser) exprStmt(LHS Expr) *ExprStmt {
	es := new(ExprStmt)
	es.node.Init(p.s.ct().p)
	es.Expr = LHS
	return es
}

func (p *parser) pexpr(prec int) Expr {
	rt := p.unaryExpr()

	for tokenMap[p.s.ct().tok] > prec {

		if p.s.ct().tok == "?" {
			return p.trinaryExpr(rt)
		}
		if p.s.ct().tok == ":" {
			if p.qcounts[len(p.qcounts)-1] > 0 {
				return rt
			}
		}

		t := new(BinaryExpr)
		t.Init(p.s.ct().p)
		t.op = p.s.ct().tok
		t.LHS = rt
		prec := tokenMap[p.s.ct().tok]
		p.next()
		if p.s.ct().tok == "]" {
			return rt
		}
		t.RHS = p.pexpr(prec)
		rt = t
	}
	return rt

}

func (p *parser) uexpr() Expr {
	p.qcounts = append(p.qcounts, 0)
	rt := p.pexpr(0)
	p.qcounts = p.qcounts[0 : len(p.qcounts)-1]
	return rt
}

func (p *parser) trinaryExpr(lhs Expr) Expr {
	rt := new(TrinaryExpr)
	rt.Init(p.s.ct().p)
	rt.LHS = lhs
	p.want("?")

	p.qcounts[len(p.qcounts)-1]++
	rt.MS = p.pexpr(0)
	p.want(":")
	p.qcounts[len(p.qcounts)-1]--
	rt.RHS = p.pexpr(0)
	return rt
}

func (p *parser) indexExpr(lhs Expr) Expr {
	p.want("[")
	rt := new(IndexExpr)
	rt.Init(p.s.ct().p)
	rt.X = lhs
	if p.got("]") {
		return rt
	}
	rt.E = p.uexpr()
	p.want("]")
	return rt
}

func (p *parser) callExpr(lhs Expr) Expr {
	p.want("(")
	rt := new(CallExpr)
	rt.Init(p.s.ct().p)
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

func (p *parser) newWLit() *WLit {
	rt := new(WLit)
	rt.Init(p.s.ct().p)
	return rt
}

func (p *parser) wLit() *WLit {
	wl := p.newWLit()
	wl.Value = p.s.ct().lit
	p.next()
	return wl
}

func (p *parser) varExpr() Expr {
	rt := new(VarExpr)
	rt.Init(p.s.ct().p)
	w := p.wLit().Value
	if rep, ok := p.dm[w]; ok {
		nrt := new(NumberExpr)
		nrt.Init(p.s.ct().p)
		nrt.Il = newW(rep)
		return nrt
	} else {
		wl := new(WLit)
		wl.Init(p.s.ct().p)
		wl.Value = w
		rt.Wl = wl
	}
	return rt
}
func (p *parser) stringExpr() Expr {
	ne := new(StringExpr)
	ne.Init(p.s.ct().p)
	ne.W = p.wLit()

	str := ne.W.Value
	str = str[1 : len(str)-1]
	ne.W.Value = str

	return ne

}
func (p *parser) numberExpr() Expr {
	ne := new(NumberExpr)
	ne.Init(p.s.ct().p)

	ne.Il = p.wLit()
	return ne

}
