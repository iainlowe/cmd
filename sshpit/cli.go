package main

import (
	"flag"
	"log"
	"os"
)

var logfile string
var addr string

func init() {
	flag.StringVar(&addr, "addr", ":22", "The IP address and port to listen to")
	flag.StringVar(&logfile, "logfile", "-", "The path to the file to be used for logging; the file will be created so it must not exist")
}

func main() {
	flag.Parse()

	if logfile != "-" {
		var f *os.File
		var err error
		if f, err = os.OpenFile(logfile, os.O_RDWR|os.O_APPEND, 0660); err != nil {
			panic(err)
		}
		log.SetOutput(f)
	}

	s := NewServer()
	s.ListenAndServe(addr)
}
