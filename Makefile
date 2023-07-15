# Run this on a docker container
tmp ?= /tmp/bbg

.PHONY: all
all: test

# test all
.PHONY: test
test: $(tmp)  test1 test2 selfhost test0 compare-test

$(tmp):
	mkdir -p $(tmp)

t/expected.txt: t/*.go lib/*/*
	export FOO=bar; go run t/*.go myargs > t/expected.txt

$(tmp)/pre: pre/*.go lib/*/* $(tmp)
	go build -o $@ ./pre

$(tmp)/bbg: *.go lib/*/* src/*/* $(tmp)
	go build -o $@ .

$(tmp)/pre-test.d: $(tmp)/pre t/*.go src/*/* lib/*/*
	./compile $< $@ t/*.go

$(tmp)/bbg-test.d: $(tmp)/bbg t/*.go
	./compile $< $@ t/*.go

$(tmp)/bbg-bbg-test.d: $(tmp)/bbg-bbg t/*.go
	./compile $< $@ t/*.go

$(tmp)/pre-bbg.d: $(tmp)/pre *.go src/*/*
	./compile $< $(@) *.go

$(tmp)/pre-bbg: $(tmp)/pre-bbg.d
	./assemble_and_link $@ $<

$(tmp)/pre-bbg-test.d: $(tmp)/pre-bbg t/*.go
	./compile $< $@ t/*.go

$(tmp)/bbg-bbg.d: $(tmp)/bbg
	./compile $< $(@) *.go

$(tmp)/bbg-bbg: $(tmp)/bbg-bbg.d
	./assemble_and_link $@ $<

$(tmp)/pre-test: $(tmp)/pre-test.d
	./assemble_and_link $@ $<

$(tmp)/bbg-test: $(tmp)/bbg-test.d
	./assemble_and_link $@ $<

$(tmp)/bbg-bbg-test: $(tmp)/bbg-bbg-test.d
	./assemble_and_link $@ $<

# compare output of test0 and test1
.PHONY: compare-test
compare-test: $(tmp)/pre-test.d $(tmp)/bbg-test.d $(tmp)/bbg-bbg-test.d $(tmp)/pre-bbg-test.d
	diff -u $(tmp)/pre-test.d/all $(tmp)/bbg-test.d/all
	diff -u $(tmp)/bbg-test.d/all $(tmp)/pre-bbg-test.d/all
	diff -u $(tmp)/bbg-test.d/all $(tmp)/bbg-bbg-test.d/all

.PHONY: test0
test0: $(tmp)/pre-test t/expected.txt
	./test.sh $< $(tmp)

.PHONY: test1
test1: $(tmp)/bbg-test t/expected.txt
	./test.sh $< $(tmp)

.PHONY: test2
test2: $(tmp)/bbg-bbg-test t/expected.txt
	./test.sh $< $(tmp)

$(tmp)/bbg-bbg-bbg.d: $(tmp)/bbg-bbg
	./compile $< $(@) *.go

# test self hosting by comparing 2gen.s and 3gen.s
.PHONY: selfhost
selfhost: $(tmp)/bbg-bbg.d $(tmp)/bbg-bbg-bbg.d
	diff $(tmp)/bbg-bbg.d/all $(tmp)/bbg-bbg-bbg.d/all
	@echo "self host is ok"

.PHONY: fmt
fmt:
	gofmt -w *.go t/*.go pre/*.go src/*/*.go lib/*/*.go

.PHONY: clean
clean:
	rm -f ./tmp/*
	rm -fr $(tmp) ./.shared/*

# to learn the official Go's assembly
.PHONY: example
example:
	make example/example.s example/min.s

example/example.s: example
	# -N: disable optimizations, -S: print assembly listing -l: no inline
	go tool compile -N -S -l example/example.go > example/example.s

example/min.s: example
	go tool compile -N -S -l example/min.go > example/min.s
