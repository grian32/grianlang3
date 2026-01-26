#!/bin/bash

for f in ./*.c ; do
  clang -S -emit-llvm -O2 "$f" -o "${f%.c}.ll"
done