# Run this on Linux
tmp = /tmp/babygo

.PHONY: all
all: babygo2

.PHONY: test
test: test0 test1 test2

$(tmp):
	mkdir -p $(tmp)

t/expected.txt: t/test.go
	go run t/test.go > t/expected.txt

precompiler: pre/precompiler.go runtime.go $(tmp)
	go build -o $(tmp)/precompiler pre/precompiler.go && cp $(tmp)/precompiler .

./tmp/precompiler_test: precompiler t/test.go
	./precompiler < t/test.go > $(tmp)/precompiler_test.s
	cp $(tmp)/precompiler_test.s ./tmp/ # for debug
	as -o $(tmp)/precompiler_test.o $(tmp)/precompiler_test.s runtime.s
	ld -o ./tmp/precompiler_test $(tmp)/precompiler_test.o

.PHONY: test0
test0: ./tmp/precompiler_test t/expected.txt
	./test.sh ./tmp/precompiler_test

babygo: main.go runtime.go runtime.s precompiler
	./precompiler < main.go > $(tmp)/babygo.s
	cp $(tmp)/babygo.s ./tmp/ # for debug
	as -o $(tmp)/babygo.o $(tmp)/babygo.s runtime.s
	ld -o $(tmp)/babygo $(tmp)/babygo.o
	cp $(tmp)/babygo babygo

.PHONY: test1
test1:	babygo t/test.go
	./babygo < t/test.go > $(tmp)/test.s
	cp $(tmp)/test.s ./tmp/ # for debug
	as -o $(tmp)/test.o $(tmp)/test.s runtime.s
	ld -o $(tmp)/test $(tmp)/test.o
	./test.sh $(tmp)/test

babygo2: babygo
	./babygo < main.go > $(tmp)/2gen.s
	diff $(tmp)/babygo.s $(tmp)/2gen.s
	cp $(tmp)/2gen.s ./tmp/ # for debug
	as -o $(tmp)/2gen.o $(tmp)/2gen.s runtime.s
	ld -o $(tmp)/2gen $(tmp)/2gen.o
	cp $(tmp)/2gen babygo2

.PHONY: test2
test2: babygo2
	./babygo2 < t/test.go > $(tmp)/test2.s
	as -o $(tmp)/test2.o $(tmp)/test2.s runtime.s
	ld -o $(tmp)/test2 $(tmp)/test2.o
	./test.sh $(tmp)/test2

# test self hosting in a more direct way
.PHONY: test3
test3: $(tmp)
	go build -o $(tmp)/bbg main.go
	$(tmp)/bbg < main.go > $(tmp)/bbg2.s
	as -o $(tmp)/bbg2.o $(tmp)/bbg2.s runtime.s
	ld -o $(tmp)/bbg2 $(tmp)/bbg2.o
	$(tmp)/bbg2 < main.go > $(tmp)/bbg3.s
	diff $(tmp)/bbg2.s $(tmp)/bbg3.s

.PHONY: fmt
fmt: *.go t/*.go
	gofmt -w *.go t/*.go

.PHONY: clean
clean:
	rm -f ./tmp/*
	rm -fr $(tmp)
	rm -f precompiler babygo babygo2


# to learn the official Go's assembly
sample/sample.s: sample/sample.go
	go tool compile -N -S sample/sample.go > sample/sample.s
