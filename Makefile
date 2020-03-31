# Run this on Linux

all: a.out

main.s: main.go t/source.go
	go run main.go > main.s

a.o: main.s runtime.s
	as -o a.o main.s runtime.s

a.out: a.o
	ld -o a.out a.o

test: a.out
	./test.sh

# to learn Go's assembly
sample/sample.s: sample/sample.go
	go tool compile -N -S sample/sample.go > sample/sample.s


clean:
	rm -f a.o a.out main.s

