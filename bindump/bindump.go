package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func bindumpFile(path string) {
	bin, _ := ioutil.ReadFile(path)
	bindump(bin)
}

func bindump(b []byte) {
	for i, c := range b {
		fmt.Printf("%02x", c)

		if i%2 != 0 {
			fmt.Print(" ")
		}

		if i > 0 && i%16 == 0 {
			fmt.Print("\n")
		}
	}

	fmt.Println()
}

var (
	verbose *bool = flag.Bool("v", false, "be verbose")
)

func main() {
	flag.Usage = func() {
		fmt.Println("usage: bindump [options] [FILE]")
		fmt.Println("dumps binary data in hex format; if FILE is missing or is '-' data is read from stdin")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() == 1 {
		if flag.Args()[0] != "-" {
			bindumpFile(flag.Args()[0])
			return
		}
	}

	b, _ := ioutil.ReadAll(os.Stdin)
	bindump(b)
}
