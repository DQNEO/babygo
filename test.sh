#!/bin/bash
set -u
./a.out 1>/tmp/actual.1 2> /tmp/actual.2
exit_status=$?
diff t/expected.1 /tmp/actual.1
if [[ $? -ne 0 ]]; then
  echo "stdout differs"
  exit 1
fi

diff t/expected.2 /tmp/actual.2
if [[ $? -ne 0 ]]; then
  echo "stderr differs"
  exit 1
fi

if [[ $exit_status -eq 42 ]]; then
  echo ok
else
  echo error
  exit 1
fi
