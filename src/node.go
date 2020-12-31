package main

import "text/scanner"
import "fmt"

var _fmt = fmt.Print

type Node interface {
	Gpos() scanner.Position
	Init(scanner.Position)
	aNode()
}

type node struct {
	scanner.Position
}

func (n node) Gpos() scanner.Position { return n.Position }
func (n *node) Init(s scanner.Position) {

	n.Position = s
}
func (node) aNode() {}

type Stmt interface {
	Node
	aStmt()
}

type stmt struct{ node }

func (stmt) aStmt() {}

type File struct {
	SList []Stmt
	FList []*FuncDecl
	node
}

func (f *File) getFunc(a string) *FuncDecl {
	for _, v := range f.FList {
		if v.Wl.Value == a {
			return v
		}
	}
	return nil
}

type BlockStmt struct {
	SList []Stmt
	stmt
}

type BreakStmt struct {
	stmt
}

type ContinueStmt struct {
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

type ExprStmt struct {
	Expr
	stmt
}

type AssignStmt struct {
	LHSa   []Expr
	RHSa   []Expr
	Op     string
	irange bool
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

func newW(a string) *WLit {
	rt := new(WLit)
	rt.Value = a
	return rt
}

type StringExpr struct {
	W *WLit
	expr
}

type NumberExpr struct {
	Il *WLit
	expr
}

type VarExpr struct {
	Wl *WLit
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
	ilen int
	Len  Expr
	K    Kind
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

type NameType struct {
	N *WLit
	K Kind
}

type TypeStmt struct {
	Wl *WLit
	stmt
	Kind
}

type FuncDecl struct {
	Wl     *WLit
	PList  []*Field
	NTlist []*NameType
	PCount int
	PSize  int
	K      Kind
	B      *BlockStmt
	node
}

func (fd *FuncDecl) transform() {
	for _, f := range fd.PList {
		k := f.Kind

		for _, w := range f.List {
			nt := new(NameType)
			nt.K = k
			nt.N = w
			fd.NTlist = append(fd.NTlist, nt)
		}
		if ark, ok := k.(*ArKind); ok {
			fd.PSize += len(f.List) * atoi(nil, ark.Len.(*NumberExpr).Il.Value)
		} else {
			fd.PSize += len(f.List)
		}

	}
	fd.PCount = len(fd.NTlist)
}

func (fd *FuncDecl) getKind(a int) Kind {
	index := 0
	for _, f := range fd.PList {
		for _ = range f.List {
			if index == a {
				return f.Kind
			}
			index++
		}
	}
	return nil
}
