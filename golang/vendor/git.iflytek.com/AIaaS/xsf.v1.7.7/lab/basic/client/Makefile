.PHONY:run build clean

all: clean build run

run:
	client.exe -mode 1 -tm 1000

build: export GOOS=windows
build:
	go clean
	go build

clean:
	del log\client.log
