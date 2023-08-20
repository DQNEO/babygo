package ir

import (
	"github.com/DQNEO/babygo/internal/types"
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/token"
)

type IfcConversion struct {
	Tpos  token.Pos
	Value MetaExpr
	Type  types.Type // Target Type
}

type MetaStructLiteralElement struct {
	Tpos      token.Pos
	Field     *ast.Field
	FieldType types.Type
	Value     MetaExpr
}

type FuncValue struct {
	IsDirect     bool     // direct or indirect
	Symbol       string   // for direct call
	Expr         MetaExpr // for indirect call
	IfcMethodCal bool
	MethodName   string
	IfcType      types.Type
}

type EvalContext struct {
	MaybeOK bool
	Type    types.Type
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
	Name      string
	Obj       *ast.Object // method owner id
	Pos       token.Pos
	IsType    bool
	Type      types.Type // type of the ident, or type itself if ident is type
	MetaIdent *MetaIdent // for expr
	Func      *Func      // for func
}

type NamedType struct {
	MethodSet map[string]*Method
}

type MetaStmt interface {
	Pos() token.Pos
}

type MetaBlockStmt struct {
	Tpos token.Pos
	List []MetaStmt
}

type MetaExprStmt struct {
	Tpos token.Pos
	X    MetaExpr
}

type MetaVarDecl struct {
	Tpos    token.Pos
	Single  *MetaSingleAssign
	LhsType types.Type
}

type MetaSingleAssign struct {
	Tpos token.Pos
	Lhs  MetaExpr
	Rhs  MetaExpr // can be nil
}

type MetaTupleAssign struct {
	Tpos     token.Pos
	IsOK     bool // OK or funcall
	Lhss     []MetaExpr
	Rhs      MetaExpr
	RhsTypes []types.Type
}

type MetaReturnStmt struct {
	Tpos              token.Pos
	IsTuple           bool
	SingleAssignments []*MetaSingleAssign
	TupleAssign       *MetaTupleAssign
}

type MetaIfStmt struct {
	Tpos token.Pos
	Init MetaStmt
	Cond MetaExpr
	Body *MetaBlockStmt
	Else MetaStmt
}

type MetaForContainer struct {
	Tpos      token.Pos
	LabelPost string // for continue
	LabelExit string // for break
	Outer     *MetaForContainer
	Body      *MetaBlockStmt

	ForRangeStmt *MetaForRangeStmt
	ForStmt      *MetaForForStmt
}

type MetaForForStmt struct {
	Tpos token.Pos
	Init MetaStmt
	Cond MetaExpr
	Post MetaStmt
}

type MetaForRangeStmt struct {
	Tpos         token.Pos
	IsMap        bool
	LenVar       *Variable
	Indexvar     *Variable
	MapVar       *Variable // map
	ItemVar      *Variable // map element
	X            MetaExpr
	Key          MetaExpr
	Value        MetaExpr
	MapVarAssign *MetaSingleAssign
}

type MetaBranchStmt struct {
	Tpos             token.Pos
	ContainerForStmt *MetaForContainer
	ContinueOrBreak  int // 1: continue, 2:break
}

type MetaSwitchStmt struct {
	Tpos  token.Pos
	Init  MetaStmt
	Cases []*MetaCaseClause
	Tag   MetaExpr
}

type MetaCaseClause struct {
	Tpos     token.Pos
	ListMeta []MetaExpr
	Body     []MetaStmt
}

type MetaTypeSwitchStmt struct {
	Tpos            token.Pos
	Subject         MetaExpr
	SubjectVariable *Variable
	AssignObj       *ast.Object
	Cases           []*MetaTypeSwitchCaseClose
}

type MetaTypeSwitchCaseClose struct {
	Tpos     token.Pos
	Variable *Variable
	//VariableType *Type
	Types []types.Type
	Body  []MetaStmt
}

type MetaGoStmt struct {
	Tpos token.Pos
	Fun  MetaExpr
}

type MetaDeferStmt struct {
	Tpos       token.Pos
	Fun        MetaExpr
	FuncAssign *MetaSingleAssign
}

func (s *MetaBlockStmt) Pos() token.Pos           { return s.Tpos }
func (s *MetaExprStmt) Pos() token.Pos            { return s.Tpos }
func (s *MetaVarDecl) Pos() token.Pos             { return s.Tpos }
func (s *MetaSingleAssign) Pos() token.Pos        { return s.Tpos }
func (s *MetaTupleAssign) Pos() token.Pos         { return s.Tpos }
func (s *MetaReturnStmt) Pos() token.Pos          { return s.Tpos }
func (s *MetaIfStmt) Pos() token.Pos              { return s.Tpos }
func (s *MetaForContainer) Pos() token.Pos        { return s.Tpos }
func (s *MetaForForStmt) Pos() token.Pos          { return s.Tpos }
func (s *MetaForRangeStmt) Pos() token.Pos        { return s.Tpos }
func (s *MetaBranchStmt) Pos() token.Pos          { return s.Tpos }
func (s *MetaSwitchStmt) Pos() token.Pos          { return s.Tpos }
func (s *MetaCaseClause) Pos() token.Pos          { return s.Tpos }
func (s *MetaTypeSwitchStmt) Pos() token.Pos      { return s.Tpos }
func (s *MetaTypeSwitchCaseClose) Pos() token.Pos { return s.Tpos }
func (s *MetaGoStmt) Pos() token.Pos              { return s.Tpos }
func (s *MetaDeferStmt) Pos() token.Pos           { return s.Tpos }

type MetaExpr interface {
	Pos() token.Pos
	//	GetType() types.Type
}

type MetaBasicLit struct {
	Tpos     token.Pos
	Type     types.Type
	Kind     string
	RawValue string // for emitting .data data
	CharVal  int
	IntVal   int
	StrVal   *SLiteral
}

type MetaCompositLit struct {
	Tpos token.Pos
	Type types.Type // type of the composite
	Kind string     // "struct", "array", "slice" // @TODO "map"

	// for struct
	StructElements []*MetaStructLiteralElement // for "struct"

	// for array or slice
	Len     int
	ElmType types.Type
	Elms    []MetaExpr
}

type MetaIdent struct {
	Tpos token.Pos
	Type types.Type
	Kind string // "blank|nil|true|false|var|con|fun|typ"
	Name string

	Variable *Variable // for "var"

	Const *Const // for "con"
}

type MetaForeignFuncWrapper struct {
	Tpos token.Pos
	QI   QualifiedIdent
}

type MetaSelectorExpr struct {
	Tpos           token.Pos
	IsQI           bool
	QI             QualifiedIdent
	Type           types.Type
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
	Tpos        token.Pos
	Type        types.Type   // result type
	Types       []types.Type // result types when tuple
	ParamTypes  []types.Type // param types to accept
	Args        []MetaExpr   // args sent from caller
	HasEllipsis bool
	FuncVal     *FuncValue
}

type MetaCallLen struct {
	Tpos token.Pos
	Type types.Type // result type
	Arg0 MetaExpr
}

type MetaCallCap struct {
	Tpos token.Pos
	Type types.Type // result type
	Arg0 MetaExpr
}

type MetaCallNew struct {
	Tpos     token.Pos
	Type     types.Type // result type
	TypeArg0 types.Type
}

type MetaCallMake struct {
	Tpos     token.Pos
	Type     types.Type // result type
	TypeArg0 types.Type
	Arg1     MetaExpr
	Arg2     MetaExpr
}

type MetaCallAppend struct {
	Tpos token.Pos
	Type types.Type // result type
	Arg0 MetaExpr
	Arg1 MetaExpr
}

type MetaCallPanic struct {
	Tpos token.Pos
	Type types.Type // result type
	Arg0 MetaExpr
}

type MetaCallDelete struct {
	Tpos token.Pos
	Type types.Type // result type
	Arg0 MetaExpr
	Arg1 MetaExpr
}

type MetaConversionExpr struct {
	Tpos token.Pos
	Type types.Type // To type
	Arg0 MetaExpr
}

type MetaIndexExpr struct {
	Tpos    token.Pos
	IsMap   bool // mp[k]
	NeedsOK bool // when map, is it ok syntax ?
	Index   MetaExpr
	X       MetaExpr
	Type    types.Type
}

type MetaSliceExpr struct {
	Tpos token.Pos
	Type types.Type
	Low  MetaExpr
	High MetaExpr
	Max  MetaExpr
	X    MetaExpr
}
type MetaStarExpr struct {
	Tpos token.Pos
	Type types.Type
	X    MetaExpr
}
type MetaUnaryExpr struct {
	Tpos token.Pos
	X    MetaExpr
	Type types.Type
	Op   string
}
type MetaBinaryExpr struct {
	Tpos token.Pos
	Type types.Type
	Op   string
	X    MetaExpr
	Y    MetaExpr
}

type MetaTypeAssertExpr struct {
	Tpos    token.Pos
	NeedsOK bool
	X       MetaExpr
	Type    types.Type
}

func (e *MetaBasicLit) Pos() token.Pos             { return e.Tpos }
func (e *MetaCompositLit) Pos() token.Pos          { return e.Tpos }
func (e *MetaIdent) Pos() token.Pos                { return e.Tpos }
func (e *MetaForeignFuncWrapper) Pos() token.Pos   { return e.Tpos }
func (e *MetaSelectorExpr) Pos() token.Pos         { return e.Tpos }
func (e *MetaCallExpr) Pos() token.Pos             { return e.Tpos }
func (e *MetaCallLen) Pos() token.Pos              { return e.Tpos }
func (e *MetaCallCap) Pos() token.Pos              { return e.Tpos }
func (e *MetaCallNew) Pos() token.Pos              { return e.Tpos }
func (e *MetaCallMake) Pos() token.Pos             { return e.Tpos }
func (e *MetaCallAppend) Pos() token.Pos           { return e.Tpos }
func (e *MetaCallPanic) Pos() token.Pos            { return e.Tpos }
func (e *MetaCallDelete) Pos() token.Pos           { return e.Tpos }
func (e *MetaConversionExpr) Pos() token.Pos       { return e.Tpos }
func (e *MetaIndexExpr) Pos() token.Pos            { return e.Tpos }
func (e *MetaSliceExpr) Pos() token.Pos            { return e.Tpos }
func (e *MetaStarExpr) Pos() token.Pos             { return e.Tpos }
func (e *MetaUnaryExpr) Pos() token.Pos            { return e.Tpos }
func (e *MetaBinaryExpr) Pos() token.Pos           { return e.Tpos }
func (e *MetaTypeAssertExpr) Pos() token.Pos       { return e.Tpos }
func (e *IfcConversion) Pos() token.Pos            { return e.Tpos }
func (e *MetaStructLiteralElement) Pos() token.Pos { return e.Tpos }
func (e *Variable) Pos() token.Pos                 { return e.Tpos }
func (e *Const) Pos() token.Pos                    { return e.Tpos }

type Signature struct {
	ParamTypes  []types.Type
	ReturnTypes []types.Type
}

type Func struct {
	PkgName   string
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
	HasDefer  bool
	DeferVar  *Variable
}

type Method struct {
	PkgName      string
	RcvNamedType *ast.Ident
	IsPtrMethod  bool
	Name         string
	FuncType     *ast.FuncType
}

type Variable struct {
	Tpos         token.Pos
	Name         string
	IsGlobal     bool
	GlobalSymbol string
	LocalOffset  int
	Type         types.Type
}

type Const struct {
	Tpos         token.Pos
	Name         string
	IsGlobal     bool
	GlobalSymbol string // "pkg.Foo"
	Literal      *MetaBasicLit
	Type         types.Type
}

// Package vars or consts
type PackageVarConst struct {
	Spec    *ast.ValueSpec
	Name    *ast.Ident
	Val     ast.Expr   // can be nil
	MetaVal MetaExpr   // can be nil
	Type    types.Type // cannot be nil
	MetaVar *MetaIdent // only for var
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
	Types          []types.Type
	Consts         []*PackageVarConst
	Funcs          []*Func
	Vars           []*PackageVarConst
	HasInitFunc    bool
	StringLiterals []*SLiteral
	Fset           *token.FileSet
	FileNoMap      map[string]int // for .loc
}

var RuntimeCmpStringFuncSignature = &Signature{
	ParamTypes:  []types.Type{types.String, types.String},
	ReturnTypes: []types.Type{types.Bool},
}

var RuntimeCatStringsSignature = &Signature{
	ParamTypes:  []types.Type{types.String, types.String},
	ReturnTypes: []types.Type{types.String},
}

var RuntimeMakeMapSignature = &Signature{
	ParamTypes:  []types.Type{types.Uintptr, types.Uintptr},
	ReturnTypes: []types.Type{types.Uintptr},
}

var RuntimeMakeSliceSignature = &Signature{
	ParamTypes:  []types.Type{types.Int, types.Int, types.Int},
	ReturnTypes: []types.Type{types.GGeneralSliceType},
}

var RuntimeGetAddrForMapGetSignature = &Signature{
	ParamTypes:  []types.Type{types.Uintptr, types.EmptyInterface},
	ReturnTypes: []types.Type{types.Bool, types.Uintptr},
}

var RuntimeGetAddrForMapSetSignature = &Signature{
	ParamTypes:  []types.Type{types.Uintptr, types.EmptyInterface},
	ReturnTypes: []types.Type{types.Uintptr},
}

var BuiltinPanicSignature = &Signature{
	ParamTypes:  []types.Type{types.EmptyInterface},
	ReturnTypes: nil,
}
