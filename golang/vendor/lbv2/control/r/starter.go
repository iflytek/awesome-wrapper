package main

import (
	"flag"
	"fmt"
	"os"
)

func Usage() {
	fmt.Fprint(os.Stderr, "Usage of ", os.Args[0], ":\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\n")
}

func main() {
	flag.Usage = Usage
	cfgFile := flag.String("cfg", "lbv2.toml", "set server config file")
	flag.Parse()

	runClient(*cfgFile)
}
