# Babygo: a go compiler made from scratch


# Target machine

x86-64 Linux

# Test

```terminal
$ ./docker-run
# make test
```

# How to do self hosting

```terminal
$ ./docker-run

#  go build -o babygo main.go # 1st generation compiler
# ./babygo < main.go > /tmp/babygo2.s
# as -o babygo2 /tmp/babygo2.s runtime.s
# as -o babygo2.o /tmp/babygo2.s runtime.s
# ld -o babygo2 babygo2.o # 2nd generation compiler
# ./babygo2 < main.go > /tmp/babygo3.s
# diff /tmp/babygo2.s /tmp/babygo3.s # assert babygo2.s and babygo3.s are same.
```

# LICENSE

MIT

# AUTHOR

@DQNEO
