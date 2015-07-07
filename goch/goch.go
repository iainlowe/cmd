package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const version string = "0.0.5 Wed May 27 14:40:45 2015 -0400 linux/amd64"

var (
	outputfile    *string = flag.String("o", "", "write result to this output file")
	overwrite     *bool   = flag.Bool("w", false, "overwrite output file if it exists")
	exclude       *string = flag.String("x", "", "exclude functions matching this regex")
	showversion   *bool   = flag.Bool("version", false, "display version and exit")
	packageover   *string = flag.String("P", "", "override the output package name")
	discarderrors *bool   = flag.Bool("noerr", false, "discard returned error values from C")
	sprefix       *string = flag.String("p", "C", "prepend this string to generated function names")
)

func init() {
	flag.Usage = func() {
		prog := filepath.Base(os.Args[0])
		msg := `Each function in the input will generate a matching Go function that calls the
C function and returns it's return value managing memory appropriately.

` + prog + ` will also copy comments from the header file into the generated file if
they are "attached" to a function definition (ie. there must be no spaces between the
comment block and the function signature).

By default, ` + prog + ` prepends the letter C to all generated functions (ie. a function
called MyFunc will generate a wrapper called CMyFunc). Use the -p flag to override this.`

		fmt.Fprintf(os.Stderr, "usage: %s [options] FILE\n\n%s generates a Go source wrapper for a C header file.\n\n%s\n\nOPTIONS\n", prog, prog, msg)
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

		if len(pparts) != 2 {
			Fatalf("syntax error: \"%s\" is not a valid parameter specification", param)
		}

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

		if strings.Contains(fname, "[]") {
			Fatalf("syntax error: \"%s\" is not a valid parameter specification", param)
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
	case "float":
		return "float32"
	case "void":
		return ""
	default:
		Fatalln("unknown C type", ctype)
	}

	panic("should not have reached here!")
	return ""
}

var w io.Writer

func Print(a ...interface{})                 { fmt.Fprint(w, a...) }
func Println(a ...interface{})               { Print(fmt.Sprintln(a...)) }
func Printf(format string, a ...interface{}) { Print(fmt.Sprintf(format, a...)) }

func Fatal(a ...interface{})                 { log.Fatal(a...) }
func Fatalln(a ...interface{})               { Fatal(fmt.Sprintln(a...)) }
func Fatalf(format string, a ...interface{}) { Fatal(fmt.Sprintf(format, a...)) }

func generate(src string) {
	f, _ := os.Open(src)
	b, _ := ioutil.ReadAll(f)
	s, _ := filepath.Abs(src)

	if *packageover == "" {
		*packageover = filepath.Base(filepath.Dir(s))
	}

	Println("package", *packageover)
	Println("\n// Auto-generated Go wrapper for", src, "\n// DO NOT EDIT THIS FILE BY HAND\n")

	if *discarderrors {
		Println("// WARNING: C error return logic disabled\n")
	}

	Printf("//#include \"%s\"\n", src)

	if !strings.Contains(string(b), "#include <stdlib.h>") {
		Printf("//#include <stdlib.h>\n")
	}

	Println("import \"C\"\n")
	Println("import \"unsafe\"\n")

	lastComment := ""

	for _, line := range split(string(b), "\n") {
		line = strings.TrimSpace(line)

		switch {
		case line == "":
			lastComment = ""
			continue
		case prefix(line, "#"):
			continue
		case prefix(line, "//"):
			lastComment = line
		default:
			line = trim(replace(line, "extern", "", -1))

			returnType := split(line, " ")[0]
			funcName := trim(split(split(line, "(")[0], " ")[1])
			params := split(split(split(line, "(")[1], ")")[0], ",")

			if *exclude != "" && regexp.MustCompile(*exclude).MatchString(funcName) {
				Printf("//skipped %s in output\n\n", funcName)
				continue
			}

			if funcName[0] == '*' {
				returnType += "*"
				funcName = funcName[1:]
			}

			goReturnType := ctypeToGoType(returnType)

			parg := []string{}
			for i, p := range params {
				params[i] = trim(p)
				if params[i] != "" {
					parg = append(parg, params[i])
				}
			}

			plist := strings2params(parg)

			if lastComment != "" {
				Println(lastComment)
				lastComment = ""
			}

			Printf("func %s%s(", *sprefix, funcName)

			pstrings := []string{}

			for _, p := range plist {
				pstrings = append(pstrings, p.String())
			}

			Print(strings.Join(pstrings, ", "))

			Print(") ")

			if returnType != "void" {
				Print("(", goReturnType)
				if !*discarderrors {
					Print(", error")
				}
				Print(")")
			} else {
				if !*discarderrors {
					Print("error")
				}
			}

			Print(" {\n")

			for _, p := range plist {
				switch p.Type {
				case "char*":
					Print("\t", p.CName, " := C.CString(", p.Name, ")\n")
					Print("\tdefer C.free(unsafe.Pointer(", p.CName, "))\n")
				case "short":
					Print("\t", p.CName, " := C.short(0)\n")
					Print("\tif ", p.Name, " {\n\t\t", p.CName, " = C.short(1)\n\t}\n")
				case "int":
					Print("\t", p.CName, " := C.int(", p.Name, ")\n")
				case "float":
					Print("\t", p.CName, " := C.float(", p.Name, ")\n")
				case "char**":
					Print("\t", p.CName, " := []*C.char{}\n")
					Print("\tfor _, val := range ", p.Name, " {\n\t\tcval := C.CString(val)\n\t\tdefer C.free(unsafe.Pointer(cval))\n\t\t", p.CName, " = append(", p.CName, ", cval)\n\t}\n")
				}

				Println()
			}

			estring := ", err"

			if *discarderrors {
				estring = ""
			}

			if returnType != "void" {
				Print("\tv", estring, " := C.", funcName, "(")
			} else if !*discarderrors {
				Print("\t_", estring, " := C.", funcName, "(")
			} else {
				Print("\tC.", funcName, "(")
			}

			cargs := []string{}

			for _, p := range plist {
				if p.Type == "char**" {
					cargs = append(cargs, "&"+p.CName+"[0]")
				} else {
					cargs = append(cargs, p.CName)
				}
			}

			Print(strings.Join(cargs, ", "))

			Print(")")

			switch returnType {
			case "void":
				if !*discarderrors {
					Print("\n\n\treturn err")
				}
			case "char*":
				Print("\n\n\ttmp := C.GoString(v)\t// copy the string so ownership doesn't get confused\n\treturn (v + \" \")[:len(v)]", estring)
			case "short":
				Print("\n\n\treturn v > 0", estring)
			default:
				Print("\n\n\treturn "+goReturnType+"(v)", estring)
			}

			Println("\n}\n")

			// fmt.Println(returnType, funcName, plist)
		}
	}
}

func main() {
	flag.Parse()

	if *showversion {
		fmt.Println(version)
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *outputfile == "" || *outputfile == "-" {
		w = os.Stdout
	} else if _, err := os.Stat(*outputfile); err == nil {
		if *overwrite {
			w, _ = os.Open(*outputfile)
		} else {
			panic("output file exists!")
		}
	} else {
		w, _ = os.Create(*outputfile)
	}

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
