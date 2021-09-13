# Run this on Linux
tmp = /tmp/bbg

.PHONY: all
all: test

.PHONY: test
# test all
test: test0 test1 selfhost  compare-test

$(tmp):
	mkdir -p $(tmp)

t/expected.txt: t/*.go lib/*/*
	export FOO=bar; go run t/*.go myargs > t/expected.txt

$(tmp)/pre: pre/*.go lib/*/* $(tmp)
	go build -o $@ ./pre

$(tmp)/babygo: *.go lib/*/* src/*/* $(tmp)
	go build -o $@ ./

$(tmp)/cross: *.go src/*/* $(tmp)/pre
	rm /tmp/work/*.s
	$(tmp)/pre *.go
	cat src/runtime/runtime.s /tmp/work/*.s > $(@).s
	as -o $(tmp)/a.o $(@).s
	ld -o $@ $(tmp)/a.o

$(tmp)/babygo2: $(tmp)/babygo src/*/*
	rm /tmp/work/*.s
	$(tmp)/babygo *.go
	cat src/runtime/runtime.s /tmp/work/*.s > $(@).s
	as -o $(tmp)/a.o $(@).s
	ld -o $@ $(tmp)/a.o

$(tmp)/pre-test.s: t/*.go src/*/* $(tmp)/pre
	rm /tmp/work/*.s
	$(tmp)/pre t/*.go
	cat src/runtime/runtime.s /tmp/work/*.s > $@

$(tmp)/cross-test.s: t/*.go $(tmp)/cross
	rm /tmp/work/*.s
	$(tmp)/cross t/*.go
	cat src/runtime/runtime.s /tmp/work/*.s > $@

$(tmp)/babygo-test.s: t/*.go src/*/* $(tmp)/babygo
	rm /tmp/work/*.s
	$(tmp)/babygo t/*.go
	cat src/runtime/runtime.s /tmp/work/*.s > $@

$(tmp)/babygo2-test.s: t/*.go $(tmp)/babygo2
	rm /tmp/work/*.s
	$(tmp)/babygo2 t/*.go
	cat src/runtime/runtime.s /tmp/work/*.s > $@

# compare output of test0 and test1
.PHONY: compare-test
compare-test: $(tmp)/pre-test.s $(tmp)/babygo-test.s $(tmp)/babygo2-test.s $(tmp)/cross-test.s
	diff -u $(tmp)/babygo-test.s $(tmp)/pre-test.s
	diff -u $(tmp)/babygo-test.s $(tmp)/cross-test.s
	diff -u $(tmp)/babygo-test.s $(tmp)/babygo2-test.s

.PHONY: test0
test0: $(tmp)/test0 t/expected.txt
	./test.sh $<

.PHONY: test1
test1: $(tmp)/test1 t/expected.txt
	./test.sh $<

$(tmp)/test0: $(tmp)/pre-test.s src/*/*
	as -o $(tmp)/a.o $<
	ld -o $@ $(tmp)/a.o

$(tmp)/test1: $(tmp)/babygo-test.s src/*/*
	as -o $(tmp)/a.o $<
	ld -o $@ $(tmp)/a.o

# test self hosting by comparing 2gen.s and 3gen.s
.PHONY: selfhost
selfhost: $(tmp)/babygo $(tmp)/babygo2
	@echo "testing self host ..."
	rm /tmp/work/*.s
	$(tmp)/babygo2 *.go
	cat src/runtime/runtime.s /tmp/work/*.s > $(tmp)/babygo2-main.s
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
