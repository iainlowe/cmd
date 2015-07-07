package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
)

var (
	server = flag.Bool("server", false, "run in server mode")
)

func main() {
	flag.Parse()

	if *server {
		listener, err := net.Listen("tcp", ":2222")
		if err != nil {
			panic(err)
		}

		for {
			c, err := listener.Accept()
			if err != nil {
				panic(err)
			}

			log.Println("Received connection from", c.RemoteAddr())

			fmt.Fprint(c, "OK")
			c.Close()
		}
	} else {
		c, err := net.Dial("tcp", flag.Arg(0))
		if err != nil {
			panic(err)
		}
		b, err := ioutil.ReadAll(c)
		if err != nil {
			panic(err)
		}
		fmt.Println("Received string from remote:", string(b))
	}
}
