package main

import "fmt"

func visitVarStmt(n VarStmt) {
	for _, vd := range n.List {
		fmt.Println("var: ", vd.Value)
	}

	visitKind(n.Kind)
}

func visitField(n Field) {
	for _, vd := range n.List {
		fmt.Println("fvar: ", vd.Value)
	}

	visitKind(n.Kind)
}

func visitTypeStmt(n TypeStmt) {
	fmt.Println("type: ", n.Wl.Value)
	visitKind(n.Kind)
}

func visitFuncStmt(n FuncStmt) {
	pnode(n)
	fmt.Println("fid: ", n.Wl.Value)
	for _, vd := range n.PList {
		visitField(vd)
	}
	fmt.Println(n.Kind)
	visitBlockStmt(n.B)
}

func visitKind(n Kind) {
	pnode(n)
	sk := n.(SKind)
	fmt.Println("Skind", sk.Wl.Value)
}

func visitBinaryExpr(n BinaryExpr) {
	visitExpr(n.LHS)
	fmt.Println("Op", n.op, ".")
	visitExpr(n.RHS)
}
func visitTrinaryExpr(n TrinaryExpr) {
	visitExpr(n.LHS)
	visitExpr(n.MS)
	visitExpr(n.RHS)
}

func visitCallExpr(n CallExpr) {
	visitExpr(n.ID)
	for _, v := range n.Params {
		visitExpr(v)
	}
}

func visitUnaryExpr(n UnaryExpr) {
	visitExpr(n.E)
	fmt.Println("Uop", n.op)
}

func visitIndexExpr(n IndexExpr) {
	visitExpr(n.X)
	if n.Start != nil {
		visitExpr(n.Start)
	}

	if n.End != nil {
		visitExpr(n.End)
	}
}

func visitExpr(n Expr) {
	pnode(n)
	switch t := n.(type) {
	case NumberExpr:
		fmt.Println("Number", t.Il.Value)
	case VarExpr:
		fmt.Println("Var", t.Wl.Value)
	case IndexExpr:
		visitIndexExpr(t)
	case BinaryExpr:
		visitBinaryExpr(t)
	case CallExpr:
		visitCallExpr(t)
	case UnaryExpr:
		visitUnaryExpr(t)
	}

}
func visitWhileStmt(n WhileStmt) {
	visitExpr(n.Cond)
	visitBlockStmt(n.B)
}

func visitIfStmt(n IfStmt) {
	visitExpr(n.Cond)
	visitStmt(n.Then)
	visitStmt(n.Else)
}

func visitExprStmt(e ExprStmt) {
	pnode(e)
	visitExpr(e.Expr)
}

func visitAssignStmt(a AssignStmt) {
	pnode(a)
	visitExpr(a.LHS)
	fmt.Println("Op", a.Op)
	if a.RHS != nil {
		visitExpr(a.RHS)
	}
}

func visitBlockStmt(t BlockStmt) {
	pnode(t)
	for _, s := range t.SList {
		visitStmt(s)
	}
}

func visitStmt(s Stmt) {
	switch t := s.(type) {
	case BlockStmt:
		visitBlockStmt(t)
	case VarStmt:
		visitVarStmt(t)
	case TypeStmt:
		visitTypeStmt(t)
	case FuncStmt:
		visitFuncStmt(t)
	case ExprStmt:
		visitExprStmt(t)
	case AssignStmt:
		visitAssignStmt(t)
	case IfStmt:
		visitIfStmt(t)
	case WhileStmt:
		visitWhileStmt(t)
	}
}

func visitFile(f File) {
	pnode(f)
	for _, s := range f.SList {
		visitStmt(s)
	}
}
