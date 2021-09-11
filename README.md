# Babygo, a go compiler made from scratch

![Test](https://github.com/DQNEO/babygo/workflows/Test/badge.svg)

Babygo is a small and simple go compiler. (Smallest and simplest in the world, I believe.)
It is made from scratch and can compile itself.

* No dependency to any libraries. Standard libraries and calling of system calls are home made.
* Lexer, parser and code generator are hand written.
* Emit assemble code which resutls in a single static binary.

It depends only on `as` as an assembler and `ld` as a linker.

It is composed of only a fiew files.

* main.go - the main compiler
* runtime.s - low level of runtime
* src/ - internal packages
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

If you are not using Linux, you can use [a dedicated docker image](https://hub.docker.com/r/dqneo/ubuntu-build-essential/tags) for this project.

```termiinal
$ docker pull dqneo/ubuntu-build-essential:go
$ ./docker-run
```

# Usage

## Hello world

```terminal
# Build babygo
$ go build -o babygo *.go

# Compile the hello world program by babygo
$ ./babygo t/hello.go > /tmp/hello.s

# Assemble and link
$ as -o hello.o /tmp/hello.s runtime.s
$ ld -o hello hello.o

# Run hello world
$ ./hello
hello world!
```

## How to do self hosting

```terminal
# Build babygo (1st generation)
$ go build -o babygo *.go

# Build babygo by babygo (2nd generation)
$ ./babygo *.go > /tmp/babygo2.s
$ as -o babygo2.o /tmp/babygo2.s runtime.s
$ ld -o babygo2 babygo2.o # 2nd generation compiler

# Assert babygo2.s and babygo3.s are exactly same
$ ./babygo2 *.go > /tmp/babygo3.s
$ diff /tmp/babygo2.s /tmp/babygo3.s
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
