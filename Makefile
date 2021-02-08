# Run this on Linux
tmp = /tmp/babygo

.PHONY: all
all: test

.PHONY: test
# test all
test: test0 compare-test selfhost

$(tmp):
	mkdir -p $(tmp)

t/expected.txt: t/test.go
	export FOO=bar; go run t/test.go t/another.go myargs > t/expected.txt

$(tmp)/pre: $(tmp) pre/precompiler.go
	go build -o $(tmp)/pre pre/precompiler.go

$(tmp)/cross:  *.go runtime.s $(tmp)/pre
	$(tmp)/pre  *.go > $(tmp)/pre-main.s
	cp $(tmp)/pre-main.s ./.shared/ # for debug
	as -o $(tmp)/cross.o $(tmp)/pre-main.s runtime.s
	ld -e _rt0_amd64_linux -o $(tmp)/cross $(tmp)/cross.o

$(tmp)/babygo: $(tmp)  *.go
	go build -o $(tmp)/babygo  *.go

$(tmp)/babygo2: $(tmp)/babygo runtime.s
	$(tmp)/babygo *.go > $(tmp)/babygo-main.s
	cp $(tmp)/babygo-main.s ./.shared/ # for debug
	as -o $(tmp)/babygo2.o $(tmp)/babygo-main.s runtime.s
	ld -e _rt0_amd64_linux -o $(tmp)/babygo2 $(tmp)/babygo2.o

$(tmp)/pre-test.s: t/test.go $(tmp)/pre runtime.s
	$(tmp)/pre t/test.go t/another.go > $(tmp)/pre-test.s
	cp $(tmp)/pre-test.s ./.shared/

$(tmp)/cross-test.s: t/test.go $(tmp)/cross
	$(tmp)/cross t/test.go t/another.go > $(tmp)/cross-test.s
	cp $(tmp)/cross-test.s ./.shared/

$(tmp)/babygo-test.s: t/test.go $(tmp)/babygo runtime.s
	$(tmp)/babygo t/test.go t/another.go > $(tmp)/babygo-test.s
	cp $(tmp)/babygo-test.s ./.shared/

$(tmp)/babygo2-test.s: t/test.go $(tmp)/babygo2
	$(tmp)/babygo2 t/test.go t/another.go > $(tmp)/babygo2-test.s
	cp $(tmp)/babygo2-test.s ./.shared/

# compare output of test0 and test1
.PHONY: compare-test
compare-test: $(tmp)/pre-test.s $(tmp)/babygo-test.s $(tmp)/babygo2-test.s $(tmp)/cross-test.s
	diff -u $(tmp)/pre-test.s $(tmp)/babygo-test.s
	diff -u $(tmp)/pre-test.s $(tmp)/babygo2-test.s
	diff -u $(tmp)/pre-test.s $(tmp)/cross-test.s

# Run test binaries
$(tmp)/test0: $(tmp)/pre-test.s runtime.s
	as -o $(tmp)/test0.o $(tmp)/pre-test.s runtime.s
	ld -e _rt0_amd64_linux -o $(tmp)/test0 $(tmp)/test0.o

.PHONY: test0
test0: $(tmp)/test0 t/expected.txt
	./test.sh $(tmp)/test0

# test self hosting by comparing 2gen.s and 3gen.s
.PHONY: selfhost
selfhost: $(tmp)/babygo $(tmp)/babygo2 $(tmp)/babygo-main.s
	@echo "testing self host ..."
	$(tmp)/babygo2   *.go > $(tmp)/babygo2-main.s
	diff $(tmp)/babygo-main.s $(tmp)/babygo2-main.s
	@echo "self host is ok"

.PHONY: fmt
fmt: *.go t/*.go pre/*.go
	gofmt -w *.go t/*.go pre/*.go

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
	# -N: disable optimizations, -S: print assembly listing
	go tool compile -N -S sample/sample.go > sample/sample.s

sample/min.s: sample/min.go
	go tool compile -N -S sample/min.go > sample/min.s
