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

func visitVarDecl(n VarDecl) {
  fmt.Println("var: ", n.Wl.Value)
  visitKind(n.Kind)
}

func visitDeclStmt(d DeclStmt) {
  pnode(d)
  visitDecl(d.Decl)
}

func visitDecl(d Decl) {
  switch t := d.(type) {
  case VarDecl:
    visitVarDecl(t)
  }
}

func visitKind(n Kind) {
  pnode(n)
  sk := n.(SKind)
  fmt.Println("Skind",sk.Wl.Value)
}

func visitIntExpr(n IntExpr) {
  visitExpr(n.LHS)
  println("Op",n.op,".")
  visitExpr(n.RHS)
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
      lhs := p.numberExpr()
			f.SList = append(f.SList, p.exprStmt(lhs))
    case "name":
      lhs := p.varExpr()
      if p.tok == "=" {
        f.SList = append(f.SList, p.assignStmt(lhs))
      } else {
        f.SList = append(f.SList, p.exprStmt(lhs))
      }
    case "var":
      f.SList = append(f.SList, p.declStmt())
		default:
			panic("tok," + p.tok)
		}
    p.want(";")
	}
	fmt.Println(f.SList)
	visitFile(f)

	return f
}

func (p *parser) declStmt() DeclStmt {
	ds := DeclStmt{}
	ds.Position = p.p
	ds.Decl = p.varDecl()
	return ds
}

func (p *parser) varDecl() Decl {
       d := VarDecl{}
       d.Position = p.p
       p.want("var")

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

func (p *parser) exprStmt(LHS Expr) ExprStmt {
	es := ExprStmt{}
	es.Position = p.p
	rt := p.expr(LHS)
  es.Expr = rt
	return es
}

func (p *parser) expr(LHS Expr) Expr {
	if p.tok == ";" {
		return LHS
	}
	if p.tok == "op" {
		return p.intExpr(LHS)
	}
  panic("")
}

func (p *parser) intExpr(lhs Expr) Expr {
	op := p.lit
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
  il.Position = p.p
	if p.tok != "literal" {
		panic("")
	}
	il.Value = p.lit
	p.next()
	return il

}
func (p *parser) wLit() WLit {
	wl := WLit{}
  wl.Position = p.p
	if p.tok != "name" {
		panic("")
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
