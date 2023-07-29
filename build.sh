#!/usr/bin/env bash
#
# Usage:
#   build.sh -o BINARY MAIN_DIR
#
####
set -eu
REPO_ROOT=$(cd $(dirname $0);pwd)
if [[ -z $WORKDIR ]]; then
  echo "WORKDIR must be set" >/dev/stderr
  exit 1
fi

OUT_FILE=$2
OUT_FILE_ABS=$(realpath $OUT_FILE)
MAIN_PKG_PATH=$3
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
ld -o a.out *.o
cp a.out $OUT_FILE_ABS

echo $OUT_FILE
