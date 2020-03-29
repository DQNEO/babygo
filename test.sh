#!/bin/bash

./a.out

if [[ $? -eq 42 ]]; then
  echo ok
else
  echo error
  exit 1
fi
