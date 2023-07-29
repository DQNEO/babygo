#!/usr/bin/env bash
#   unsafe unsafe
#   runtime runtime
#   reflect reflect
#   syscall syscall
#   os os

#   mylib2 github.com/DQNEO/babygo/lib/mylib2
#   mymap github.com/DQNEO/babygo/lib/mymap
#   strconv github.com/DQNEO/babygo/lib/strconv
#   strings github.com/DQNEO/babygo/lib/strings
#   path github.com/DQNEO/babygo/lib/path
#   fmt github.com/DQNEO/babygo/lib/fmt
#   mylib github.com/DQNEO/babygo/lib/mylib
#   token github.com/DQNEO/babygo/lib/token
#   ast github.com/DQNEO/babygo/lib/ast
#   universe github.com/DQNEO/babygo/internal/universe
#   main main
set -eu
declare -r PKGS="
unsafe
runtime
reflect
syscall
os

github.com/DQNEO/babygo/lib/mylib2
github.com/DQNEO/babygo/lib/mymap
github.com/DQNEO/babygo/lib/strconv
github.com/DQNEO/babygo/lib/strings
github.com/DQNEO/babygo/lib/path
github.com/DQNEO/babygo/lib/fmt
github.com/DQNEO/babygo/lib/mylib
github.com/DQNEO/babygo/lib/token
github.com/DQNEO/babygo/lib/ast
github.com/DQNEO/babygo/internal/universe

"

export WORKDIR=/tmpfs/bbg/t2
REPO_ROOT=$(cd $(dirname $0);pwd)

mkdir -p $WORKDIR

for p in $PKGS
do
  p2=$(echo $p | tr '/' '.')
  go run . compile -o $WORKDIR/$p2  $p
done

# Compile main
go run . compile -o $WORKDIR/main  ./t

set -x
cd $WORKDIR

# Generate __INIT__.s
{
cat <<EOF
.text
# Initializes all packages except for runtime
.global __INIT__.init
__INIT__.init:
EOF

PKGSM="${PKGS} main"
for p in $PKGSM
do
  if [[ $p == "runtime" ]] ; then
      continue
  fi
  p2=$(echo $p | tr '/' '.')
  basename=${p##*/}
  echo "  callq ${basename}.__initVars"
  if grep -E '^func init' ${p2}.dcl.go >/dev/null ; then
    echo "  callq ${basename}.init"
  fi
done
echo "  ret"
} > __INIT__.s

# Assemble .s files
for a in *.s
do
  b=${a%*.s}
  as $a -o ${b}.o
done

# Link
ld -o bbg-test *.o

cd $REPO_ROOT
./test.sh $WORKDIR/bbg-test $WORKDIR
