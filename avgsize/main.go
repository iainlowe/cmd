package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

var (
	format *string = flag.String("format", "B", "one of (B/b, K/k, M/m, G/g); output in this unit")
)

func avgSize(path string) {
	var n, sz float64

	filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
		}
		if !info.IsDir() {
			n++
			sz += float64(info.Size())
		}
		return nil
	})

	switch strings.ToUpper(*format) {
	case "K":
		fmt.Printf("% 10dK %s\n", int64(math.Ceil(sz/n/1024.0)), strings.TrimPrefix(strings.TrimSuffix(path, "/"), "./"))
	default:
		fmt.Printf("% 10d %s\n", int64(sz/n), strings.TrimPrefix(strings.TrimSuffix(path, "/"), "./"))
	}
}

func main() {
	flag.Usage = func() {
		fmt.Println("usage: avgsize [-format bkmg] [PATH]..\n\nPrints the average size of files in PATH and sub-directories.\n\nOPTIONS")
		flag.PrintDefaults()
	}
	flag.Parse()

	for _, arg := range flag.Args() {
		avgSize(arg)
	}

	if flag.NArg() == 0 {
		avgSize(".")
	}

}
