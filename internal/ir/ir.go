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

type ExportedIdent struct {
	PkgName   string
	Ident     *ast.Ident
	Pos       token.Pos
	IsType    bool
	Type      *types.Type // type of the ident, or type itself if ident is type
	MetaIdent *MetaIdent  // for expr
}

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
	Pos      token.Pos
	Type     *types.Type
	Kind     string
	RawValue string // for emitting .data data
	CharVal  int
	IntVal   int
	StrVal   *SLiteral
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

type MetaForeignFuncWrapper struct {
	Pos token.Pos
	QI  QualifiedIdent
	FF  *ForeignFunc
}

type MetaSelectorExpr struct {
	Pos            token.Pos
	IsQI           bool
	QI             QualifiedIdent
	Type           *types.Type
	X              MetaExpr
	SelName        string
	ForeignObjKind string // "var|con|fun"
	ForeignValue   MetaExpr

	// for struct field
	Field     *ast.Field
	Offset    int
	NeedDeref bool
}

// general funcall
type MetaCallExpr struct {
	Pos         token.Pos
	Type        *types.Type   // result type
	Types       []*types.Type // result types when tuple
	ParamTypes  []*types.Type // param types to accept
	Args        []MetaExpr    // args sent from caller
	HasEllipsis bool
	FuncVal     *FuncValue
}

type MetaCallLen struct {
	Pos  token.Pos
	Type *types.Type // result type
	Arg0 MetaExpr
}

type MetaCallCap struct {
	Pos  token.Pos
	Type *types.Type // result type
	Arg0 MetaExpr
}

type MetaCallNew struct {
	Pos      token.Pos
	Type     *types.Type // result type
	TypeArg0 *types.Type
}

type MetaCallMake struct {
	Pos      token.Pos
	Type     *types.Type // result type
	TypeArg0 *types.Type
	Arg1     MetaExpr
	Arg2     MetaExpr
}

type MetaCallAppend struct {
	Pos  token.Pos
	Type *types.Type // result type
	Arg0 MetaExpr
	Arg1 MetaExpr
}

type MetaCallPanic struct {
	Pos  token.Pos
	Type *types.Type // result type
	Arg0 MetaExpr
}

type MetaCallDelete struct {
	Pos  token.Pos
	Type *types.Type // result type
	Arg0 MetaExpr
	Arg1 MetaExpr
}

type MetaConversionExpr struct {
	Pos  token.Pos
	Type *types.Type // To type
	Arg0 MetaExpr
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

type Signature struct {
	ParamTypes  []*types.Type
	ReturnTypes []*types.Type
}

type ForeignFunc struct {
	Symbol      string
	FuncType    *ast.FuncType
	ReturnTypes []*types.Type
	ParamTypes  []*types.Type
}

type Func struct {
	Name      string
	HasBody   bool
	Stmts     []MetaStmt
	Localarea int
	Argsarea  int
	LocalVars []*Variable
	Params    []*Variable
	Retvars   []*Variable
	Method    *Method
	Decl      *ast.FuncDecl
	Signature *Signature
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
type PackageVarConst struct {
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
	Imports        []string
	AstFiles       []*ast.File
	StringLiterals []*SLiteral
	StringIndex    int
	Decls          []ast.Decl
	Fset           *token.FileSet
	FileNoMap      map[string]int // for .loc
}

type AnalyzedPackage struct {
	Path           string
	Name           string
	Imports        []string
	Types          []*types.Type
	Consts         []*PackageVarConst
	Funcs          []*Func
	Vars           []*PackageVarConst
	HasInitFunc    bool
	StringLiterals []*SLiteral
	Fset           *token.FileSet
	FileNoMap      map[string]int // for .loc
}

var RuntimeCmpStringFuncSignature = &Signature{
	ParamTypes:  []*types.Type{types.String, types.String},
	ReturnTypes: []*types.Type{types.Bool},
}

var RuntimeCatStringsSignature = &Signature{
	ParamTypes:  []*types.Type{types.String, types.String},
	ReturnTypes: []*types.Type{types.String},
}

var RuntimeMakeMapSignature = &Signature{
	ParamTypes:  []*types.Type{types.Uintptr, types.Uintptr},
	ReturnTypes: []*types.Type{types.Uintptr},
}

var RuntimeMakeSliceSignature = &Signature{
	ParamTypes:  []*types.Type{types.Int, types.Int, types.Int},
	ReturnTypes: []*types.Type{types.GeneralSliceType},
}

var RuntimeGetAddrForMapGetSignature = &Signature{
	ParamTypes:  []*types.Type{types.Uintptr, types.Eface},
	ReturnTypes: []*types.Type{types.Bool, types.Uintptr},
}

var RuntimeGetAddrForMapSetSignature = &Signature{
	ParamTypes:  []*types.Type{types.Uintptr, types.Eface},
	ReturnTypes: []*types.Type{types.Uintptr},
}

var BuiltinPanicSignature = &Signature{
	ParamTypes:  []*types.Type{types.Eface},
	ReturnTypes: nil,
}
