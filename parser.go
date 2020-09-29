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

func visitCallExpr(n CallExpr) {
  visitExpr(n.ID)
  for _, v := range n.Params {
    visitExpr(v)
  }
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
  case CallExpr:
    visitCallExpr(t)
  }

}
func visitExprStmt(e ExprStmt) {
  pnode(e)
  visitExpr(e.Expr)
}

func visitAssignStmt(a AssignStmt) {
  pnode(a)
  visitExpr(a.LHS)
  visitExpr(a.RHS)
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

func (p *parser) unaryExpr() Expr {
  switch p.tok {
    case "literal":
      return p.numberExpr()
    case "name":
      return p.varExpr()
  }
  panic(p.tok)
}

func (p *parser) fileA() File {
	f := File{}
	f.Position = p.p

	p.next()
	for p.tok != "EOF" {
		switch p.tok {
		case "literal":
      lhs := p.unaryExpr()
			f.SList = append(f.SList, p.exprStmt(lhs))
    case "name":
      lhs := p.unaryExpr()
      if p.tok == "=" {
        f.SList = append(f.SList, p.assignStmt(lhs))
      } else if p.tok == "(" {
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

func (p *parser) assignStmt(LHS Expr) AssignStmt {
  if !(p.got("=") || p.got(":=")) {
    panic (p.tok)
  }
  rt := AssignStmt{}
  rt.Position = p.p
  rt.LHS = LHS
  ua := p.unaryExpr()
  rt.RHS = p.expr(ua)
  return rt

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
  if p.tok == "," || p.tok == ")" {
    return LHS
  }
	if p.tok == "+" {
		return p.intExpr(LHS)
	}
  if p.tok == "(" {
    return p.callExpr(LHS)
  }
  panic(p.p)
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
