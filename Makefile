# Run this on Linux
tmp = /tmp/bbg

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
	go build -o $@ ./pre

$(tmp)/babygo: $(tmp)  *.go lib/*/*
	go build -o $@ .

$(tmp)/cross: *.go src/*/* $(tmp)/pre
	rm /tmp/work/*.s
	$(tmp)/pre *.go
	cat src/runtime/runtime.s /tmp/work/*.s > $(tmp)/pre-main.s
	cp $(tmp)/pre-main.s ./.shared/ # for debug
	as -o $(tmp)/a.o $(tmp)/pre-main.s
	ld -o $@ $(tmp)/a.o

$(tmp)/babygo2: $(tmp)/babygo src/*/*
	rm /tmp/work/*.s
	$(tmp)/babygo *.go
	cat src/runtime/runtime.s /tmp/work/*.s > $(tmp)/babygo-main.s
	cp $(tmp)/babygo-main.s ./.shared/ # for debug
	as -o $(tmp)/a.o $(tmp)/babygo-main.s
	ld -o $@ $(tmp)/a.o

$(tmp)/pre-test.s: t/test.go src/*/* $(tmp)/pre
	rm /tmp/work/*.s
	$(tmp)/pre t/test.go t/another.go
	cat src/runtime/runtime.s /tmp/work/*.s > $@
	cp $@ ./.shared/

$(tmp)/cross-test.s: t/test.go $(tmp)/cross
	rm /tmp/work/*.s
	$(tmp)/cross t/test.go t/another.go
	cat src/runtime/runtime.s /tmp/work/*.s > $@
	cp $@ ./.shared/

$(tmp)/babygo-test.s: t/test.go src/*/* $(tmp)/babygo
	rm /tmp/work/*.s
	$(tmp)/babygo t/test.go t/another.go
	cat src/runtime/runtime.s /tmp/work/*.s > $@
	cp $@ ./.shared/

$(tmp)/babygo2-test.s: t/test.go $(tmp)/babygo2
	rm /tmp/work/*.s
	$(tmp)/babygo2 t/test.go t/another.go
	cat src/runtime/runtime.s /tmp/work/*.s > $@
	cp $@ ./.shared/

# compare output of test0 and test1
.PHONY: compare-test
compare-test: $(tmp)/pre-test.s $(tmp)/babygo-test.s $(tmp)/babygo2-test.s $(tmp)/cross-test.s
	diff -u $(tmp)/babygo-test.s $(tmp)/babygo2-test.s
	diff -u $(tmp)/babygo-test.s $(tmp)/cross-test.s

.PHONY: test0
test0: $(tmp)/test0 t/expected.txt
	./test.sh $<

.PHONY: test1
test1: $(tmp)/test1 t/expected.txt
	./test.sh $<

.PHONY: testcross
testcross: $(tmp)/testcross t/expected.txt
	./test.sh $<

$(tmp)/test0: $(tmp)/pre-test.s src/*/*
	as -o $(tmp)/a.o $<
	ld -o $@ $(tmp)/a.o

$(tmp)/test1: $(tmp)/babygo-test.s src/*/*
	as -o $(tmp)/a.o $<
	ld -o $@ $(tmp)/a.o

$(tmp)/testcross: $(tmp)/cross-test.s src/*/*
	as -o $(tmp)/a.o $<
	ld -o $@ $(tmp)/a.o

# test self hosting by comparing 2gen.s and 3gen.s
.PHONY: selfhost
selfhost: $(tmp)/babygo $(tmp)/babygo2 $(tmp)/babygo-main.s
	@echo "testing self host ..."
	rm /tmp/work/*.s
	$(tmp)/babygo2   *.go
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
