package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func abs(path string) string {
	p, _ := filepath.Abs(path)
	return p
}

var (
	git *bool = flag.Bool("git", false, "search for .git directory")
	hg  *bool = flag.Bool("hg", false, "search for .hg directory")
	bzr *bool = flag.Bool("bzr", false, "search for .bzr directory")

	exact *bool   = flag.Bool("x", false, "match name exactly")
	first *bool   = flag.Bool("1", false, "return the first match")
	start *string = flag.String("C", ".", "override search start directory")
)

func init() {
	flag.Usage = func() {
		fmt.Println("usage: rfind [options] EXPR...")
		fmt.Println()
		fmt.Println("rfind searches for matching files walking up the directory tree towards /")
		fmt.Println()
		flag.PrintDefaults()
	}
}

func search(root string, expr string) []string {
	matches := []string{}

	re := regexp.MustCompile(strings.Replace(expr, ".", "\\.", -1))

	for path := *start; abs(path) != "/"; path += "/.." {
		infos, _ := ioutil.ReadDir(path)
		for _, info := range infos {
			matched := filepath.Join(abs(path), info.Name())

			if re.MatchString(matched) {
				// fmt.Println(">", matched)
				matches = append(matches, matched)

				if *first {
					fmt.Println(matched)
					os.Exit(0)
				}
			}
		}
	}

	return matches
}

func rfind(exprs ...string) {
	matches := []string{}

	for _, expr := range exprs {
		matches = append(matches, search(*start, expr)...)
	}

	sort.Strings(matches)

	if len(matches) == 1 {
		fmt.Println("a", matches[0])
	} else {
		fmt.Println(strings.Join(matches, "\n"))
	}
}

func main() {
	flag.Parse()

	switch {
	case *git:
		*first = true
		rfind(".git")
	case *hg:
		*first = true
		rfind(".hg")
	case *bzr:
		*first = true
		rfind(".bzr")
	default:
		if flag.NArg() == 0 {
			flag.Usage()
			os.Exit(1)
		}

		rfind(flag.Args()...)
	}
}
