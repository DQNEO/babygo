package ast

import "github.com/DQNEO/babygo/lib/token"

var Con string = "Con"
var Typ string = "Typ"
var Var string = "Var"
var Fun string = "Fun"
var Pkg string = "Pkg"

type Signature struct {
	Params  *FieldList
	Results *FieldList
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

type Object struct {
	Kind     string
	Name     string
	Decl     interface{} // *ValueSpec|*FuncDecl|*TypeSpec|*Field|*AssignStmt
	Variable *Variable
	PkgName  string
}

type Expr interface{}

type Field struct {
	Name   *Ident
	Type   Expr
	Offset int
}

type FieldList struct {
	List []*Field
}

type Ident struct {
	Name string
	Obj  *Object
}

type Ellipsis struct {
	Elt Expr
}

type BasicLit struct {
	Kind  token.Token // token.INT, token.CHAR, or token.STRING
	Value string
}

type CompositeLit struct {
	Type Expr
	Elts []Expr
}

type KeyValueExpr struct {
	Key   Expr
	Value Expr
}

type ParenExpr struct {
	X Expr
}

type SelectorExpr struct {
	X   Expr
	Sel *Ident
}

type IndexExpr struct {
	X     Expr
	Index Expr
}

type SliceExpr struct {
	X      Expr
	Low    Expr
	High   Expr
	Max    Expr
	Slice3 bool
}

type CallExpr struct {
	Fun      Expr   // function expression
	Args     []Expr // function arguments; or nil
	Ellipsis token.Pos
}

type StarExpr struct {
	X Expr
}

type UnaryExpr struct {
	X  Expr
	Op token.Token
}

type BinaryExpr struct {
	X  Expr
	Y  Expr
	Op token.Token
}

type TypeAssertExpr struct {
	X    Expr
	Type Expr // asserted type; nil means type switch X.(type)
}

// Type nodes
type ArrayType struct {
	Len Expr
	Elt Expr
}

type StructType struct {
	Fields *FieldList
}

type InterfaceType struct {
	Methods []string
}

type FuncType struct {
	Params  *FieldList
	Results *FieldList
}

type Stmt interface{}

type DeclStmt struct {
	Decl Decl
}

type ExprStmt struct {
	X Expr
}

type IncDecStmt struct {
	X   Expr
	Tok token.Token
}

type AssignStmt struct {
	Lhs     []Expr
	Tok     token.Token
	Rhs     []Expr
	IsRange bool
}

type ReturnStmt struct {
	Results []Expr
	Meta    *MetaReturnStmt
}

type BranchStmt struct {
	Tok        token.Token
	Label      string
	CurrentFor *MetaForStmt
}

type BlockStmt struct {
	List []Stmt
}

type IfStmt struct {
	Init Stmt
	Cond Expr
	Body *BlockStmt
	Else Stmt
}

type CaseClause struct {
	List []Expr
	Body []Stmt
}

type SwitchStmt struct {
	Init Expr
	Tag  Expr
	Body *BlockStmt
	// lableExit string
}

type TypeSwitchStmt struct {
	Assign Stmt
	Body   *BlockStmt
	Node   *NodeTypeSwitchStmt
}

type Func struct {
	Localarea int
	Argsarea  int
	LocalVars []*Variable
	Params    []*Variable
	Retvars   []*Variable
	FuncType  *FuncType
	RcvType   Expr
	Name      string
	Body      *BlockStmt
	Method    *Method
}

type Method struct {
	PkgName      string
	RcvNamedType *Ident
	IsPtrMethod  bool
	Name         string
	FuncType     *FuncType
}

type MetaReturnStmt struct {
	Fnc *Func
}

type NodeTypeSwitchStmt struct {
	Subject         Expr
	SubjectVariable *Variable
	AssignIdent     *Ident
	Cases           []*TypeSwitchCaseClose
}

type TypeSwitchCaseClose struct {
	Variable     *Variable
	VariableType *Type
	Orig         *CaseClause
}

type ForStmt struct {
	Init Stmt
	Cond Expr
	Post Stmt
	Body *BlockStmt
	Meta *MetaForStmt
}

type MetaForStmt struct {
	LabelPost   string
	LabelExit   string
	RngLenvar   *Variable
	RngIndexvar *Variable
	Outer       *MetaForStmt
}

type RangeStmt struct {
	Key   Expr
	Value Expr
	X     Expr
	Body  *BlockStmt
	Meta  *MetaForStmt
	Tok   token.Token
}

type ImportSpec struct {
	Path string
}

type ValueSpec struct {
	Names  []*Ident
	Type   Expr
	Values []Expr
}

type TypeSpec struct {
	Name   *Ident
	Assign bool // isAlias
	Type   Expr
}

// Pseudo interface for *ast.Decl
// *GenDecl | *FuncDecl
type Decl interface {
}

type Spec interface{}

type GenDecl struct {
	Specs []Spec
}

type FuncDecl struct {
	Recv *FieldList
	Name *Ident
	Type *FuncType
	Body *BlockStmt
}

type File struct {
	Name       string
	Imports    []*ImportSpec
	Decls      []Decl
	Unresolved []*Ident
	Scope      *Scope
}

type Scope struct {
	Outer   *Scope
	Objects []*ObjectEntry
}

type ObjectEntry struct {
	Name string
	Obj  *Object
}

func NewScope(outer *Scope) *Scope {
	return &Scope{
		Outer: outer,
	}
}

func (s *Scope) Insert(obj *Object) {
	if s == nil {
		panic("s sholud not be nil\n")
	}

	s.Objects = append(s.Objects, &ObjectEntry{
		Name: obj.Name,
		Obj:  obj,
	})
}

func (s *Scope) Lookup(name string) *Object {
	for _, oe := range s.Objects {
		if oe.Name == name {
			return oe.Obj
		}
	}

	return nil
}
