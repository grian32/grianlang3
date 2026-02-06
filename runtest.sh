#!/bin/bash
go build
./grianlang3 test.gl3
./out
rm -rf lltemp
