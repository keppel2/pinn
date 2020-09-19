package main

type Pos interface {
}

type Node interface {
	Gpos() Pos
	aNode()
}

type node struct {
	Pos
}

func (n node) Gpos() Pos { return n.Pos }
func (node) aNode()      {}

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

type Kind interface {
	Node
	aKind()
}

type kind struct{ node }

func (kind) aKind() {}

type SKind struct {
	wl WLit
	kind
}

type VarDecl struct {
	wl WLit
	Kind
	decl
}

type NumberExpr struct {
	il ILit
	expr
}
