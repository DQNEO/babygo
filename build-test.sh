#!/usr/bin/env bash
set -eu
REPO_ROOT=$(cd $(dirname $0);pwd)
if [[ -z $WORKDIR ]]; then
  echo "WORKDIR must be set" >/dev/stderr
  exit 1
fi

OUT_FILE=$1
MAIN_PKG_PATH=$2
mkdir -p $WORKDIR
PKGS=""
while read p
do
  PKGS="$PKGS $p"
  p2=$(echo $p | tr '/' '.')
  go run . compile -o $WORKDIR/$p2  $p
done

# Compile main
go run . compile -o $WORKDIR/main $MAIN_PKG_PATH

set -x
cd $WORKDIR

# Assemble .s files
for a in *.s
do
  b=${a%*.s}
  as $a -o ${b}.o
done


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

# Link
as -o __INIT__.o __INIT__.s
ld -o $OUT_FILE *.o

cd $REPO_ROOT
echo $WORKDIR/$OUT_FILE
