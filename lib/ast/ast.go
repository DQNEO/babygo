package ast

import (
	"github.com/DQNEO/babygo/lib/mymap"
	"github.com/DQNEO/babygo/lib/token"
)

var Con ObjKind = "Con"
var Typ ObjKind = "Typ"
var Var ObjKind = "Var"
var Fun ObjKind = "Fun"
var Pkg ObjKind = "Pkg"

type Signature struct {
	Params  *FieldList
	Results *FieldList
}

type ObjKind string

func (ok ObjKind) String() string {
	return string(ok)
}

type Object struct {
	Kind     ObjKind
	Name     string
	Decl     interface{} // *ValueSpec|*FuncDecl|*TypeSpec|*Field|*AssignStmt
	Variable *Variable
	PkgName  string
}

type Expr interface{}

type Field struct {
	Names  []*Ident
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
	Node   *MetaTypeSwitchStmt
}

type ForStmt struct {
	Init Stmt
	Cond Expr
	Post Stmt
	Body *BlockStmt
	Meta *MetaForStmt
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
	Path *BasicLit
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
	Name       *Ident
	Imports    []*ImportSpec
	Decls      []Decl
	Unresolved []*Ident
	Scope      *Scope
}

type Scope struct {
	Outer   *Scope
	Objects *mymap.Map
}

func NewScope(outer *Scope) *Scope {
	return &Scope{
		Outer: outer,
		Objects: &mymap.Map{},
	}
}

func (s *Scope) Insert(obj *Object) {
	if s == nil {
		panic("s sholud not be nil\n")
	}
	s.Objects.Set(obj.Name, obj)
}

func (s *Scope) Lookup(name string) *Object {
	v, ok := s.Objects.Get(name)
	if !ok {
		return nil
	}

	return v.(*Object)
}
