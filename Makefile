# Run this on Linux



precompiler: precompiler.go runtime.go
	go build -o precompiler precompiler.go

test.s: precompiler.go runtime.go t1/test.go
	ln -sf ../t1/test.go t/source.go
	go run precompiler.go > /tmp/test.s && cp /tmp/test.s test.s

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
2gen.s: precompiler.go runtime.go runtime.s 2gen/main.go 2gen/2gentest.go
	ln -sf ../2gen/main.go t/source.go
	go run precompiler.go > /tmp/2gen.s && cp /tmp/2gen.s 2gen.s

2gen.o: 2gen.s
	as -o 2gen.o 2gen.s runtime.s

self: 2gen.o
	ld -o self 2gen.o

test2: self
	ln -sf 2gentest.go 2gen/input.go
	./self | sed -e '/^#/d' > 2gen_out.s
	go run 2gen/main.go |  sed -e '/^#/d' > /tmp/2gen_out.s
	diff 2gen_out.s /tmp/2gen_out.s >/dev/null && echo 'ok'

test3: self 2gen/2gentest.go
	ln -sf 2gentest.go 2gen/input.go
	./self | as -o a.o runtime.s - && ld a.o
	./a.out > /tmp/a.txt
	go run 2gen/2gentest.go > /tmp/b.txt
	diff /tmp/a.txt /tmp/b.txt && echo 'ok'

test4: self
	ln -sf main.go 2gen/input.go
	./self > /tmp/a.s && cp /tmp/a.s .
	diff 2gen.s  /tmp/a.s >/dev/null

test-all:
	make test test2 test3

fmt: *.go t1/*.go 2gen/*.go
	gofmt -w *.go t1/*.go 2gen/*.go

clean:
	rm -f test.s test.o test.out t1/test
	rm -f self 2gen.o 2gen.s

