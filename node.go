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

type AssignStmt struct {
	LHS Expr
	RHS Expr
	Op  string
	stmt
}

type IfStmt struct {
	Cond Expr
	Then Stmt
	Else Stmt
	stmt
}

type WhileStmt struct {
	Cond Expr
	B    BlockStmt
	stmt
}

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

type TrinaryExpr struct {
	LHS Expr
	MS  Expr
	RHS Expr
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
	X     Expr
	Start Expr
	End   Expr
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

type VarStmt struct {
	List []WLit
	Kind
	stmt
}

type Field struct {
	List []WLit
	Kind
	Dots bool
	node
}

type TypeStmt struct {
	Wl WLit
	Kind
	stmt
}

type FuncStmt struct {
	Wl    WLit
	PList []Field
	Kind
	B BlockStmt
	stmt
}
