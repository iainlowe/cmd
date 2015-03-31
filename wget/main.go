package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	flag "github.com/ogier/pflag"
)

var outputPath string
var verbose bool

func init() {
	flag.StringVarP(&outputPath, "output-file", "O", "", "File path to write to (or - for stdout)")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Be verbose.")

	flag.Usage = func() {
		fmt.Println("usage: wget [options] URL\n\nOptions:")
		flag.PrintDefaults()
	}
}

func get(url string) {
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		fmt.Println("abort:", err)
		os.Exit(5)
	}

	if outputPath == "" {
		outputPath = filepath.Base(url)
	}

	var w io.WriteCloser

	if outputPath == "-" {
		w = os.Stdout
	} else {
		if _, err = os.Stat(outputPath); err == nil {
			fmt.Println("abort: output file already exists!")
			os.Exit(6)
		}

		w, err = os.Create(outputPath)
		defer w.Close()

		p, _ := filepath.Abs(outputPath)

		if verbose {
			fmt.Printf("downloading %s to %s\n", url, p)
		}
	}

	io.Copy(w, resp.Body)
}

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	get(flag.Args()[0])
}
