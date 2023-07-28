#!/usr/bin/env bash
set -ux
readonly program=$1
readonly tmpdir=$2
export FOO=bar
${program} myargs 1> ${tmpdir}/actual.1 2> ${tmpdir}/actual.2
exit_status=$?
if [[ $exit_status -eq 0 ]]; then
  :
else
  echo FAILED
  echo
  echo "    ${program} myargs"
  echo
  echo "exit status = $exit_status"
  cat ${tmpdir}/actual.2
  exit 1
fi

diff -u t/expected.txt ${tmpdir}/actual.1
if [[ $? -ne 0 ]]; then
  echo FAILED
  exit 1
fi

echo "ok"
