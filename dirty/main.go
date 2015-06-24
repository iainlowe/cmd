package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	quiet = flag.Bool("q", false, "be quiet (just list dirs)")
	root  = flag.String("C", ".", "root for search start")
)

func fullrel(s string) string {
	s, _ = filepath.Abs(s)
	wd, _ := os.Getwd()
	return filepath.Join(".", strings.TrimPrefix(s, wd))

}

func findDirty(d string) {
	d = fullrel(d)

	if filepath.Base(d) == ".git" {
		return
	}

	if _, err := os.Stat(filepath.Join(d, ".git")); err == nil {
		c := exec.Command("git", "status", "-s")
		c.Dir = d
		b, err := c.CombinedOutput()

		if err != nil {
			fmt.Println("error:", d, err)
			fmt.Println(string(b))
		}

		if strings.TrimSpace(string(b)) != "" {
			fmt.Println(d)

			if !*quiet {
				fmt.Println(string(b))
			}
		}
	}

	infos, _ := ioutil.ReadDir(d)

	for _, info := range infos {
		if info.IsDir() {
			findDirty(filepath.Join(d, info.Name()))
		}
	}
}

func main() {
	flag.Usage = func() {
		fmt.Println("usage: dirty [-C <dir>] [-q]\nList dirty repositories.")
	}
	flag.Parse()
	findDirty(*root)
}
