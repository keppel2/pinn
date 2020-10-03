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

func visitDecl(d Decl) {
	switch t := d.(type) {
	case VarDecl:
		visitVarDecl(t)
  case TypeDecl:
    visitTypeDecl(t)
	}
}

func visitKind(n Kind) {
	pnode(n)
	sk := n.(SKind)
	fmt.Println("Skind", sk.Wl.Value)
}

func visitIntExpr(n IntExpr) {
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

func visitExpr(n Expr) {
	pnode(n)
	switch t := n.(type) {
	case NumberExpr:
		println("Number", t.Il.Value)
	case VarExpr:
		println("Var", t.Wl.Value)
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


