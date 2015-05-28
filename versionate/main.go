package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)
import "github.com/blang/semver"

//go:generate go run $GOFILE -i -w $GOFILE

func readSource(path string) []byte {
	var (
		b   []byte
		in  *os.File
		err error
	)

	if in, err = os.Open(path); err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	if b, err = ioutil.ReadAll(in); err != nil {
		log.Fatal(err)
	}

	return b
}

func versionate(path string) {
	var out *os.File
	var err error
	var ver semver.Version

	path, _ = filepath.Abs(path)
	log.Println(path)

	src := readSource(path)

	if !re.Match(src) {
		fmt.Println("abort: no version const")
		os.Exit(5)
	} else {
		vs := strings.Split(strings.Split(string(re.Find(src)), "\"")[1], " ")[0]
		ver, _ = semver.Make(vs)
	}

	if *overwrite {
		if out, err = os.Open(path); err != nil {
			log.Fatal(err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	// fmt.Println(re.Match(b))
	// _ = out
	// if ver == semver.Make("s") {
	// 	cmd := exec.Command("git", strings.Split("rev-parse --short HEAD", " ")...)
	// 	cmdb, _ := cmd.CombinedOutput()

	// 	v := strings.TrimSpace(string(cmdb))
	// 	if ver, err = semver.Make(v); err != nil {
	// 		ver, _ = semver.Make("0.0.0-devel+" + v)
	// 	}
	// }

	if *incPatch {
		ver.Patch++
	}

	if *incMinor {
		ver.Minor++
	}

	if *incMajor {
		ver.Major++
	}
	output := re.ReplaceAll(src, []byte(fmt.Sprintf(`aconst version string = "%s %s %s/%s"`, ver.String(), time.Now().Format("Mon Jan 2 15:04:05 2006 -0700"), runtime.GOOS, runtime.GOARCH))[1:])
	if *overwrite {
		ioutil.WriteFile(path, output, 0644)
	} else {
		if _, err = os.Stdout.Write(output); err != nil {
			log.Fatal(err)
		}
	}
}

const version string = "0.0.3 Wed May 27 21:24:43 2015 -0400 linux/amd64"

var (
	overwrite *bool          = flag.Bool("w", false, "overwrite input file")
	re        *regexp.Regexp = regexp.MustCompile("\\bconst version string = \"([^\"]*)\"")

	incPatch *bool = flag.Bool("i", false, "increment patch level")
	incMinor *bool = flag.Bool("im", false, "increment minor level")
	incMajor *bool = flag.Bool("iM", false, "increment major level")
)

func main() {
	flag.Parse()

	if _, err := os.Stat("main.go"); flag.NArg() == 0 && err == nil {
		versionate("main.go")
	}

	for _, path := range flag.Args() {
		versionate(path)
	}
}
