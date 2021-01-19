#!/bin/bash
set -u
program=$1
${program} myargs 1>/tmp/actual.1 2> /tmp/actual.2
exit_status=$?

if [[ $exit_status -eq 0 ]]; then
  echo ok
else
  echo FAILED
  echo
  echo "    ${program} myargs 1>/tmp/actual.1 2> /tmp/actual.2"
  echo
  exit $exit_status
fi
