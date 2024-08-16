BINARY=client.exe

.PHONY: clean build run

run:build
	call $(BINARY) -tm 700 -mode 0

build:clean
	go build  -o ${BINARY}

clean:
	go clean