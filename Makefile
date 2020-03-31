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

t/source: t/source.go
	go build -o t/source t/source.go

t/expected.2: t/source
	t/source 2> t/expected.2 || echo 0

expect:
	make t/expected.2

clean:
	rm -f a.s a.o a.out

