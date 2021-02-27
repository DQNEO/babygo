package main

import (
	"github.com/DQNEO/babygo/lib/ast"
)

type Type = ast.Type
type Variable = ast.Variable
type Func = ast.Func

type MetaForStmt struct {
	LabelPost   string
	LabelExit   string
	RngLenvar   *Variable
	RngIndexvar *Variable
	Outer       *MetaForStmt
}

type Method = ast.Method
