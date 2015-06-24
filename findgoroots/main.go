package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func findRootsR(path string) {
	var infos []os.FileInfo
	var err error

	if infos, err = ioutil.ReadDir(path); err != nil {
		return
	}

	for _, info := range infos {
		if info.IsDir() && info.Name()[0] != '.' {
			// log.Println("checking", info.Name())
			dpath := filepath.Join(path, info.Name())

			if !isGoRoot(dpath) {
				findRootsR(dpath)
			}
		}
	}
}

func isGoRoot(path string) bool {
	// log.Println("isGoRoot(" + path + ")")
	var infos []os.FileInfo
	var err error

	if infos, err = ioutil.ReadDir(path); err != nil {
		return false
	}

	score := 0

	for _, info := range infos {
		if !info.IsDir() {
			continue
		}

		nm := info.Name()

		if nm == "src" || nm == "bin" || nm == "pkg" {
			score++
		}

		// if nm == "src" {
		// 	if _, err := os.Stat(filepath.Join(path, "src/runtime/error.go")); err != nil {
		// 		score--
		// 	}
		// }
	}

	if score == 3 {
		fmt.Println(path)
		return true
	}

	return false
}

var (
	root         *string = flag.String("C", ".", "root directory to search")
	includepaths *bool   = flag.String("-p", false, "include found GOPATHs")
)

func main() {
	flag.Parse()

	findRootsR(root)
}
