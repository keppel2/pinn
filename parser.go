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
		panic("expecting " + tok)
	}
}

func pnode(n Node) {
  fmt.Println(reflect.TypeOf(n) , n.Gpos())
}

func contains(s []string, t string) bool {
	for _, v := range s {
		if v == t {
			return true
		}
	}
	return false
}

func visitDeclStmt(d DeclStmt) {
  pnode(d)
}

func visitIntExpr(n IntExpr) {
  pnode(n)

}


func visitExpr(n Expr) {
  pnode(n)
  switch t := n.(type) {
  case NumberExpr:
    println("Number",t.Il.Value)
  case VarExpr:
    println("Var",t.Wl.Value)
  case IntExpr:
    visitIntExpr(t)
  }

}
func visitExprStmt(e ExprStmt) {
  pnode(e)
  visitExpr(e.Expr)
}

func visitAssignStmt(a AssignStmt) {
  pnode(a)
}

func visitStmt(s Stmt) {
	switch t := s.(type) {
	case DeclStmt:
		visitDeclStmt(t)
	case ExprStmt:
		visitExprStmt(t)
	case AssignStmt:
		visitAssignStmt(t)
	}
}

func visitFile(f File) {
  pnode(f)
	for _, s := range f.SList {
		visitStmt(s)
	}
}

func (p *parser) fileA() File {
	f := File{}
	f.Position = p.p

	p.next()
	for p.tok != "EOF" {
		switch p.tok {
		case "literal":
			f.SList = append(f.SList, p.exprStmt())
		default:
			panic("tok," + p.tok)
		}
//		if !p.got(";") {
//			panic("No semi")
//		}
	}
	fmt.Println(f.SList)
	visitFile(f)

	return f
}

func (p *parser) declStmt(f func() Decl) DeclStmt {
	ds := DeclStmt{}
	ds.Position = p.p
	ds.Decl = f()
	return ds
}

func (p *parser) exprStmt() ExprStmt {
	es := ExprStmt{}
	es.Position = p.p
	rt := p.expr()
	if p.tok != ";" {
		panic("")
	}
	p.next()
  es.Expr = rt
	return es
}

func (p *parser) expr() Expr {
	var lhs Expr
	switch p.tok {
	case "literal":
		lhs = p.numberExpr()
	case "name":
		lhs = p.varExpr()
	default:
		panic(p.tok)
	}
	if p.tok == ";" {
		return lhs
	}
	if p.tok == "op" {
		return p.intExpr(lhs)
	}
  panic("")
}

func (p *parser) intExpr(lhs Expr) Expr {
	op := p.op
  p.next()
	rhs := p.expr()
	rt := IntExpr{}
	rt.LHS = lhs
	rt.RHS = rhs
	rt.op = op
	return rt
}

func (p *parser) iLit() ILit {
	il := ILit{}
	if p.tok != "literal" {
		panic("")
	}
	il.Value = p.lit
	p.next()
	return il

}
func (p *parser) wLit() WLit {
	wl := WLit{}
	if p.tok != "name" {
		panic("")
	}
	wl.Value = p.lit
	p.next()
	return wl
}

func (p *parser) varExpr() Expr {
  rt := VarExpr{}
  rt.Wl = p.wLit()
  return rt
}

func (p *parser) numberExpr() Expr {
	ne := NumberExpr{}

	ne.Il = p.iLit()
	return ne

}
