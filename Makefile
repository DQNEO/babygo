# Run this on Linux

all: a.out

a.s: main.go t/source.go
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


clean:
	rm -f a.s a.o a.out

