# Run this on Linux
tmp = /tmp/bbg
work = /tmp/work

.PHONY: all
all: test

# test all
.PHONY: test
test: $(tmp) $(work) test0 test1 selfhost compare-test

$(tmp):
	mkdir -p $(tmp)

$(work):
	mkdir -p $(work)

t/expected.txt: t/*.go lib/*/*
	export FOO=bar; go run t/*.go myargs > t/expected.txt

$(tmp)/pre: pre/*.go lib/*/* $(tmp)
	go build -o $@ ./pre

$(tmp)/bbg: *.go lib/*/* src/*/* $(tmp)
	go build -o $@ ./

$(tmp)/pre-bbg: $(tmp)/pre *.go src/*/*
	rm /tmp/work/*.s
	$< *.go
	cat /tmp/work/*.s > $(@).s
	as -o $(tmp)/a.o $(@).s
	ld -o $@ $(tmp)/a.o

$(tmp)/bbg-bbg: $(tmp)/bbg src/*/*
	rm /tmp/work/*.s
	$< *.go
	cat /tmp/work/*.s > $(@).s
	as -o $(tmp)/a.o $(@).s
	ld -o $@ $(tmp)/a.o

$(tmp)/pre-test.s: $(tmp)/pre t/*.go src/*/*
	rm /tmp/work/*.s
	$< t/*.go
	cat /tmp/work/*.s > $@

$(tmp)/pre-bbg-test.s: $(tmp)/pre-bbg t/*.go
	rm /tmp/work/*.s
	$< t/*.go
	cat /tmp/work/*.s > $@

$(tmp)/bbg-test.s: $(tmp)/bbg t/*.go
	rm /tmp/work/*.s
	$< t/*.go
	cat /tmp/work/*.s > $@

$(tmp)/bbg-bbg-test.s: $(tmp)/bbg-bbg t/*.go
	rm /tmp/work/*.s
	$< t/*.go
	cat /tmp/work/*.s > $@

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
	as -o $(tmp)/a.o $<
	ld -o $@ $(tmp)/a.o

$(tmp)/bbg-test: $(tmp)/bbg-test.s
	as -o $(tmp)/a.o $<
	ld -o $@ $(tmp)/a.o

# test self hosting by comparing 2gen.s and 3gen.s
.PHONY: selfhost
selfhost: $(tmp)/bbg-bbg
	@echo "testing self host ..."
	rm /tmp/work/*.s
	$< *.go
	cat /tmp/work/*.s > $(tmp)/bbg-bbg-bbg.s
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
