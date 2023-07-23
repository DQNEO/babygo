package ir

import (
	"github.com/DQNEO/babygo/internal/types"
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/token"
)

type MetaStructLiteralElement struct {
	Pos       token.Pos
	Field     *ast.Field
	FieldType *types.Type
	ValueMeta MetaExpr
}

type MetaArg struct {
	Meta      MetaExpr
	ParamType *types.Type // expected type
}

type FuncValue struct {
	IsDirect bool     // direct or indirect
	Symbol   string   // for direct call
	Expr     MetaExpr // for indirect call
}

type EvalContext struct {
	MaybeOK bool
	Type    *types.Type
}

// --- walk ---
type SLiteral struct {
	Label  string
	Strlen int
	Value  string // raw value
}

type QualifiedIdent string

type NamedType struct {
	MethodSet map[string]*Method
}

type MetaStmt interface{}

type MetaBlockStmt struct {
	Pos  token.Pos
	List []MetaStmt
}

type MetaExprStmt struct {
	Pos token.Pos
	X   MetaExpr
}

type MetaVarDecl struct {
	Pos     token.Pos
	Single  *MetaSingleAssign
	LhsType *types.Type
}

type MetaSingleAssign struct {
	Pos token.Pos
	Lhs MetaExpr
	Rhs MetaExpr // can be nil
}

type MetaTupleAssign struct {
	Pos      token.Pos
	IsOK     bool // OK or funcall
	Lhss     []MetaExpr
	Rhs      MetaExpr
	RhsTypes []*types.Type
}

type MetaReturnStmt struct {
	Pos     token.Pos
	Fnc     *Func
	Results []MetaExpr
}

type MetaIfStmt struct {
	Pos  token.Pos
	Init MetaStmt
	Cond MetaExpr
	Body *MetaBlockStmt
	Else MetaStmt
}

type MetaForContainer struct {
	Pos       token.Pos
	LabelPost string // for continue
	LabelExit string // for break
	Outer     *MetaForContainer
	Body      *MetaBlockStmt

	ForRangeStmt *MetaForRangeStmt
	ForStmt      *MetaForForStmt
}

type MetaForForStmt struct {
	Pos  token.Pos
	Init MetaStmt
	Cond MetaExpr
	Post MetaStmt
}

type MetaForRangeStmt struct {
	Pos      token.Pos
	IsMap    bool
	LenVar   *Variable
	Indexvar *Variable
	MapVar   *Variable // map
	ItemVar  *Variable // map element
	X        MetaExpr
	Key      MetaExpr
	Value    MetaExpr
}

type MetaBranchStmt struct {
	Pos              token.Pos
	ContainerForStmt *MetaForContainer
	ContinueOrBreak  int // 1: continue, 2:break
}

type MetaSwitchStmt struct {
	Pos   token.Pos
	Init  MetaStmt
	Cases []*MetaCaseClause
	Tag   MetaExpr
}

type MetaCaseClause struct {
	Pos      token.Pos
	ListMeta []MetaExpr
	Body     []MetaStmt
}

type MetaTypeSwitchStmt struct {
	Pos             token.Pos
	Subject         MetaExpr
	SubjectVariable *Variable
	AssignObj       *ast.Object
	Cases           []*MetaTypeSwitchCaseClose
}

type MetaTypeSwitchCaseClose struct {
	Pos      token.Pos
	Variable *Variable
	//VariableType *Type
	Types []*types.Type
	Body  []MetaStmt
}

type MetaGoStmt struct {
	Pos token.Pos
	Fun MetaExpr
}

type MetaExpr interface{}

type MetaBasicLit struct {
	Pos     token.Pos
	Type    *types.Type
	Kind    string
	Value   string
	CharVal int
	IntVal  int
	StrVal  *SLiteral
}

type MetaCompositLit struct {
	Pos  token.Pos
	Type *types.Type // type of the composite
	Kind string      // "struct", "array", "slice" // @TODO "map"

	// for struct
	StructElements []*MetaStructLiteralElement // for "struct"

	// for array or slice
	Len      int
	ElmType  *types.Type
	MetaElms []MetaExpr
}

type MetaIdent struct {
	Pos  token.Pos
	Type *types.Type
	Kind string // "blank|nil|true|false|var|con|fun|typ"
	Name string

	Variable *Variable // for "var"

	Const *Const // for "con"
}

type MetaSelectorExpr struct {
	Pos     token.Pos
	IsQI    bool
	QI      QualifiedIdent
	Type    *types.Type
	X       MetaExpr
	SelName string
}

type MetaCallExpr struct {
	Pos   token.Pos
	Type  *types.Type   // result type
	Types []*types.Type // result types when tuple

	IsConversion bool

	// For Conversion
	ToType *types.Type

	Arg0     MetaExpr // For conversion, len, cap
	TypeArg0 *types.Type
	Arg1     MetaExpr
	Arg2     MetaExpr

	// For funcall
	HasEllipsis bool
	Builtin     *ast.Object

	// general funcall
	ReturnTypes []*types.Type
	FuncVal     *FuncValue
	//receiver ast.Expr
	MetaArgs []*MetaArg
}

type MetaIndexExpr struct {
	Pos     token.Pos
	IsMap   bool // mp[k]
	NeedsOK bool // when map, is it ok syntax ?
	Index   MetaExpr
	X       MetaExpr
	Type    *types.Type
}

type MetaSliceExpr struct {
	Pos  token.Pos
	Type *types.Type
	Low  MetaExpr
	High MetaExpr
	Max  MetaExpr
	X    MetaExpr
}
type MetaStarExpr struct {
	Pos  token.Pos
	Type *types.Type
	X    MetaExpr
}
type MetaUnaryExpr struct {
	Pos  token.Pos
	X    MetaExpr
	Type *types.Type
	Op   string
}
type MetaBinaryExpr struct {
	Pos  token.Pos
	Type *types.Type
	Op   string
	X    MetaExpr
	Y    MetaExpr
}

type MetaTypeAssertExpr struct {
	Pos     token.Pos
	NeedsOK bool
	X       MetaExpr
	Type    *types.Type
}

type ForeignFunc struct {
	Symbol   string
	FuncType *ast.FuncType
}

type Func struct {
	Name      string
	Stmts     []MetaStmt
	Localarea int
	Argsarea  int
	LocalVars []*Variable
	Params    []*Variable
	Retvars   []*Variable
	FuncType  *ast.FuncType
	Method    *Method
}
type Method struct {
	PkgName      string
	RcvNamedType *ast.Ident
	IsPtrMethod  bool
	Name         string
	FuncType     *ast.FuncType
}
type Variable struct {
	Name         string
	IsGlobal     bool
	GlobalSymbol string
	LocalOffset  int
	Typ          *types.Type
}

type Const struct {
	Name         string
	IsGlobal     bool
	GlobalSymbol string // "pkg.Foo"
	Literal      *MetaBasicLit
	Type         *types.Type
}

// Package vars or consts
type PackageVals struct {
	Spec    *ast.ValueSpec
	Name    *ast.Ident
	Val     ast.Expr    // can be nil
	MetaVal MetaExpr    // can be nil
	Type    *types.Type // cannot be nil
	MetaVar *MetaIdent  // only for var
}

type PkgContainer struct {
	Path           string
	Name           string
	AstFiles       []*ast.File
	Vals           []*PackageVals
	Funcs          []*Func
	StringLiterals []*SLiteral
	StringIndex    int
	Decls          []ast.Decl
	Fset           *token.FileSet
	HasInitFunc    bool
	FileNoMap      map[string]int // for .loc
}

type AnalyzedPackage struct {
	Path           string
	Name           string
	Vals           []*PackageVals
	Funcs          []*Func
	StringLiterals []*SLiteral
	Fset           *token.FileSet
	FileNoMap      map[string]int // for .loc
}
