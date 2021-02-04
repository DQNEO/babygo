#!/bin/bash
set -u
program=$1
export FOO=bar
${program} myargs 1>/tmp/actual.1 2> /tmp/actual.2
if [[ $? -eq 0 ]]; then
  :
else
  echo FAILED
  echo
  echo "    ${program} myargs"
  echo
  exit $exit_status
fi

diff t/expected.txt /tmp/actual.1
if [[ $? -ne 0 ]]; then
  echo FAILED
  exit 1
fi

echo "ok"
