# Run this on Linux
tmp = /tmp/babygo

.PHONY: all
all: test

.PHONY: test
test: test0 test1 test2 test-self-host

$(tmp):
	mkdir -p $(tmp)

t/expected.txt: t/test.go
	go run t/test.go > t/expected.txt

precompiler: pre/precompiler.go runtime.go $(tmp)
	go build -o $(tmp)/precompiler pre/precompiler.go && cp $(tmp)/precompiler .

$(tmp)/precompiler_test: precompiler t/test.go
	./precompiler < t/test.go > $(tmp)/precompiler_test.s
	cp $(tmp)/precompiler_test.s ./.shared/ # for debug
	as -o $(tmp)/precompiler_test.o $(tmp)/precompiler_test.s runtime.s
	ld -o $(tmp)/precompiler_test $(tmp)/precompiler_test.o

.PHONY: test0
test0: $(tmp)/precompiler_test t/expected.txt
	./test.sh $(tmp)/precompiler_test

babygo: main.go
	go build -o babygo main.go

# This target is actually not needed any more. Just kept for a preference
babygo_by_precompiler: main.go runtime.go runtime.s precompiler
	./precompiler < main.go > $(tmp)/babygo.s
	cp $(tmp)/babygo.s ./.shared/ # for debug
	as -o $(tmp)/babygo.o $(tmp)/babygo.s runtime.s
	ld -o babygo_by_precompiler $(tmp)/babygo.o

.PHONY: test1
test1:	babygo t/test.go
	./babygo < t/test.go > $(tmp)/test.s
	cp $(tmp)/test.s ./.shared/ # for debug
	as -o $(tmp)/test.o $(tmp)/test.s runtime.s
	ld -o $(tmp)/test $(tmp)/test.o
	./test.sh $(tmp)/test

babygo2: babygo
	./babygo < main.go > $(tmp)/2gen.s
	cp $(tmp)/2gen.s ./.shared/ # for debug
	as -o $(tmp)/2gen.o $(tmp)/2gen.s runtime.s
	ld -o $(tmp)/2gen $(tmp)/2gen.o
	cp $(tmp)/2gen babygo2

.PHONY: test2
test2: babygo2
	./babygo2 < t/test.go > $(tmp)/test2.s
	as -o $(tmp)/test2.o $(tmp)/test2.s runtime.s
	ld -o $(tmp)/test2 $(tmp)/test2.o
	./test.sh $(tmp)/test2

# test self hosting by comparing 2gen.s and 3gen.s
.PHONY: test-self-host
test-self-host: babygo2
	./babygo2 < main.go > $(tmp)/3gen.s
	diff $(tmp)/2gen.s $(tmp)/3gen.s

.PHONY: fmt
fmt: *.go t/*.go pre/*.go
	gofmt -w *.go t/*.go pre/*.go

.PHONY: clean
clean:
	rm -f ./tmp/* ./.shared/*
	rm -fr $(tmp)
	rm -f precompiler babygo babygo2


# to learn the official Go's assembly
sample/sample.s: sample/sample.go
	go tool compile -N -S sample/sample.go > sample/sample.s
