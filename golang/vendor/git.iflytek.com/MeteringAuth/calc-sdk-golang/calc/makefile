# makefile

# varible define
GOBUILD:= go build
GOTEST:= go test

all: build test


# build sdk
build: calc.go calcError.go configcalc.go core.go
	${GOBUILD} .

# run unit test
test : calc.go calcError.go configcalc.go core.go
	${GOTEST} .

.PHONY : all build test
