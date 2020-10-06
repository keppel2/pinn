package main

import "fmt"

func visitVarDecl(n VarDecl) {
	fmt.Println("var: ", n.Wl.Value)
	visitKind(n.Kind)
}

func visitTypeDecl(n TypeDecl) {
	fmt.Println("type: ", n.Wl.Value)
	visitKind(n.Kind)
}

func visitDeclStmt(d DeclStmt) {
	pnode(d)
	visitDecl(d.Decl)
}

func visitFuncDecl(n FuncDecl) {
	pnode(n)
	fmt.Println("fid: ", n.Wl.Value)
	for _, vd := range n.PList {
		visitVarDecl(vd)
	}
	fmt.Println(n.Kind)
	visitBlockStmt(n.B)
}

func visitDecl(d Decl) {
	switch t := d.(type) {
	case VarDecl:
		visitVarDecl(t)
	case TypeDecl:
		visitTypeDecl(t)
	case FuncDecl:
		visitFuncDecl(t)
	}
}

func visitKind(n Kind) {
	pnode(n)
	sk := n.(SKind)
	fmt.Println("Skind", sk.Wl.Value)
}

func visitBinaryExpr(n BinaryExpr) {
	visitExpr(n.LHS)
	println("Op", n.op, ".")
	visitExpr(n.RHS)
}

func visitCallExpr(n CallExpr) {
	visitExpr(n.ID)
	for _, v := range n.Params {
		visitExpr(v)
	}
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
		println("Number", t.Il.Value)
	case VarExpr:
		println("Var", t.Wl.Value)
	case IndexExpr:
		visitIndexExpr(t)
	case BinaryExpr:
		visitBinaryExpr(t)
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
