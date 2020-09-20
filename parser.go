package main

import (
	"fmt"
	"io"
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

func contains(s []string, t string) bool {
	for _, v := range s {
		if v == t {
			return true
		}
	}
	return false
}

func visitDeclStmt(d DeclStmt) {
  fmt.Println("Visit DeclStmt")
}

func visitExprStmt(e ExprStmt) {
  fmt.Println("Visit ExprStmt")
}

func visitAssignStmt(a AssignStmt) {
  fmt.Println("Visit AssignStmt")
}

func visitStmt(s Stmt) {
  fmt.Println("Visit Stmt")
  switch t := s.(type) {
  case DeclStmt:
    visitDeclStmt(t)
  case ExprStmt:
    visitExprStmt(t)
  case AssignStmt:
    visitAssignStmt(t)
  }
}

func visitFile (f File) {
  fmt.Println("Visit File")
  for _, s := range f.SList {
    visitStmt(s)
  }
}

func (p *parser) fileA() File {
	f := File{}
	f.Pos = p.p

	p.next()
	for p.tok != "EOF" {
		switch p.tok {
		case "var":
			p.next()
			f.SList = append(f.SList, p.declStmt(p.varDecl))
		case "literal":
			f.SList = append(f.SList, p.exprStmt())
		default:
			panic("tok," + p.tok)
		}
		if !p.got(";") {
			panic("No semi")
		}
	}
	fmt.Println(f.SList)
  visitFile(f)

	return f
}

func (p *parser) declStmt(f func() Decl) DeclStmt {
	ds := DeclStmt{}
	ds.Pos = p.p
	ds.Decl = f()
	return ds
}

func (p *parser) exprStmt() ExprStmt {
	es := ExprStmt{}
	es.Pos = p.p
	es.Expr = p.numberExpr()
	return es

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

func (p *parser) varDecl() Decl {
	d := VarDecl{}

	d.Wl = p.wLit()
	d.Kind = p.kind()
	return d
}

func (p *parser) numberExpr() Expr {
	ne := NumberExpr{}

	ne.Il = p.iLit()
	return ne

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
	panic("")

}
