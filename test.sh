#!/bin/bash
set -u
program=$1
${program} myargs 1>/tmp/actual.1 2> /tmp/actual.2
exit_status=$?
diff t/expected.txt /tmp/actual.1
if [[ $? -ne 0 ]]; then
  echo "stdout differs"
  exit 1
fi

if [[ $exit_status -eq 0 ]]; then
  echo ok
else
  echo FAILED
  exit $exit_status
fi
