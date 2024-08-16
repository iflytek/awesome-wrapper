#!/usr/bin/env bash
export GOPATH=`pwd`/../../ #(lb文件夹需在该src目录下)
go clean
go build -v -o lbv2 #编译

./lbv2 -m 0 -c lbv2.toml  -s lbv2