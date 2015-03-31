//+build ignore

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// var writechanges *bool = flag.Bool("w", false, "write result to (source) file instead of stdout")
var outputfile *string = flag.String("o", "", "write result to this output file")
var sprefix *string = flag.String("p", "C", "prepend this string to generated function names")

func init() {
	flag.Usage = func() {
		prog := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "usage: %s [options] FILE\n\n%s generates a Go source wrapper for a C header file.\n\nOPTIONS\n", prog, prog)
		flag.PrintDefaults()
	}
}

var (
	trim    = strings.TrimSpace
	prefix  = strings.HasPrefix
	replace = strings.Replace
	split   = strings.Split
)

type Param struct {
	Type  string
	Name  string
	CName string
}

func (p Param) String() string {
	return fmt.Sprintf("%s %s", p.Name, ctypeToGoType(p.Type))
}

func strings2params(params []string) (ps []Param) {
	for _, param := range params {
		var ftype, fname string

		pparts := split(param, " ")

		ftype = pparts[0]
		fname = pparts[1]

		if prefix(fname, "*") {
			ftype += "*"
			fname = fname[1:]
		}

		if strings.HasSuffix(fname, "[]") {
			ftype += "*"
			fname = fname[:len(fname)-2]
		}

		ps = append(ps, Param{ftype, fname, "c" + strings.Title(fname)})
	}

	return ps
}

func ctypeToGoType(ctype string) string {
	switch ctype {
	case "char*":
		return "string"
	case "char**":
		return "[]string"
	case "int":
		return "int"
	case "short":
		return "bool"
	default:
		log.Fatalln("unknown C type", ctype)
	}

	panic("should not have reached here!")
	return ""
}

var w io.Writer

func Print(a ...interface{}) {
	fmt.Fprint(w, a...)
}
func Println(a ...interface{}) {
	Print(fmt.Sprintln(a...))
}
func Printf(format string, a ...interface{}) {
	Print(fmt.Sprintf(format, a...))
}

func generate(src string) {
	f, _ := os.Open(src)
	b, _ := ioutil.ReadAll(f)
	s, _ := filepath.Abs(src)

	Println("package", filepath.Base(filepath.Dir(s)))
	Println("\n// Auto-generated Go wrapper for", src, "\n// DO NOT EDIT THIS FILE BY HAND\n")

	Printf("//#include \"%s\"\n", src)

	if !strings.Contains(string(b), "#include <stdlib.h>") {
		Printf("//#include <stdlib.h>\n")
	}

	Println("import \"C\"\n")
	Println("import \"unsafe\"\n")

	for _, line := range split(string(b), "\n") {
		line = strings.TrimSpace(line)

		switch {
		case line == "":
			continue
		case prefix(line, "#") || prefix(line, "/"):
			continue
		default:
			line = trim(replace(line, "extern", "", -1))

			returnType := split(line, " ")[0]
			funcName := trim(split(split(line, "(")[0], " ")[1])
			params := split(split(split(line, "(")[1], ")")[0], ",")

			for i, p := range params {
				params[i] = trim(p)
			}

			plist := strings2params(params)

			Printf("func %s%s (", sprefix, funcName)

			pstrings := []string{}

			for _, p := range plist {
				pstrings = append(pstrings, p.String())
			}

			Print(strings.Join(pstrings, ", "))

			Print(")")

			if returnType != "void" {
				Print(" ", returnType)
			}

			Print(" {\n")

			for _, p := range plist {
				switch p.Type {
				case "char*":
					Print("\t", p.CName, " := C.CString(", p.Name, ")\n")
					Print("\tdefer C.free(unsafe.Pointer(", p.CName, "))\n")
				case "short":
					Print("\t", p.CName, " := C.short(0)\n")
					Print("\tif ", p.Name, " { ", p.CName, " = C.short(1) }\n")
				case "int":
					Print("\t", p.CName, " := C.int(", p.Name, ")\n")
				case "char**":
					Print("\t", p.CName, " := []*C.char{}\n")
					Print("\tfor _, val := range ", p.Name, " {\n\t\tcval := C.CString(val)\n\t\tdefer C.free(unsafe.Pointer(&val))\n\t\t", p.CName, " = append(", p.CName, ", cval)\n\t}\n")
				}

				Println()
			}

			if returnType != "void" {
				Print("\treturn ")
			} else {
				Print("\t")
			}

			Print("C.", funcName, "(")

			cargs := []string{}

			for _, p := range plist {
				if p.Type == "char**" {
					cargs = append(cargs, "&"+p.CName+"[0]")

				} else {
					cargs = append(cargs, p.CName)
				}
			}

			Print(strings.Join(cargs, ", "))

			Print(")\n")
			Println("}\n")

			// fmt.Println(returnType, funcName, plist)
		}
	}
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *outputfile == "" || *outputfile == "-" {
		w = os.Stdout
	} else if _, err := os.Stat(*outputfile); err == nil {
		panic("output file exists!")
	} else {
		w, _ = os.Create(*outputfile)
	}

	fmt.Fprintln(os.Stderr, "output", *outputfile)

	for _, fname := range flag.Args() {
		if !strings.HasSuffix(fname, ".h") {
			fmt.Fprintf(os.Stderr, "WARN (%s) skipped input file without .h extension\n", fname)
			continue
		}

		if _, err := os.Stat(fname); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL (%s) file does not exist\n", fname)
			continue
		}

		generate(fname)
	}
}