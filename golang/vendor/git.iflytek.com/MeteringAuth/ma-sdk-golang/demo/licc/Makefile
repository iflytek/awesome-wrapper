#/bin/bash
build:
	GO111MODULE="off" CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPATH=$(PWD)/../../../../../.. go build -v .
run:
	./licc -m 1 -p metrics -g reporter -s ma-client -u http://10.1.87.70:6868
