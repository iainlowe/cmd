package main

import (
	"flag"
	"fmt"
	"math"
)

func bytesToHuman(n uint64) string {
	G := uint64(math.Pow(1024, 3))
	M := uint64(math.Pow(1024, 2))
	K := uint64(1024)

	s := ""

	switch {
	case n > G:
		s = fmt.Sprintf("%vG", n/G)
	case n > M:
		s = fmt.Sprintf("%vM", n/M)
	case n > K:
		s = fmt.Sprintf("%vK", n/K)
	default:
		s = fmt.Sprintf("%vB", n)
	}

	return s
}

var (
	verbose *bool = flag.Bool("v", false, "be verbose")
	human   *bool = flag.Bool("H", false, "human-readable output")
)

func warn(a ...interface{}) error {
	fmt.Println(a...)
	return nil
}

func main() {
	flag.Parse()

	checkfs()
}
