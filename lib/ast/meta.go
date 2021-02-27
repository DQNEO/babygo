package ast

type Func struct {
	Localarea int
	Argsarea  int
	LocalVars []*Variable
	Params    []*Variable
	Retvars   []*Variable
	FuncType  *FuncType
	RcvType   Expr
	Name      string
	Stmts     []Stmt
	Method    *Method
}

type Method struct {
	PkgName      string
	RcvNamedType *Ident
	IsPtrMethod  bool
	Name         string
	FuncType     *FuncType
}


type Type struct {
	//kind string
	E Expr
}

type Variable struct {
	Name         string
	IsGlobal     bool
	GlobalSymbol string
	LocalOffset  int
	Typ          *Type
}
