package ast

import (
	"github.com/DQNEO/babygo/lib/fmt"
	"github.com/DQNEO/babygo/lib/token"
)

var Con ObjKind = "Con"
var Typ ObjKind = "Typ"
var Var ObjKind = "Var"
var Fun ObjKind = "Fun"
var Pkg ObjKind = "Pkg"

type Signature struct {
	Params   *FieldList
	Results  *FieldList
	StartPos token.Pos
}

type ObjKind string

func (ok ObjKind) String() string {
	return string(ok)
}

type Object struct {
	Kind ObjKind
	Name string
	Decl interface{} // *ValueSpec|*FuncDecl|*TypeSpec|*Field|*AssignStmt
	Data interface{}
}

type Expr interface {
	//Pos() token.Pos
}

type Field struct {
	Names  []*Ident
	Type   Expr
	Offset int
}

type FieldList struct {
	Opening token.Pos
	List    []*Field
}

type Ident struct {
	NamePos token.Pos // identifier position
	Name    string
	Obj     *Object
}

type Ellipsis struct {
	Elt      Expr
	Ellipsis token.Pos
}

type BasicLit struct {
	Kind     token.Token // token.INT, token.CHAR, or token.STRING
	Value    string
	ValuePos token.Pos
}

type CompositeLit struct {
	Type   Expr
	Elts   []Expr
	Lbrace token.Pos
}

type KeyValueExpr struct {
	Key   Expr
	Value Expr
}

type ParenExpr struct {
	X      Expr
	Lparen token.Pos
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
	X    Expr
	Star token.Pos
}

type UnaryExpr struct {
	X     Expr
	Op    token.Token
	OpPos token.Pos
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
	Len    Expr
	Elt    Expr
	Lbrack token.Pos
}

type StructType struct {
	Fields *FieldList
	Struct token.Pos
}

type InterfaceType struct {
	Methods   []string
	Interface token.Pos
}

type MapType struct {
	Key   Expr
	Value Expr
	Map   token.Pos
}

type FuncType struct {
	Params  *FieldList
	Results *FieldList
	FPos    token.Pos
}

func (x *Ident) Pos() token.Pos    { return x.NamePos }
func (x *Ellipsis) Pos() token.Pos { return x.Ellipsis }
func (x *BasicLit) Pos() token.Pos { return x.ValuePos }
func (x *CompositeLit) Pos() token.Pos {
	return x.Lbrace
}
func (x *ParenExpr) Pos() token.Pos      { return x.Lparen }
func (x *SelectorExpr) Pos() token.Pos   { return pos(x.X) }
func (x *IndexExpr) Pos() token.Pos      { return pos(x.X) }
func (x *SliceExpr) Pos() token.Pos      { return pos(x.X) }
func (x *TypeAssertExpr) Pos() token.Pos { return pos(x.X) }
func (x *CallExpr) Pos() token.Pos       { return pos(x.Fun) }
func (x *StarExpr) Pos() token.Pos       { return x.Star }
func (x *UnaryExpr) Pos() token.Pos      { return x.OpPos }
func (x *BinaryExpr) Pos() token.Pos     { return pos(x.X) }
func (x *KeyValueExpr) Pos() token.Pos   { return pos(x.Key) }
func (x *ArrayType) Pos() token.Pos      { return x.Lbrack }
func (x *StructType) Pos() token.Pos     { return x.Struct }
func (x *FuncType) Pos() token.Pos {
	return x.FPos
}

func (x *Field) Pos() token.Pos {
	if len(x.Names) > 0 {
		return x.Names[0].Pos()
	}
	return token.NoPos
}

func (x *InterfaceType) Pos() token.Pos { return x.Interface }
func (x *MapType) Pos() token.Pos       { return x.Map }

func pos(expr Expr) token.Pos {
	switch e := expr.(type) {
	case *Ident:
		return e.Pos()
	case *Ellipsis:
		return e.Pos()
	case *BasicLit:
		return e.Pos()
	case *CompositeLit:
		return e.Pos()
	case *ParenExpr:
		return e.Pos()
	case *SelectorExpr:
		return e.Pos()
	case *IndexExpr:
		return e.Pos()
	case *SliceExpr:
		return e.Pos()
	case *TypeAssertExpr:
		return e.Pos()
	case *CallExpr:
		return e.Pos()
	case *StarExpr:
		return e.Pos()
	case *UnaryExpr:
		return e.Pos()
	case *BinaryExpr:
		return e.Pos()
	case *KeyValueExpr:
		return e.Pos()
	case *ArrayType:
		return e.Pos()
	case *MapType:
		return e.Pos()
	case *StructType:
		return e.Pos()
	case *FuncType:
		return e.Pos()
	case *InterfaceType:
		return e.Pos()
	}
	panic(fmt.Sprintf("Unknown type:%T", expr))
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
}

type BranchStmt struct {
	Tok   token.Token
	Label string
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
}

type ForStmt struct {
	Init Stmt
	Cond Expr
	Post Stmt
	Body *BlockStmt
}

type RangeStmt struct {
	Key   Expr
	Value Expr
	X     Expr
	Body  *BlockStmt
	Tok   token.Token
}

type GoStmt struct {
	Call *CallExpr
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
	Name     *Ident
	Assign   token.Pos
	IsAssign bool // isAlias
	Type     Expr
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
	TPos token.Pos
}

func (x *FuncDecl) Pos() token.Pos {
	return x.TPos
}

type File struct {
	Name       *Ident
	Package    token.Pos
	FileStart  token.Pos
	FileEnd    token.Pos
	Imports    []*ImportSpec
	Decls      []Decl
	Unresolved []*Ident
	Scope      *Scope
}

type Scope struct {
	Outer   *Scope
	Objects map[string]*Object
}

func NewScope(outer *Scope) *Scope {
	return &Scope{
		Outer:   outer,
		Objects: make(map[string]*Object),
	}
}

func (s *Scope) Insert(obj *Object) {
	if s == nil {
		panic("s sholud not be nil\n")
	}
	s.Objects[obj.Name] = obj
}

func (s *Scope) Lookup(name string) *Object {
	obj, ok := s.Objects[name]
	if !ok {
		return nil
	}

	return obj
}
