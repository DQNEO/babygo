# Babygo, a go compiler made from scratch

![Test](https://github.com/DQNEO/babygo/workflows/Test/badge.svg)

Babygo is a small and simple go compiler. (Smallest and simplest in the world, I believe.)
It is made from scratch and can compile itself.

* No dependency to any libraries. Standard libraries and calling of system calls are home made.
* Lexer, parser and code generator are hand written.
* Emit assemble code which resutls in a single static binary.

It depends only on `as` as an assembler and `ld` as a linker.

It is composed of only a few files.

* main.go - the main compiler
* parser.go - parser
* scanner.go - scanner(or lexer)
* src/ - standard packages
* lib/ - libraries

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

## Hello world

```terminal
# Build babygo
$ go build -o babygo

# Compile the hello world program by babygo
$ ./babygo example/hello.go

# Assemble and link
$ as -o hello.o /tmp/*.s
$ ld -o hello hello.o

# Run hello world
$ ./hello
hello world!
```

## How to do self hosting

```terminal
# Build babygo (1st generation)
$ go build -o babygo

# Build babygo by babygo (2nd generation)
$ rm /tmp/*.s
$ ./babygo *.go
$ as -o babygo2.o /tmp/*.s
$ ld -o babygo2 babygo2.o # 2nd generation compiler

# You can generate babygo3 (3rd generation), babygo4, and so on...
$ rm /tmp/*.s
$ ./babygo2 *.go
$ as -o babygo3.o /tmp/*.s
$ ld -o babygo3 babygo3.o # 3rd generation compiler
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
