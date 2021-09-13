# Run this on Linux
tmp = /tmp/bbg

.PHONY: all
all: test

# test all
.PHONY: test
test: $(tmp) test0 test1 selfhost compare-test

$(tmp):
	mkdir -p $(tmp)

t/expected.txt: t/*.go lib/*/*
	export FOO=bar; go run t/*.go myargs > t/expected.txt

$(tmp)/pre: pre/*.go lib/*/* $(tmp)
	go build -o $@ ./pre

$(tmp)/bbg: *.go lib/*/* src/*/* $(tmp)
	go build -o $@ ./

$(tmp)/pre-bbg: $(tmp)/pre *.go src/*/*
	./compile $< $(@).s *.go
	./assemble_and_link $(@).s $@ $(tmp)

$(tmp)/bbg-bbg: $(tmp)/bbg src/*/*
	./compile $< $(@).s *.go
	./assemble_and_link $(@).s $@ $(tmp)

$(tmp)/pre-test.s: $(tmp)/pre t/*.go src/*/*
	./compile $< $(@) t/*.go

$(tmp)/pre-bbg-test.s: $(tmp)/pre-bbg t/*.go
	./compile $< $(@) t/*.go

$(tmp)/bbg-test.s: $(tmp)/bbg t/*.go
	./compile $< $(@) t/*.go

$(tmp)/bbg-bbg-test.s: $(tmp)/bbg-bbg t/*.go
	./compile $< $(@) t/*.go

# compare output of test0 and test1
.PHONY: compare-test
compare-test: $(tmp)/pre-test.s $(tmp)/bbg-test.s $(tmp)/bbg-bbg-test.s $(tmp)/pre-bbg-test.s
	diff -u $(tmp)/pre-test.s $(tmp)/bbg-test.s
	diff -u $(tmp)/bbg-test.s $(tmp)/pre-bbg-test.s
	diff -u $(tmp)/bbg-test.s $(tmp)/bbg-bbg-test.s

.PHONY: test0
test0: $(tmp)/pre-test t/expected.txt
	./test.sh $<

.PHONY: test1
test1: $(tmp)/bbg-test t/expected.txt
	./test.sh $<

$(tmp)/pre-test: $(tmp)/pre-test.s
	./assemble_and_link $< $@ $(tmp)

$(tmp)/bbg-test: $(tmp)/bbg-test.s
	./assemble_and_link $< $@ $(tmp)

# test self hosting by comparing 2gen.s and 3gen.s
.PHONY: selfhost
selfhost: $(tmp)/bbg-bbg
	@echo "testing self host ..."
	./compile $< $(tmp)/bbg-bbg-bbg.s *.go

	diff $(tmp)/bbg-bbg.s $(tmp)/bbg-bbg-bbg.s
	@echo "self host is ok"

.PHONY: fmt
fmt:
	gofmt -w *.go t/*.go pre/*.go src/*/*.go lib/*/*.go

.PHONY: clean
clean:
	rm -f ./tmp/* ./.shared/*
	rm -fr $(tmp)

# to learn the official Go's assembly
.PHONY: sample
sample:
	make sample/sample.s sample/min.s

sample/sample.s: sample/sample.go
	# -N: disable optimizations, -S: print assembly listing -l: no inline
	go tool compile -N -S -l sample/sample.go > sample/sample.s

sample/min.s: sample/min.go
	go tool compile -N -S -l sample/min.go > sample/min.s
