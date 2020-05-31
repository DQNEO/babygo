# Run this on Linux

.PHONEY: test2 test

all: test

test.s: main.go runtime.go t1/test.go
	ln -sf ../t1/test.go t/source.go
	go run main.go > /tmp/test.s && cp /tmp/test.s test.s

test.o: test.s runtime.s
	as -o test.o test.s runtime.s

test.out: test.o
	ld -o test.out test.o

test: test.out t1/expected.1
	./test.sh

# to learn Go's assembly
sample/sample.s: sample/sample.go
	go tool compile -N -S sample/sample.go > sample/sample.s

t1/test: t1/test.go
	go build -o t1/test t1/test.go

t1/expected.1: t1/test
	t1/test 1> t1/expected.1

# 2nd gen compiler
2gen.s: main.go runtime.go runtime.s 2gen/2gen.go
	ln -sf ../2gen/2gen.go t/source.go
	go run main.go > /tmp/2gen.s && cp /tmp/2gen.s 2gen.s

2gen.o: 2gen.s
	as -o 2gen.o 2gen.s runtime.s

self: 2gen.o
	ld -o self 2gen.o

test2: self
	./self > 2gen_out.s
	go run 2gen/2gen.go > /tmp/2gen_out.s
	diff 2gen_out.s /tmp/2gen_out.s && echo 'ok'

test-all:
	make test test2

fmt: *.go t1/*.go 2gen/*.go
	gofmt -w *.go t1/*.go 2gen/*.go

clean:
	rm -f test.s test.o test.out t1/test
	rm -f self 2gen.o 2gen.s

