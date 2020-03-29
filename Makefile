# Run this on Linux

a.out: runtime.s
	as -o a.o runtime.s && ld -o a.out a.o

test: a.out
	./test.sh

clean:
	rm -f a.o a.out
