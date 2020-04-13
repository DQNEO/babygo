# Run this on Linux

.PHONEY: test2 test

all: test

test.s: main.go runtime.go t/test.go
	go run main.go > test.s

test.o: test.s runtime.s
	as -o test.o test.s runtime.s

test.out: test.o
	ld -o test.out test.o

test: test.out t/expected.1 t/test.go
	./test.sh

# to learn Go's assembly
sample/sample.s: sample/sample.go
	go tool compile -N -S sample/sample.go > sample/sample.s

t/test: t/test.go
	go build -o t/test t/test.go

t/expected.1: t/test
	t/test 1> t/expected.1

# 2gen compiler
#2gen.s: test.out
#	./test.out > 2gen.s

#2gen.o: 2gen.s
#	as -o 2gen.o 2gen.s runtime.s

#2gen: 2gen.o
#	ld -o 2gen 2gen.o

#test2: 2gen
#	./2gen && echo 2gen ok

clean:
	rm -f test.s test.o test.out t/test
	rm -f 2gen 2gen.o 2gen.s

