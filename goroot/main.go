package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	pp := append([]string{runtime.GOROOT()}, os.Args[1:]...)
	fmt.Println(filepath.Join(pp...))
}
