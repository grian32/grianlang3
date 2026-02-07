#!/bin/bash

for f in ./*.c ; do
  clang -S -O3 -emit-llvm -O2 "$f" -o "${f%.c}.ll"
done