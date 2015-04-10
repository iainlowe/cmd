package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	flag "github.com/ogier/pflag"
)

var (
	fskips       *string = flag.StringP("exclude", "x", "", "list of names to exclude during rip (comma-separated)")
	fpicks       *string = flag.StringP("include", "i", "", "list of names to include during rip (comma-separated)")
	outputfile   *string = flag.StringP("output", "o", "-", "write output to this file ('-' for stdout)")
	skipComments *bool   = flag.Bool("no-comments", false, "do not output comments from ripped package")
)

const usage string = `usage: gorip [options] <package>

Parses the supplied package and outputs all declarations (exported and un-exported).

If -x/--exclude is specified, declarations matching those names will not be output.

If -i/--include is specified, ONLY declarations matching those names will be output, disregarding
the -x/--exclude option.

OPTIONS`

func init() {
	flag.Usage = func() {
		fmt.Println(usage)
		flag.PrintDefaults()
	}
}

func in(ss []string, s string) bool {
	for _, sq := range ss {
		if sq == s {
			return true
		}
	}

	return false
}

func getCurrentPackageName() (pkgname string) {
	var (
		err  error
		pkgs map[string]*ast.Package
	)

	bsfset := token.NewFileSet()

	if pkgs, err = parser.ParseDir(bsfset, ".", nil, parser.ImportsOnly); err != nil {
		log.Fatal(err)
	}

	for pn, _ := range pkgs {
		pkgname = pn
	}

	if pkgname == "" {
		panic("not in a go package directory!")
	}

	return pkgname
}

func main() {
	var (
		skips   []string
		picks   []string
		outputf *os.File
		err     error
		pkgs    map[string]*ast.Package
	)

	flag.Parse()

	pkgname := flag.Args()[0]

	if *fskips != "" {
		skips = strings.Split(*fskips, ",")
	}

	if *fpicks != "" {
		picks = strings.Split(*fpicks, ",")
	}

	// log.Println(len(skips[:]), skips, len(picks[:]), picks)

	outputpkg := getCurrentPackageName()
	fset := token.NewFileSet()

	pkgpath := filepath.Join(os.Getenv("GOPATH"), "src", pkgname)

	if _, err := os.Stat(pkgpath); err != nil {
		panic(err)
	}

	if pkgs, err = parser.ParseDir(fset, pkgpath, nil, parser.ParseComments); err != nil {
		log.Fatal(err)
	}

	pkg := pkgs[filepath.Base(pkgpath)]

	if *outputfile == "-" {
		outputf = os.Stdout
	} else {
		if outputf, err = os.Create(*outputfile); err != nil {
			log.Fatal(err)
		}
	}

	if outputf == nil {
		log.Fatal("failed to set output file")
	}

	astfile := ast.MergePackageFiles(pkg, ast.FilterFuncDuplicates|ast.FilterImportDuplicates)
	ast.FilterFile(astfile, func(name string) bool {
		if in(skips, name) || (len(picks) > 0 && !in(picks, name)) {
			return false
		}
		return true
	})

	fmt.Fprintln(outputf, "package", outputpkg)
	fmt.Fprintln(outputf)
	format.Node(outputf, fset, astfile)
}
