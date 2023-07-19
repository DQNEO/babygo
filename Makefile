# Run this on a docker container
tmp ?= /tmp/bbg

#.PHONY: all
#all: test

# run all tests
.PHONY: test
test: $(tmp)  test1 test2 selfhost test0 compare-test

$(tmp):
	mkdir -p $(tmp)

# make pre compiler (a rich binary)
$(tmp)/pre: pre/*.go lib/*/* $(tmp)
	go build -o $@ ./pre

# make babygo 1gen compiler (a rich binary)
$(tmp)/bbg: *.go lib/*/* src/*/* $(tmp)
	go build -o $@ .

# Generate asm files for babygo 2gen compiler by pre compiler compiling babygo
$(tmp)/pre-bbg.d: $(tmp)/pre *.go src/*/* lib/*/*
	./compile -o $(@) $< *.go

# Make babygo 2 gen compiler (a thin binary) by pre compiler
$(tmp)/pre-bbg: $(tmp)/pre-bbg.d
	./assemble_and_link $@ $<

# Generate asm files for babygo 2gen compiler by babygo 1gen compiler compiling babygo
$(tmp)/bbg-bbg.d: $(tmp)/bbg
	./compile -o $(@) $< *.go

# Make babygo 2gen compiler (a thin binary)
$(tmp)/bbg-bbg: $(tmp)/bbg-bbg.d
	./assemble_and_link $@ $<

# Generate asm files for babygo 3gen compiler by babygo 2gen compiler compiling babygo
$(tmp)/bbg-bbg-bbg.d: $(tmp)/bbg-bbg
	./compile -o $(@) $<  *.go

# Generate asm files for a test binary by babygo compiler compiling test
$(tmp)/pre-bbg-test.d: $(tmp)/pre-bbg t/*.go
	./compile -o $@ $< t/*.go

# Generate asm files for a test binary by pre compiler compiling test
$(tmp)/pre-test.d: $(tmp)/pre t/*.go src/*/* lib/*/*
	./compile -o $@ $< t/*.go

# Make a test binary by pre compiler compiling test
$(tmp)/pre-test: $(tmp)/pre-test.d
	./assemble_and_link $@ $<

# Generate asm files for a test binary by babygo 1gen compiler compiling test
$(tmp)/bbg-test.d: $(tmp)/bbg t/*.go
	./compile -o $@ $< t/*.go

# Generate asm files for a test binary by babygo 1gen compiler compiling test
$(tmp)/bbg-test: $(tmp)/bbg-test.d
	./assemble_and_link $@ $<

# Generate asm files for a test binary by babygo 2gen compiler compiling test
$(tmp)/bbg-bbg-test.d: $(tmp)/bbg-bbg t/*.go
	./compile -o $@ $< t/*.go

# Generate asm files for a test binary by babygo 2gen compiler compiling test
$(tmp)/bbg-bbg-test: $(tmp)/bbg-bbg-test.d
	./assemble_and_link $@ $<


# make test expectations
t/expected.txt: t/*.go lib/*/*
	export FOO=bar; go run t/*.go myargs > t/expected.txt

# test the test binary made by pre compiler
.PHONY: test0
test0: $(tmp)/pre-test t/expected.txt
	./test.sh $< $(tmp)

# test the test binary made by babygo 1gen compiler
.PHONY: test1
test1: $(tmp)/bbg-test t/expected.txt
	./test.sh $< $(tmp)

# test the test binary made by babygo 2gen compiler
.PHONY: test2
test2: $(tmp)/bbg-bbg-test t/expected.txt
	./test.sh $< $(tmp)

# do selfhost check by comparing 2gen and 3gen asm files
.PHONY: selfhost
selfhost: $(tmp)/bbg-bbg.d $(tmp)/bbg-bbg-bbg.d
	diff $(tmp)/bbg-bbg.d/all $(tmp)/bbg-bbg-bbg.d/all
	@echo "self host is ok"

# compare output of test0 and test1
.PHONY: compare-test
compare-test: $(tmp)/pre-test.d $(tmp)/bbg-test.d $(tmp)/bbg-bbg-test.d $(tmp)/pre-bbg-test.d
	diff -u $(tmp)/pre-test.d/all $(tmp)/bbg-test.d/all
	diff -u $(tmp)/bbg-test.d/all $(tmp)/pre-bbg-test.d/all
	diff -u $(tmp)/bbg-test.d/all $(tmp)/bbg-bbg-test.d/all

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
