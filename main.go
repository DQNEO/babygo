package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"strconv"
)

func emitExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		val := e.Value
		ival, _ := strconv.Atoi(val)
		fmt.Printf("# %T\n", expr)
		fmt.Printf("  movq $%d, %%rax\n", ival)
		fmt.Printf("  pushq %%rax\n")
	default:
		panic(fmt.Sprintf("Unexpected expr type %T", expr))
	}
}

func main() {
	source := "42"
	expr, err := parser.ParseExpr(source)
	if err != nil {
		panic(err)
	}
	fmt.Printf(".text\n")
	fmt.Printf(".global main.main\n")
	fmt.Printf("main.main:\n")
	emitExpr(expr)
	fmt.Printf("  popq %%rax\n")
	fmt.Printf("  pushq %%rax\n")
	fmt.Printf("  callq os.Exit\n")
	fmt.Printf("  ret\n")
}
