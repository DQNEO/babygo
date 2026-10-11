package sema_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DQNEO/babygo/internal/ir"
	"github.com/DQNEO/babygo/internal/sema"
	"github.com/DQNEO/babygo/internal/types2"
	"github.com/DQNEO/babygo/internal/universe"
	"github.com/DQNEO/babygo/lib/ast"
	"github.com/DQNEO/babygo/lib/parser"
	"github.com/DQNEO/babygo/lib/token"
)

func parseAndWalk(t *testing.T, code string) *ir.AnalyzedPackage {
	t.Helper()
	dir := t.TempDir()
	file := filepath.Join(dir, "main.go")
	if err := os.WriteFile(file, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	astFile, parserErr := parser.ParseFile(fset, file, nil, 0)
	if parserErr != nil {
		t.Fatalf("parser error: %v", parserErr)
	}

	uni := universe.CreateUniverse()
	pkgScope := ast.NewScope(uni)
	for name, obj := range astFile.Scope.Objects {
		pkgScope.Objects[name] = obj
	}

	for _, ident := range astFile.Unresolved {
		obj := pkgScope.Lookup(ident.Name)
		if obj != nil {
			ident.Obj = obj
		} else {
			obj := uni.Lookup(ident.Name)
			if obj != nil {
				ident.Obj = obj
			}
		}
	}

	pkgContainer := &ir.PkgContainer{
		Name:     "main",
		Path:     "main",
		Fset:     fset,
		AstFiles: []*ast.File{astFile},
		Decls:    astFile.Decls,
	}

	return sema.Walk(pkgContainer)
}

func TestDeferMethodCallLowering(t *testing.T) {
	code := `package main

type Worker struct {
	id int
}

func (w *Worker) Finish() {}

func DoWork() {
	var w *Worker
	defer w.Finish()
}
`
	analyzed := parseAndWalk(t, code)
	if analyzed == nil {
		t.Fatal("sema.Walk returned nil")
	}

	var doWorkFn *ir.Func
	for _, fn := range analyzed.Funcs {
		if fn.Name == "DoWork" {
			doWorkFn = fn
			break
		}
	}
	if doWorkFn == nil {
		t.Fatal("DoWork function not found")
	}

	if !doWorkFn.HasDefer {
		t.Fatal("expected doWorkFn.HasDefer to be true")
	}

	if len(doWorkFn.Defers) != 1 {
		t.Fatalf("expected 1 defer, got %d", len(doWorkFn.Defers))
	}

	def := doWorkFn.Defers[0]
	if def.ReceiverAssign == nil {
		t.Fatal("expected def.ReceiverAssign to be non-nil for method defer")
	}

	if def.DeferCall == nil {
		t.Fatal("expected def.DeferCall to be non-nil for method defer")
	}

	if def.DeferCall.FuncVal == nil || def.DeferCall.FuncVal.MethodName != "Finish" {
		t.Fatalf("expected DeferCall MethodName to be 'Finish', got %v", def.DeferCall.FuncVal)
	}

	if len(def.DeferCall.Args) != 1 {
		t.Fatalf("expected 1 argument (receiver), got %d", len(def.DeferCall.Args))
	}

	lhsVar, ok := def.ReceiverAssign.Lhs.(*ir.Variable)
	if !ok {
		t.Fatalf("expected ReceiverAssign.Lhs to be *ir.Variable, got %T", def.ReceiverAssign.Lhs)
	}

	if types2.Kind(lhsVar.Type) != types2.T_POINTER {
		t.Fatalf("expected receiver variable type to be pointer, got kind %v", types2.Kind(lhsVar.Type))
	}
}

func TestDeferValueMethodCallLowering(t *testing.T) {
	code := `package main

type Worker struct {
	id int
}

func (w Worker) Finish() {}

func DoWork() {
	var w Worker
	defer w.Finish()
}
`
	analyzed := parseAndWalk(t, code)
	var doWorkFn *ir.Func
	for _, fn := range analyzed.Funcs {
		if fn.Name == "DoWork" {
			doWorkFn = fn
			break
		}
	}
	if doWorkFn == nil {
		t.Fatal("DoWork function not found")
	}

	if !doWorkFn.HasDefer {
		t.Fatal("expected doWorkFn.HasDefer to be true")
	}

	if len(doWorkFn.Defers) != 1 {
		t.Fatalf("expected 1 defer, got %d", len(doWorkFn.Defers))
	}

	def := doWorkFn.Defers[0]
	if def.ReceiverAssign == nil {
		t.Fatal("expected def.ReceiverAssign to be non-nil")
	}
	if def.DeferCall == nil {
		t.Fatal("expected def.DeferCall to be non-nil")
	}

	lhsVar, ok := def.ReceiverAssign.Lhs.(*ir.Variable)
	if !ok {
		t.Fatalf("expected ReceiverAssign.Lhs to be *ir.Variable, got %T", def.ReceiverAssign.Lhs)
	}
	if types2.Kind(lhsVar.Type) != types2.T_STRUCT {
		t.Fatalf("expected receiver variable type to be struct, got %v", types2.Kind(lhsVar.Type))
	}
}
