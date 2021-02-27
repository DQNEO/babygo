package ast


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

