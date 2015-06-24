package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	dryrun = flag.Bool("n", false, "dry run")
	quiet  = flag.Bool("q", false, "be quiet")
)

func println(v ...interface{}) {
	if !*quiet {
		fmt.Println(v...)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Println("usage: gorm [options] github.com/package/todelete\n\nRemoves a Go package and cleans any empty resulting directories.\n\nOPTIONS")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	var clean func(p string)

	_clean := func(n int, p string) bool {
		if n != 0 {
			return false
		}

		os.Remove(p)
		if !*quiet {
			fmt.Println(p)
		}

		return true
	}

	remove := func(p string) {
		os.RemoveAll(p)
		println(p)
		clean(filepath.Dir(p))
	}

	if *dryrun {
		_clean = func(n int, p string) bool {
			if n != 1 {
				return false
			}
			fmt.Println("would clean", p)
			return true
		}

		remove = func(p string) {
			fmt.Println("would remove", p)
			clean(filepath.Dir(p))
		}
	}

	clean = func(p string) {
		files, _ := ioutil.ReadDir(p)
		if !_clean(len(files), p) {
			return
		}
		clean(filepath.Dir(p))
	}

	for _, pkg := range flag.Args() {
		pkgpath := filepath.Join(os.Getenv("GOPATH"), "src", pkg)
		if _, err := os.Stat(pkgpath); err == nil {
			remove(pkgpath)
		}
	}
}
