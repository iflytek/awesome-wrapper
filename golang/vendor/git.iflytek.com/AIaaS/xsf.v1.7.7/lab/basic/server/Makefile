.PHONY: all offline online build

all: build online

offline:
	server.exe -m 0 -c server.toml -p xsf -g xsf -s xsf-server -u "http://10.1.87.69:6868"

online:
	server.exe -m 1 -c server.toml -p xsf -g xsf -s xsf-server -u "http://10.1.87.69:6868"

build:
	go clean
	go build

linux:export GOOS=linux
linux:
	go clean
	go build