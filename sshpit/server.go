package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"code.google.com/p/go.crypto/ssh"
)

type Server struct {
	ssh.ServerConfig
}

func NewServer() *Server {
	s := &Server{}

	s.PasswordCallback = passwordCallback
	s.AuthLogCallback = authLogCallback

	privateBytes, err := ioutil.ReadFile("/root/.ssh/id_rsa")
	if err != nil {
		panic("Failed to load private key")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Failed to parse private key")
	}

	s.AddHostKey(private)

	return s
}

func (s *Server) ListenAndServe(addr string) {
	if addr == "" {
		addr = ":22"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic("failed to listen for connection")
	}
	log.Println("listening on", listener.Addr())
	s.Listen(listener)
}

func (s *Server) Listen(l net.Listener) error {
	var c net.Conn
	var err error

	for {
		if c, err = l.Accept(); err != nil {
			fmt.Errorf("err: %s", err)
			continue
		}
		if err = s.Serve(c); err != nil {
			fmt.Errorf("err: %s", err)
			continue
		}
	}
}
