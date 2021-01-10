package main

import "fmt"
import "reflect"

const indent = "  "

type visitor struct {
	ilevel int
	s      string
}

func (v *visitor) prn(s ...interface{}) {
	for i := 0; i < v.ilevel; i++ {
		v.s += fmt.Sprint(indent)
	}
	v.s += fmt.Sprintln(s)
}

func (v *visitor) iminus() {
	v.ilevel--
}

func (v *visitor) init() {
}

func (v *visitor) pnode(n Node) {
	v.prn(reflect.TypeOf(n), n.Gpos())
	v.ilevel++
}

func (v *visitor) visitVarStmt(n *VarStmt) {
	for _, vd := range n.List {
		v.prn("var: ", vd.Value)
	}

	v.visitKind(n.Kind)
}

func (v *visitor) visitField(n *Field) {
	for _, vd := range n.List {
		v.prn("fvar: ", vd.Value)
	}

	v.visitKind(n.Kind)
}

func (v *visitor) visitTypeStmt(n *TypeStmt) {
	v.prn("type: ", n.Wl.Value)
	v.visitKind(n.Kind)
}
func (v *visitor) visitForStmt(n *ForStmt) {
	if n.Inits != nil {
		v.prn("init")
		v.visitStmt(n.Inits)
	}
	if n.E != nil {
		v.visitExpr(n.E)
	}
	if n.Loop != nil {
		v.prn("loop")
		v.visitStmt(n.Loop)
	}
	v.visitBlockStmt(n.B)
}

func (v *visitor) visitReturnStmt(n *ReturnStmt) {
	for _, va := range n.EL {
		v.visitExpr(va)
	}
}

func (v *visitor) visitFuncDecl(n *FuncDecl) {
	v.prn("fid: ", n.Wl.Value)
	v.ilevel++
	defer v.iminus()
	for _, vd := range n.PList {
		v.visitField(vd)
	}
	for _, k := range n.K {
		v.visitKind(k)
	}
	if n.B != nil {
		defer v.iminus()
		v.pnode(n.B)
		v.visitBlockStmt(n.B)
	}
}

func (v *visitor) visitKind(n Kind) {
	v.pnode(n)
	defer v.iminus()
	switch t := n.(type) {
	case *MKind:
		v.visitMKind(t)
	case *SlKind:
		v.visitSlKind(t)
	case *ArKind:
		v.visitArKind(t)
	case *SKind:
		v.visitSKind(t)
	}
}
func (v *visitor) visitMKind(n *MKind) {
	v.visitKind(n.K)
}

func (v *visitor) visitSlKind(n *SlKind) {
	v.visitKind(n.K)
}
func (v *visitor) visitArKind(n *ArKind) {
	v.visitExpr(n.Len)
	v.visitKind(n.K)
}
func (v *visitor) visitSKind(n *SKind) {
	v.prn("Skind", n.Wl.Value)
}

func (v *visitor) visitBinaryExpr(n *BinaryExpr) {
	v.visitExpr(n.LHS)
	v.prn("op", n.op, "end")
	v.visitExpr(n.RHS)
}
func (v *visitor) visitTrinaryExpr(n *TrinaryExpr) {
	v.visitExpr(n.LHS)
	v.visitExpr(n.MS)
	v.visitExpr(n.RHS)
}

func (v *visitor) visitCallExpr(n *CallExpr) {
	v.visitExpr(n.ID)
	for _, va := range n.Params {
		v.visitExpr(va)
	}
}

func (v *visitor) visitUnaryExpr(n *UnaryExpr) {
	if n.E != nil {
		v.visitExpr(n.E)
	}
	v.prn("op", n.op)
}

func (v *visitor) visitIndexExpr(n *IndexExpr) {
	v.visitExpr(n.X)
	v.visitExpr(n.E)
	/*
		if n.Start != nil {
			v.visitExpr(n.Start)
		}
		v.prn("Inc", n.Inc)

		if n.End != nil {
			v.visitExpr(n.End)
		}
	*/
}

func (v *visitor) visitArrayExpr(n *ArrayExpr) {
	for _, e := range n.EL {
		v.visitExpr(e)
	}
}

func (v *visitor) visitExpr(n Expr) {
	v.pnode(n)
	defer v.iminus()
	switch t := n.(type) {
	case *TrinaryExpr:
		v.visitTrinaryExpr(t)
	case *NumberExpr:
		v.prn("Number", t.Il.Value)
	case *VarExpr:
		v.prn("Var", t.Wl.Value)
	case *IndexExpr:
		v.visitIndexExpr(t)
	case *BinaryExpr:
		v.visitBinaryExpr(t)
	case *CallExpr:
		v.visitCallExpr(t)
	case *UnaryExpr:
		v.visitUnaryExpr(t)
	case *ArrayExpr:
		v.visitArrayExpr(t)
	}

}
func (v *visitor) visitWhileStmt(n *WhileStmt) {
	v.visitExpr(n.Cond)
	v.visitBlockStmt(n.B)
}

func (v *visitor) visitLoopStmt(n *LoopStmt) {
	v.visitBlockStmt(n.B)
}
func (v *visitor) visitIfStmt(n *IfStmt) {
	v.visitExpr(n.Cond)
	if n.Then != nil {
		v.visitStmt(n.Then)
	}
	if n.Else != nil {
		v.visitStmt(n.Else)
	}
}

func (v *visitor) visitExprStmt(e *ExprStmt) {
	v.visitExpr(e.Expr)
}

func (v *visitor) visitAssignStmt(a *AssignStmt) {
	for _, e := range a.LHSa {
		v.visitExpr(e)
	}
	v.prn("Op", a.Op)
	for _, e := range a.RHSa {
		v.visitExpr(e)
	}
}

func (v *visitor) visitBlockStmt(t *BlockStmt) {
	for _, s := range t.SList {
		v.visitStmt(s)
	}
}

func (v *visitor) visitStmt(s Stmt) {
	v.pnode(s)
	defer v.iminus()
	switch t := s.(type) {
	case *ForStmt:
		v.visitForStmt(t)
	case *BlockStmt:
		v.visitBlockStmt(t)
	case *VarStmt:
		v.visitVarStmt(t)
	case *TypeStmt:
		v.visitTypeStmt(t)
	case *ExprStmt:
		v.visitExprStmt(t)
	case *AssignStmt:
		v.visitAssignStmt(t)
	case *IfStmt:
		v.visitIfStmt(t)
	case *WhileStmt:
		v.visitWhileStmt(t)
	case *LoopStmt:
		v.visitLoopStmt(t)
	case *ReturnStmt:
		v.visitReturnStmt(t)
	}
}

func (v *visitor) visitFile(f *File) {
	v.pnode(f)
	defer v.iminus()
	for _, d := range f.FList {
		v.visitFuncDecl(d)
	}
	for _, s := range f.SList {
		v.visitStmt(s)
	}
}
