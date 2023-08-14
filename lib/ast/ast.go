package ast

import (
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
	Pos() token.Pos
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
	Methods   *FieldList
	Interface token.Pos
}

type MapType struct {
	Key   Expr
	Value Expr
	Map   token.Pos
}

type FuncType struct {
	Func    token.Pos
	Params  *FieldList
	Results *FieldList // this can be nil: e.g. func f() {}
}

func (x *Ident) Pos() token.Pos    { return x.NamePos }
func (x *Ellipsis) Pos() token.Pos { return x.Ellipsis }
func (x *BasicLit) Pos() token.Pos { return x.ValuePos }
func (x *CompositeLit) Pos() token.Pos {
	return x.Lbrace
}
func (x *ParenExpr) Pos() token.Pos      { return x.Lparen }
func (x *SelectorExpr) Pos() token.Pos   { return x.X.Pos() }
func (x *IndexExpr) Pos() token.Pos      { return x.X.Pos() }
func (x *SliceExpr) Pos() token.Pos      { return x.X.Pos() }
func (x *TypeAssertExpr) Pos() token.Pos { return x.X.Pos() }
func (x *CallExpr) Pos() token.Pos       { return x.Fun.Pos() }
func (x *StarExpr) Pos() token.Pos       { return x.Star }
func (x *UnaryExpr) Pos() token.Pos      { return x.OpPos }
func (x *BinaryExpr) Pos() token.Pos     { return x.X.Pos() }
func (x *KeyValueExpr) Pos() token.Pos   { return x.Key.Pos() }
func (x *ArrayType) Pos() token.Pos      { return x.Lbrack }
func (x *StructType) Pos() token.Pos     { return x.Struct }
func (x *FuncType) Pos() token.Pos {
	return x.Func
}

func (x *Field) Pos() token.Pos {
	if len(x.Names) > 0 {
		return x.Names[0].Pos()
	}
	return token.NoPos
}

func (x *InterfaceType) Pos() token.Pos { return x.Interface }
func (x *MapType) Pos() token.Pos       { return x.Map }

type Stmt interface {
	Pos() token.Pos
}

type DeclStmt struct {
	Decl Decl
}

type ExprStmt struct {
	X Expr
}

type IncDecStmt struct {
	X      Expr
	TokPos token.Pos
	Tok    token.Token
}

type AssignStmt struct {
	Lhs     []Expr
	TokPos  token.Pos
	Tok     token.Token
	Rhs     []Expr
	IsRange bool
}

type ReturnStmt struct {
	Return  token.Pos
	Results []Expr
}

type BranchStmt struct {
	TokPos token.Pos
	Tok    token.Token
	Label  string
}

type BlockStmt struct {
	Lbrace token.Pos
	List   []Stmt
}

type IfStmt struct {
	If   token.Pos
	Init Stmt
	Cond Expr
	Body *BlockStmt
	Else Stmt
}

type CaseClause struct {
	Case token.Pos
	List []Expr
	Body []Stmt
}

type SwitchStmt struct {
	Switch token.Pos
	Init   Expr
	Tag    Expr
	Body   *BlockStmt
	// lableExit string
}

type TypeSwitchStmt struct {
	Switch token.Pos
	Assign Stmt
	Body   *BlockStmt
}

type ForStmt struct {
	For  token.Pos
	Init Stmt
	Cond Expr
	Post Stmt
	Body *BlockStmt
}

type RangeStmt struct {
	For   token.Pos
	Key   Expr
	Value Expr
	X     Expr
	Body  *BlockStmt
	Tok   token.Token
}

type GoStmt struct {
	Go   token.Pos
	Call *CallExpr
}

type DeferStmt struct {
	Defer token.Pos
	Call  *CallExpr
}

func (s *DeclStmt) Pos() token.Pos { return s.Decl.Pos() }
func (s *ExprStmt) Pos() token.Pos { return s.X.Pos() }

func (s *IncDecStmt) Pos() token.Pos     { return s.X.Pos() }
func (s *AssignStmt) Pos() token.Pos     { return s.Lhs[0].Pos() }
func (s *GoStmt) Pos() token.Pos         { return s.Go }
func (s *DeferStmt) Pos() token.Pos      { return s.Defer }
func (s *ReturnStmt) Pos() token.Pos     { return s.Return }
func (s *BranchStmt) Pos() token.Pos     { return s.TokPos }
func (s *BlockStmt) Pos() token.Pos      { return s.Lbrace }
func (s *IfStmt) Pos() token.Pos         { return s.If }
func (s *CaseClause) Pos() token.Pos     { return s.Case }
func (s *SwitchStmt) Pos() token.Pos     { return s.Switch }
func (s *TypeSwitchStmt) Pos() token.Pos { return s.Switch }
func (s *ForStmt) Pos() token.Pos        { return s.For }
func (s *RangeStmt) Pos() token.Pos      { return s.For }

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
	IsAssign bool // isAlias
	Type     Expr
}

func (s *ImportSpec) Pos() token.Pos {
	return s.Path.Pos()
}

func (s *ValueSpec) Pos() token.Pos { return s.Names[0].Pos() }
func (s *TypeSpec) Pos() token.Pos  { return s.Name.Pos() }

// Pseudo interface for *ast.Decl
// *GenDecl | *FuncDecl
type Decl interface {
	Pos() token.Pos
}

type Spec interface{}

type GenDecl struct {
	TokPos token.Pos
	Specs  []Spec
}

type FuncDecl struct {
	Recv *FieldList
	Name *Ident
	Type *FuncType
	Body *BlockStmt
}

func (x *GenDecl) Pos() token.Pos {
	return x.TokPos
}

func (x *FuncDecl) Pos() token.Pos {
	return x.Type.Pos()
}

type File struct {
	Name       *Ident // package name that is in the package clause
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
