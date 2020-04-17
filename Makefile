# Run this on Linux

.PHONEY: test2 test

all: test

test.s: main.go runtime.go t1/test.go
	ln -sf ../t1/test.go t/source.go
	go run main.go > test.s

test.o: test.s runtime.s
	as -o test.o test.s runtime.s

test.out: test.o
	ld -o test.out test.o

test: test.out t/expected.1
	./test.sh

# to learn Go's assembly
sample/sample.s: sample/sample.go
	go tool compile -N -S sample/sample.go > sample/sample.s

t1/test: t1/test.go
	go build -o t1/test t1/test.go

t/expected.1: t1/test
	t1/test 1> t/expected.1

# 2gen compiler
2gen.s: main.go runtime.go t2/self.go
	ln -sf ../t2/self.go t/source.go
	go run main.go > 2gen.s

2gen.o: 2gen.s
	as -o 2gen.o 2gen.s runtime.s

2gen: 2gen.o
	ld -o 2gen 2gen.o

test2: 2gen
	./2gen && echo 2gen ok

clean:
	rm -f test.s test.o test.out t1/test
	rm -f 2gen 2gen.o 2gen.s

