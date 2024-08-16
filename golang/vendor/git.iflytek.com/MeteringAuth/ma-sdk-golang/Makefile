#/bin/bash

vet:
	CGO_ENABLED=0 GOPATH=$(PWD)/../../../.. go vet demo/licc/*.go
	CGO_ENABLED=0 GOPATH=$(PWD)/../../../.. go vet demo/calc/*.go
	CGO_ENABLED=0 GOPATH=$(PWD)/../../../.. go vet demo/rep/*.go
clean:
	rm -f demo/licc/licc
	rm -f demo/calc/calc
	rm -f demo/rep/rep
