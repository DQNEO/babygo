# Run this on a docker container
tmp ?= /tmp/bbg

# tests without pre stuff involved
.PHONY: all
all: test1 test2 selfhost

# run all tests
.PHONY: test
test: $(tmp) test1 test2 selfhost test0 compare-test

$(tmp):
	mkdir -p $(tmp)

# prepare precompiler source file
pre/precompiler.go: *.go internal/*/* lib/*/*

# make pre compiler (a rich binary)
$(tmp)/pre:  src/*/* internal/*/* lib/*/*  $(tmp)
	rm -rf pre/internal pre/*.go
	cp *.go pre/
	cp -ar internal pre/
	find pre -name '*.go' | xargs sed -e 's#github.com/DQNEO/babygo/lib/ast#go/ast#' -e 's#github.com/DQNEO/babygo/lib/token#go/token#' -e 's#github.com/DQNEO/babygo/lib/parser#go/parser#' -e 's#babygo/internal#babygo/pre/internal#g' -i
	go build -o $@ ./pre

# make babygo 1gen compiler (a rich binary)
$(tmp)/bbg: *.go lib/*/* src/*/* $(tmp)
	go build -o $@ .

# Generate asm files for babygo 2gen compiler by pre compiler compiling babygo
# Make babygo 2 gen compiler (a thin binary) by pre compiler
$(tmp)/pb: $(tmp)/pre *.go src/*/* lib/*/*
	./bld -o $@ $< *.go

# Generate asm files for babygo 2gen compiler by babygo 1gen compiler compiling babygo
# Make babygo 2gen compiler (a thin binary)
$(tmp)/bb: $(tmp)/bbg
	./bld -o $@ $< *.go

# Generate asm files for babygo 3gen compiler by babygo 2gen compiler compiling babygo
$(tmp)/bbb: $(tmp)/bb
	./bld -o $@ $< *.go

# Generate asm files for a test binary by babygo compiler compiling test
$(tmp)/pbt: $(tmp)/pb t/*.go
	./bld -o $@ $< t/*.go

# Generate asm files for a test binary by pre compiler compiling test
# Make a test binary by pre compiler compiling test
$(tmp)/pt: $(tmp)/pre t/*.go src/*/* lib/*/*
	./bld -o $@ $< t/*.go

# Generate asm files for a test binary by babygo 1gen compiler compiling test
# Generate asm files for a test binary by babygo 1gen compiler compiling test
$(tmp)/bt: $(tmp)/bbg t/*.go
	./bld -o $@ $< t/*.go

# Generate asm files for a test binary by babygo 2gen compiler compiling test
# Generate asm files for a test binary by babygo 2gen compiler compiling test
$(tmp)/bbt: $(tmp)/bb t/*.go
	./bld -o $@ $< t/*.go


# make test expectations
t/expected.txt: t/*.go lib/*/*
	export FOO=bar; go run t/*.go myargs > t/expected.txt

# test the test binary made by pre compiler
.PHONY: test0
test0: $(tmp)/pt t/expected.txt
	./test.sh $< $(tmp)
	@echo "[PASS] test by pre"

# test the test binary made by babygo 1gen compiler
.PHONY: test1
test1: $(tmp)/bt t/expected.txt
	./test.sh $< $(tmp)
	@echo "[PASS] test by bbg"

# test the test binary made by babygo 2gen compiler
.PHONY: test2
test2: $(tmp)/bbt t/expected.txt
	./test.sh $< $(tmp)
	@echo "[PASS] test by bb"

# do selfhost check by comparing 2gen and 3gen asm files
.PHONY: selfhost
selfhost: $(tmp)/bb $(tmp)/bbb
	diff $(tmp)/bb $(tmp)/bbb  >/dev/null
	@echo "[PASS] selfhost"

# compare output of test0 and test1
.PHONY: compare-test
compare-test: $(tmp)/pt $(tmp)/bt $(tmp)/bbt $(tmp)/pbt
	diff -u $(tmp)/pt $(tmp)/bt >/dev/null
	diff -u $(tmp)/bt $(tmp)/pbt  >/dev/null
	diff -u $(tmp)/bt $(tmp)/bbt  >/dev/null
	@echo "[PASS] compare-test"

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
