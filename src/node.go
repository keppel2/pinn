package main

import "text/scanner"

type Node interface {
	Gpos() scanner.Position
	Init(scanner.Position)
	aNode()
}

type node struct {
	scanner.Position
}

func (n node) Gpos() scanner.Position   { return n.Position }
func (n *node) Init(s scanner.Position) { n.Position = s }
func (node) aNode()                     {}

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

type ReturnStmt struct {
	E Expr
	stmt
}

type ForStmt struct {
	Inits Stmt
	E     Expr
	Loop  Stmt
	B     *BlockStmt
	stmt
}

type ForrStmt struct {
	LH []Expr
	Op string
	RH Expr
	B  *BlockStmt
	stmt
}

type ExprStmt struct {
	Expr
	stmt
}

type AssignStmt struct {
	LHSa []Expr
	RHSa []Expr
	Op   string
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
	B    *BlockStmt
	stmt
}

type LoopStmt struct {
	B *BlockStmt
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
	X Expr
	E Expr
	expr
}

type ArrayExpr struct {
	EL   []Expr
	Dots bool
	expr
}

type Kind interface {
	Node
	aKind()
}

type kind struct{ node }

func (kind) aKind() {}

type MKind struct {
	K Kind
	kind
}

type SlKind struct {
	K Kind
	kind
}

type ArKind struct {
	Len Expr
	K   Kind
	kind
}

type SKind struct {
	Wl *WLit
	kind
}

type VarStmt struct {
	List []*WLit
	Kind
	stmt
}

type Field struct {
	List []*WLit
	Kind
	Dots bool
	node
}

type TypeStmt struct {
	Wl *WLit
	stmt
	Kind
}

type FuncStmt struct {
	Wl    *WLit
	PList []*Field
	K     Kind
	B     *BlockStmt
	stmt
}
