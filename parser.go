package main

import (
	"fmt"
	"io"
	"reflect"
	//	"os"
)

type parser struct {
	scan
}

func (p *parser) err(msg string) {
	panic(fmt.Sprintln(msg, p.p))
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

func pnode(n Node) {
	fmt.Println(reflect.TypeOf(n), n.Gpos())
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
	case "literal":
		return p.numberExpr()
	case "name":
		return p.varExpr()
	}
	p.err(p.tok)
	return nil
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

func (p *parser) funcStmt() DeclStmt {
	ds := DeclStmt{}
	ds.Decl = p.funcDecl()
	return ds
}

func (p *parser) varStmt() DeclStmt {
	p.want("var")
	ds := DeclStmt{}
	ds.Position = p.p
	ds.Decl = p.varDecl()
		p.want(";")
	return ds
}

func (p *parser) typeStmt() DeclStmt {
	ds := DeclStmt{}
	ds.Position = p.p
	ds.Decl = p.typeDecl()
		p.want(";")
	return ds

}

func (p *parser) funcDecl() Decl {
	rt := FuncDecl{}
	p.want("func")
	rt.Wl = p.wLit()
  p.want("(")
	if !p.got(")") {
		vd := p.varDecl().(VarDecl)
		rt.PList = append(rt.PList, vd)
		for p.got(",") {
			vd = p.varDecl().(VarDecl)
			rt.PList = append(rt.PList, vd)
		}
    p.want(")")
	}
	if p.tok != "{" {
    rt.Kind = p.kind()
	}
	rt.B = p.blockStmt().(BlockStmt)

	return rt
}

func (p *parser) stmtList() []Stmt {
	rt := make([]Stmt, 0)
	for p.tok != "EOF" && p.tok != "}" {
		switch p.tok {
		case "literal":
			lhs := p.unaryExpr()
			rt = append(rt, p.exprStmt(lhs))
		case "name":
			lhs := p.unaryExpr()
			if p.tok == "=" {
				rt = append(rt, p.assignStmt(lhs))
			} else if p.tok == "(" {
				rt = append(rt, p.exprStmt(lhs))
			}
		case "var":
			rt = append(rt, p.varStmt())
		case "type":
			rt = append(rt, p.typeStmt())
		case "func":
			rt = append(rt, p.funcStmt())
		default:
			p.err(p.tok)
		}

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

func (p *parser) typeDecl() Decl {
	d := TypeDecl{}
	d.Position = p.p
	p.want("type")

	d.Wl = p.wLit()
	d.Kind = p.kind()
	return d

}

func (p *parser) varDecl() Decl {
	d := VarDecl{}
	d.Position = p.p

	d.Wl = p.wLit()
	d.Kind = p.kind()
	return d
}

func (p *parser) sKind() SKind {
	rt := SKind{}
	rt.Wl = p.wLit()
	return rt
}

func (p *parser) kind() Kind {
	switch p.tok {
	case "name":
		return p.sKind()
	}
	panic(p.tok)
}

func (p *parser) assignStmt(LHS Expr) AssignStmt {
	if !(p.got("=") || p.got(":=")) {
		p.err(p.tok)
	}
	rt := AssignStmt{}
	rt.Position = p.p
	rt.LHS = LHS
	ua := p.unaryExpr()
	rt.RHS = p.expr(ua)
		p.want(";")
	return rt

}

func (p *parser) exprStmt(LHS Expr) ExprStmt {
	es := ExprStmt{}
	es.Position = p.p
	rt := p.expr(LHS)
	es.Expr = rt
		p.want(";")
	return es
}

func (p *parser) expr(LHS Expr) Expr {
	if p.tok == ";" || p.tok == "," || p.tok == ")" {
		return LHS
	}
	if p.tok == "+" || p.tok == "-" || p.tok == "/" || p.tok == "*" || p.tok == "<" || p.tok == ">" || p.tok == "==" {
		return p.intExpr(LHS)
	}
	if p.tok == "(" {
		return p.callExpr(LHS)
	}
	p.err("")
	return nil
}

func (p *parser) intExpr(lhs Expr) Expr {
	op := p.lit
	p.next()
	rt := IntExpr{}
	rt.LHS = lhs
	rt.op = op
	rhs := p.unaryExpr()
	rt.RHS = p.expr(rhs)

	return rt
}

func (p *parser) callExpr(lhs Expr) Expr {
	p.want("(")
	rt := CallExpr{}
	rt.ID = lhs
	if p.got(")") {
		return rt
	}
	e := p.expr(p.unaryExpr())
	rt.Params = append(rt.Params, e)
	for p.got(",") {
		rt.Params = append(rt.Params, p.expr(p.unaryExpr()))
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
