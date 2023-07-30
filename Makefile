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

# make p compiler (a rich binary)
$(tmp)/p:  src/*/* internal/*/* lib/*/*  $(tmp)
	rm -rf pre/internal pre/*.go
	cp *.go pre/
	cp -ar internal pre/
	find pre -name '*.go' | xargs sed -e 's#github.com/DQNEO/babygo/lib/ast#go/ast#' -e 's#github.com/DQNEO/babygo/lib/token#go/token#' -e 's#github.com/DQNEO/babygo/lib/parser#go/parser#' -e 's#babygo/internal#babygo/pre/internal#g' -i
	go build -o $@ ./pre

# babygo 1gen compiler (a rich binary)
$(tmp)/b: *.go lib/*/* src/*/* $(tmp)
	go build -o $@ .

# babygo 2gen compiler (a thin binary) by p compiler
$(tmp)/pb: $(tmp)/p *.go src/*/* lib/*/*
	./go-build -o $@ -c $< ./

# babygo 2gen compiler (a thin binary) by babygo 1gen compiler compiling babygo
$(tmp)/bb: $(tmp)/b
	./go-build -o $@ -c $< ./

# babygo 3gen compiler (a thin binary) by babygo 2gen compiler compiling babygo
$(tmp)/bbb: $(tmp)/bb
	./go-build -o $@ -c $< ./

# test binary made by p
$(tmp)/pt: $(tmp)/p t/*.go src/*/* lib/*/*
	./go-build -o $@ -c $< ./t

# test binary made by babygo 1 gen compiler
$(tmp)/bt: $(tmp)/b t/*.go
	./go-build -o $@ -c $< ./t

# test binary made by pb
$(tmp)/pbt: $(tmp)/pb t/*.go
	./go-build -o $@ -c $< ./t

# test binary made by bb
$(tmp)/bbt: $(tmp)/bb t/*.go
	./go-build -o $@ -c $< ./t


# make test expectations
t/expected.txt: t/*.go lib/*/*
	export FOO=bar; go run t/*.go myargs > t/expected.txt

# test the test binary made by p compiler
.PHONY: test0
test0: $(tmp)/pt t/expected.txt
	./test.sh $< $(tmp)
	@echo "[PASS] test by p"

# test the test binary made by babygo 1gen compiler
.PHONY: test1
test1: $(tmp)/bt t/expected.txt
	./test.sh $< $(tmp)
	@echo "[PASS] test by b"

# test the test binary made by babygo 2gen compiler
.PHONY: test2
test2: $(tmp)/bbt t/expected.txt
	./test.sh $< $(tmp)
	@echo "[PASS] test by bb"

# do selfhost check by comparing 2gen and 3gen asm files
.PHONY: selfhost
selfhost: $(tmp)/bb $(tmp)/bbb
	diff $(tmp)/bb.d/all $(tmp)/bbb.d/all  >/dev/null
	@echo "[PASS] selfhost"

# compare output of test0 and test1
.PHONY: compare-test
compare-test: $(tmp)/pt $(tmp)/bt $(tmp)/bbt $(tmp)/pbt
	diff -u $(tmp)/pt.d/all $(tmp)/bt.d/all >/dev/null
	diff -u $(tmp)/bt.d/all $(tmp)/pbt.d/all  >/dev/null
	diff -u $(tmp)/bt.d/all $(tmp)/bbt.d/all  >/dev/null
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
