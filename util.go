package main

import (
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/fmt"
)

type Type = ast.Type

func newStmt(x interface{}) ast.Stmt {
	return x
}

func isStmtAssignStmt(s ast.Stmt) bool {
	var ok bool
	_, ok = s.(*ast.AssignStmt)
	return ok
}

func isStmtCaseClause(s ast.Stmt) bool {
	var ok bool
	_, ok = s.(*ast.CaseClause)
	return ok
}

func stmt2AssignStmt(s ast.Stmt) *ast.AssignStmt {
	var r *ast.AssignStmt
	var ok bool
	r, ok = s.(*ast.AssignStmt)
	if !ok {
		panic("Not *ast.AssignStmt")
	}
	return r
}

func stmt2ExprStmt(s ast.Stmt) *ast.ExprStmt {
	var r *ast.ExprStmt
	var ok bool
	r, ok = s.(*ast.ExprStmt)
	if !ok {
		panic("Not *ast.ExprStmt")
	}
	return r
}

func stmt2CaseClause(s ast.Stmt) *ast.CaseClause {
	var r *ast.CaseClause
	var ok bool
	r, ok = s.(*ast.CaseClause)
	if !ok {
		panic("Not *ast.CaseClause")
	}
	return r
}

func expr2Ident(e ast.Expr) *ast.Ident {
	var r *ast.Ident
	var ok bool
	r, ok = e.(*ast.Ident)
	if !ok {
		panic(fmt.Sprintf("Not *ast.Ident but got: %T", e))
	}
	return r
}

func expr2UnaryExpr(e ast.Expr) *ast.UnaryExpr {
	var r *ast.UnaryExpr
	var ok bool
	r, ok = e.(*ast.UnaryExpr)
	if !ok {
		panic("Not *ast.UnaryExpr")
	}
	return r
}

func expr2Ellipsis(e ast.Expr) *ast.Ellipsis {
	var r *ast.Ellipsis
	var ok bool
	r, ok = e.(*ast.Ellipsis)
	if !ok {
		panic("Not *ast.Ellipsis")
	}
	return r
}

func expr2TypeAssertExpr(e ast.Expr) *ast.TypeAssertExpr {
	var r *ast.TypeAssertExpr
	var ok bool
	r, ok = e.(*ast.TypeAssertExpr)
	if !ok {
		panic("Not *ast.TypeAssertExpr")
	}
	return r
}

func expr2ArrayType(e ast.Expr) *ast.ArrayType {
	var r *ast.ArrayType
	var ok bool
	r, ok = e.(*ast.ArrayType)
	if !ok {
		panic("Not *ast.ArrayType")
	}
	return r
}

func expr2BasicLit(e ast.Expr) *ast.BasicLit {
	var r *ast.BasicLit
	var ok bool
	r, ok = e.(*ast.BasicLit)
	if !ok {
		panic("Not *ast.BasicLit")
	}
	return r
}

func expr2StarExpr(e ast.Expr) *ast.StarExpr {
	var r *ast.StarExpr
	var ok bool
	r, ok = e.(*ast.StarExpr)
	if !ok {
		panic("Not *ast.StarExpr")
	}
	return r
}

func isExprStarExpr(e ast.Expr) bool {
	_, ok := e.(*ast.StarExpr)
	return ok
}

func isExprEllipsis(e ast.Expr) bool {
	_, ok := e.(*ast.Ellipsis)
	return ok
}

func isExprTypeAssertExpr(e ast.Expr) bool {
	_, ok := e.(*ast.TypeAssertExpr)
	return ok
}

func isExprIdent(e ast.Expr) bool {
	_, ok := e.(*ast.Ident)
	return ok
}

func dtypeOf(x interface{}) string {
	return fmt.Sprintf("%T", x)
}
