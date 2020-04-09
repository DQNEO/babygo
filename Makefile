# Run this on Linux

.PHONEY: test2 test

all: test

a.s: main.go t/a.go
	go run main.go > a.s

a.o: a.s runtime.s
	as -o a.o a.s runtime.s

a.out: a.o
	ld -o a.out a.o

test: a.out
	./test.sh

# to learn Go's assembly
sample/sample.s: sample/sample.go
	go tool compile -N -S sample/sample.go > sample/sample.s

t/a: t/a.go
	go build -o t/a t/a.go

t/expected.1: t/a
	t/a 1> t/expected.1

expect:
	make t/expected.1

# 2gen compiler
#2gen.s: a.out
#	./a.out > 2gen.s

#2gen.o: 2gen.s
#	as -o 2gen.o 2gen.s runtime.s

#2gen: 2gen.o
#	ld -o 2gen 2gen.o

#test2: 2gen
#	./2gen && echo 2gen ok

clean:
	rm -f a.s a.o a.out
	rm -f 2gen 2gen.o 2gen.s

