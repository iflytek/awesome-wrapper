.PHONY:run build clean

all: build

run: clean
	client.exe -mode 1 -tm 1000

build: export GOOS=linux
build:
	go clean
	go build

clean:
	go clean
