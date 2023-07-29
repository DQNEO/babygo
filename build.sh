#!/usr/bin/env bash
#
# Usage:
#   build.sh -o BINARY COMPILER PKG_DIR
#
####
set -eu
REPO_ROOT=$(cd $(dirname $0);pwd)

shift;

OUT_FILE=$1
OUT_FILE_ABS=$(realpath $OUT_FILE)
compiler=$2
MAIN_PKG_PATH=$3
workdir=${OUT_FILE}.d
export WORKDIR=$workdir
mkdir -p $workdir
PKGS=""
while read p
do
  if [[ $p == "main" ]]; then
    continue
  fi
  PKGS="$PKGS $p"
  p2=$(echo $p | tr '/' '.')
  $compiler compile -o $workdir/$p2  $p
done

# Compile main
$compiler compile -o $workdir/main $MAIN_PKG_PATH

set -x
cd $workdir

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

cat *.s > all
echo $OUT_FILE
