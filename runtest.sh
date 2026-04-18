#!/bin/bash
cd builtins
./buildtollvm.sh
cd ..
go build
./grianlang3 --keepll test.gl3
./out
