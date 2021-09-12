# Run this on Linux
tmp = /tmp/babygo

.PHONY: all
all: test

.PHONY: test
# test all
test: test0 test1 testcross selfhost  compare-test

$(tmp):
	mkdir -p $(tmp)

t/expected.txt: t/test.go lib/*/*
	export FOO=bar; go run t/test.go t/another.go myargs > t/expected.txt

$(tmp)/pre: $(tmp) pre/precompiler.go lib/*/*
	go build -o $(tmp)/pre ./pre

$(tmp)/cross: *.go src/*/* $(tmp)/pre
	$(tmp)/pre *.go && mv /tmp/a.s $(tmp)/pre-main.s
	cp $(tmp)/pre-main.s ./.shared/ # for debug
	as -o $(tmp)/cross.o $(tmp)/pre-main.s src/runtime/runtime.s
	ld -o $(tmp)/cross $(tmp)/cross.o

$(tmp)/babygo: $(tmp)  *.go lib/*/*
	go build -o $(tmp)/babygo .

$(tmp)/babygo2: $(tmp)/babygo src/*/*
	$(tmp)/babygo *.go && mv /tmp/a.s $(tmp)/babygo-main.s
	cp $(tmp)/babygo-main.s ./.shared/ # for debug
	as -o $(tmp)/babygo2.o $(tmp)/babygo-main.s src/runtime/runtime.s
	ld -o $(tmp)/babygo2 $(tmp)/babygo2.o

$(tmp)/pre-test.s: t/test.go src/*/* $(tmp)/pre
	$(tmp)/pre t/test.go t/another.go && mv /tmp/a.s $(tmp)/pre-test.s
	cp $(tmp)/pre-test.s ./.shared/

$(tmp)/cross-test.s: t/test.go $(tmp)/cross
	$(tmp)/cross t/test.go t/another.go && mv /tmp/a.s $(tmp)/cross-test.s
	cp $(tmp)/cross-test.s ./.shared/

$(tmp)/babygo-test.s: t/test.go src/*/* $(tmp)/babygo
	$(tmp)/babygo t/test.go t/another.go && mv /tmp/a.s $(tmp)/babygo-test.s
	cp $(tmp)/babygo-test.s ./.shared/

$(tmp)/babygo2-test.s: t/test.go $(tmp)/babygo2
	$(tmp)/babygo2 t/test.go t/another.go && mv /tmp/a.s $(tmp)/babygo2-test.s
	cp $(tmp)/babygo2-test.s ./.shared/

# compare output of test0 and test1
.PHONY: compare-test
compare-test: $(tmp)/pre-test.s $(tmp)/babygo-test.s $(tmp)/babygo2-test.s $(tmp)/cross-test.s
	diff -u $(tmp)/pre-test.s $(tmp)/babygo-test.s
	diff -u $(tmp)/pre-test.s $(tmp)/babygo2-test.s
	diff -u $(tmp)/pre-test.s $(tmp)/cross-test.s

$(tmp)/test0: $(tmp)/pre-test.s src/*/*
	as -o $(tmp)/test0.o $(tmp)/pre-test.s src/runtime/runtime.s
	ld -o $(tmp)/test0 $(tmp)/test0.o

.PHONY: test0
test0: $(tmp)/test0 t/expected.txt
	./test.sh $(tmp)/test0

$(tmp)/test1: $(tmp)/babygo-test.s src/*/*
	as -o $(tmp)/test1.o $(tmp)/babygo-test.s src/runtime/runtime.s
	ld -o $(tmp)/test1 $(tmp)/test1.o

.PHONY: test1
test1: $(tmp)/test1 t/expected.txt
	./test.sh $(tmp)/test1

$(tmp)/testcross: $(tmp)/cross-test.s src/*/*
	as -o $(tmp)/testcross.o $(tmp)/cross-test.s src/runtime/runtime.s
	ld -o $(tmp)/testcross $(tmp)/testcross.o

.PHONY: testcross
testcross: $(tmp)/testcross t/expected.txt
	./test.sh $(tmp)/testcross

# test self hosting by comparing 2gen.s and 3gen.s
.PHONY: selfhost
selfhost: $(tmp)/babygo $(tmp)/babygo2 $(tmp)/babygo-main.s
	@echo "testing self host ..."
	$(tmp)/babygo2   *.go && mv /tmp/a.s $(tmp)/babygo2-main.s
	diff $(tmp)/babygo-main.s $(tmp)/babygo2-main.s
	@echo "self host is ok"

.PHONY: fmt
fmt:
	gofmt -w *.go t/*.go pre/*.go src/*/*.go lib/*/*.go

.PHONY: clean
clean:
	rm -f babygo*
	rm -f ./tmp/* ./.shared/*
	rm -fr $(tmp)
	rm -f precompiler babygo babygo2


# to learn the official Go's assembly
.PHONY: sample
sample:
	make sample/sample.s sample/min.s

sample/sample.s: sample/sample.go
	# -N: disable optimizations, -S: print assembly listing -l: no inline
	go tool compile -N -S -l sample/sample.go > sample/sample.s

sample/min.s: sample/min.go
	go tool compile -N -S -l sample/min.go > sample/min.s
