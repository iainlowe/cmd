package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"

	"code.google.com/p/go-uuid/uuid"
)

var sizes = map[string]int{
	"K": 1024,
	"M": 1024 * 1024,
	"G": 1024 * 1024 * 1024,
}

func main() {
	var sz string
	var numfiles int

	var ext = flag.String("e", "rand", "The extension to use for output files (omitting the . separator)")
	flag.StringVar(ext, "ext", "rand", "The extension to use for output files (omitting the . separator)")

	flag.StringVar(&sz, "s", "512M", "The size of file to generate")
	flag.StringVar(&sz, "size", "512M", "The size of file to generate")

	flag.IntVar(&numfiles, "n", 0, "The number of files to generate (0 for infinite)")
	flag.IntVar(&numfiles, "numfiles", 0, "The number of files to generate (0 for infinite)")

	flag.Parse()

	var nsz int

	if strings.ContainsAny(sz, "bBkKmMgG") {
		nsz, _ = strconv.Atoi(sz[:len(sz)-1])
		nsz *= sizes[strings.ToUpper(string(sz[len(sz)-1]))]
	} else {
		nsz, _ = strconv.Atoi(sz)
	}

	data := make([]byte, nsz)

	// for i, _ := range data {
	// 	data[i] = 'a'
	// }

	pfile, _ := os.Create("profile.pprof")
	pprof.StartCPUProfile(pfile)
	defer pprof.StopCPUProfile()

	for i := 0; i < numfiles; i++ {
		uid := uuid.NewRandom().String()
		fname := uid + "." + *ext

		f, _ := os.Create(fname)
		f.Write(data)

		f.Close()
		fmt.Println(fname, nsz, "bytes")
	}
}
