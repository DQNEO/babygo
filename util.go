package main

import (
	"github.com/DQNEO/babygo/lib/ast"
)

type Type = ast.Type

type Variable struct {
	Name         string
	IsGlobal     bool
	GlobalSymbol string
	LocalOffset  int
	Typ          *Type
}

type Func struct {
	Localarea int
	Argsarea  int
	LocalVars []*Variable
	Params    []*Variable
	Retvars   []*Variable
	FuncType  *ast.FuncType
	RcvType   ast.Expr
	Name      string
	Stmts     []ast.Stmt
	Method    *Method
}

type MetaForStmt struct {
	LabelPost   string
	LabelExit   string
	RngLenvar   *Variable
	RngIndexvar *Variable
	Outer       *MetaForStmt
}

type Method = ast.Method
