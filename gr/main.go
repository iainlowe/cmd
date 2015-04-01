package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type records struct {
	rs []record
}

func (rs records) String() string {
	s := ""

	for _, em := range rs.Emails() {
		s += fmt.Sprintf("%s\n", rs.Total(em))
	}

	return s
}

type stringset struct {
	strings []string
}

func (ss *stringset) add(s string) {
	for _, is := range ss.strings {
		if is == s {
			return
		}
	}

	ss.strings = append(ss.strings, s)
}

func (rs records) Emails() (ret []string) {
	ss := &stringset{[]string{}}

	for _, r := range rs.rs {
		ss.add(r.email)
	}

	ret = ss.strings
	sort.Strings(ret)
	return ret
}

func (rs records) Total(email string) record {
	ret := record{email: email}

	for _, r := range rs.rs {
		if r.email == email {
			ret.changes += r.changes
			ret.insertions += r.insertions
			ret.deletions += r.deletions
		}
	}

	return ret
}

type record struct {
	email      string
	changes    int
	insertions int
	deletions  int
}

func (r record) String() string {
	return fmt.Sprintf("%-6s: % 5dF % 5d+ % 5d- (net: %+d)", r.email, r.changes, r.insertions, r.deletions, r.insertions-r.deletions)
}

var (
	split = strings.Split
	atoi  = func(s string) int { n, _ := strconv.Atoi(s); return n }
)

var mailmap map[string]string = map[string]string{}

var abs = func(s string) string { r, _ := filepath.Abs(s); return r }

func loadMailMap() {
	for path := *repo; abs(path) != "/"; path += "/.." {
		infos, _ := ioutil.ReadDir(path)
		for _, info := range infos {
			if info.Name() == ".mailmap" {
				f, _ := os.Open(filepath.Join(path, ".mailmap"))
				json.NewDecoder(f).Decode(&mailmap)
			}
		}
	}
}

var (
	repo *string = flag.String("C", ".", "generate report on the repo in this folder")
)

func collectRecords() {
	cmd := exec.Command("git", append(strings.Split("log --shortstat --no-merges --format=format:%cE", " "), "-C", *repo)...)

	b, err := cmd.Output()

	if err != nil {
		if strings.Contains(err.Error(), "128") {
			fmt.Println("abort:", "not a git repo!")
			os.Exit(1)
		}
		fmt.Println("abort:", err)
		os.Exit(1)
	}

	rs := records{}
	r := record{}

	changesRe := regexp.MustCompile("([0-9]+) file")
	insertionsRe := regexp.MustCompile("([0-9]+) insert")
	deletionsRe := regexp.MustCompile("([0-9]+) delet")

	for _, line := range strings.Split(string(b), "\n") {
		switch {
		case strings.HasPrefix(line, " "):
			if m := changesRe.FindStringSubmatch(line); len(m) > 1 {
				r.changes = atoi(m[1])
			}

			if m := insertionsRe.FindStringSubmatch(line); len(m) > 1 {
				r.insertions = atoi(m[1])
			}

			if m := deletionsRe.FindStringSubmatch(line); len(m) > 1 {
				r.deletions = atoi(m[1])
			}
		case line != "":
			if repl, ok := mailmap[line]; ok {
				line = repl
			}

			r.email = line
		case strings.TrimSpace(line) == "":
			rs.rs = append(rs.rs, r)
			r = record{}
		}
	}

	cmd.Run()

	fmt.Print(rs)
}

func main() {
	flag.Usage = func() {
		fmt.Println("usage: gr [options]\n\ngr generates reports on git repos\n\nOPTIONS")
		flag.PrintDefaults()
	}
	flag.Parse()

	loadMailMap()
	collectRecords()
}
