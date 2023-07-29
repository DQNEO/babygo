# Babygo, a go compiler made from scratch

![Test](https://github.com/DQNEO/babygo/workflows/Test/badge.svg)

Babygo is a small and simple go compiler. (Smallest and simplest in the world, I believe.)
It is made from scratch and can compile itself.

* No dependency to any libraries. Standard libraries and calling of system calls are home made.
* Lexer, parser and code generator are handwritten.
* Emit assemble code which results in a single static binary.

It depends only on `as` as an assembler and `ld` as a linker.

# Design

## Lexer, Parser and AST
The design and logic of ast, lexer and parser are borrowed (or should I say "stolen")  from `go/ast`, `go/scanner` and `go/parser`.

## Code generator
The design of code generator is borrowed from [chibicc](https://github.com/rui314/chibicc) , a C compiler.

## Remaining parts (Semantic analysis, Type management etc.)
This is purely my design :)

# Environment

It supports x86-64 Linux only.

If you are not using Linux, you can use [a dedicated docker image](https://hub.docker.com/r/dqneo/ubuntu-compiler-go) for this project.

```termiinal
$ docker pull dqneo/ubuntu-compiler-go
$ ./docker-run
```

# Usage

**Currently we are changing CLI design. This section will be updated later**

## Hello world

```terminal
# Build babygo
$ go build -o babygo

# Build hello world by babygo
$ ./go-build -o hello -c ./babygo ./example/hello

# Run hello world
$ ./hello
hello world!
```

## How to do self hosting

```terminal
# Build babygo (1st generation)
$ go build -o babygo

# Build babygo (2nd generation) by babygo 1gen
$ ./go-build -o ./babygo2 -c ./babygo ./

# Build babygo (3rd generation) by babygo 2gen
$ ./go-build -o ./babygo3 -c ./babygo2 ./
```

## Test

```terminal
$ make test
```

# Reference

* https://golang.org/ref/spec The Go Programming Language Specification
* https://golang.org/pkg/go/parser/ go/parser
* https://sourceware.org/binutils/docs/as/ GNU assembler
* https://github.com/rui314/chibicc C compiler
* https://github.com/DQNEO/minigo Go compiler (my previous work)


# License

MIT

# Author

[@DQNEO](https://twitter.com/DQNEO)
