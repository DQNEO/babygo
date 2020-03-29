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
		fmt.Printf("# start %T\n", e)
		val := e.Value
		ival, _ := strconv.Atoi(val)
		fmt.Printf("  movq $%d, %%rax\n", ival)
		fmt.Printf("  pushq %%rax\n")
		fmt.Printf("# end %T\n", e)
	case *ast.BinaryExpr:
		fmt.Printf("# start %T\n", e)
		emitExpr(e.X)
		emitExpr(e.Y)
		if e.Op.String() == "+" {
			fmt.Printf("  popq %%rax\n")
			fmt.Printf("  popq %%rdi\n")
			fmt.Printf("  addq %%rdi, %%rax\n")
			fmt.Printf("  pushq %%rax\n")
		} else {
			panic(fmt.Sprintf("Unexpected binary operator %s", e.Op))
		}
		fmt.Printf("# end %T\n", e)
	default:
		panic(fmt.Sprintf("Unexpected expr type %T", expr))
	}
}

func main() {
	source := "40 + 2"
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
