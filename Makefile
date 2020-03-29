# Run this on Linux

a.out: runtime.s main.s
	as -o a.o main.s runtime.s && ld -o a.out a.o

main.s:
	go run main.go > main.s

test: a.out
	./test.sh

clean:
	rm -f a.o a.out
