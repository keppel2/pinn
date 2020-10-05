package main

import "text/scanner"

type Node interface {
	Gpos() scanner.Position
	aNode()
}

type node struct {
	scanner.Position
}

func (n node) Gpos() scanner.Position { return n.Position }
func (node) aNode()                   {}

type Stmt interface {
	Node
	aStmt()
}

type stmt struct{ node }

func (stmt) aStmt() {}

type File struct {
	SList []Stmt
	node
}

type BlockStmt struct {
	SList []Stmt
	stmt
}

type ExprStmt struct {
	Expr
	stmt
}
type DeclStmt struct {
	Decl
	stmt
}

type AssignStmt struct {
	LHS Expr
	RHS Expr
	stmt
}

type Decl interface {
	Node
	aDecl()
}

type decl struct{ node }

func (decl) aDecl() {}

type Expr interface {
	Node
	aExpr()
}

type expr struct{ node }

func (expr) aExpr() {}

type lit struct{ node }

func (lit) aLit() {}

type Lit interface {
	Node
	aLit()
}

type WLit struct {
	Value string
	lit
}

type ILit struct {
	Value string
	lit
}

type NumberExpr struct {
	Il ILit
	expr
}

type VarExpr struct {
	Wl WLit
	expr
}

type UnaryExpr struct {
	E  Expr
	op string
	expr
}

type BinaryExpr struct {
	LHS Expr
	RHS Expr
	op  string
	expr
}

type CallExpr struct {
	ID     Expr
	Params []Expr
	expr
}

type IndexExpr struct {
  X Expr
  Start Expr
  End Expr
  expr
}

type Kind interface {
	Node
	aKind()
}

type kind struct{ node }

func (kind) aKind() {}

type SKind struct {
	Wl WLit
	kind
}

type VarDecl struct {
	Wl WLit
	Kind
	decl
}

type TypeDecl struct {
	Wl WLit
	Kind
	decl
}

type FuncDecl struct {
	Wl    WLit
	PList []VarDecl
	Kind
	B BlockStmt
	decl
}
