package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/terminal"
)

type Server struct {
	ssh.ServerConfig
}

func errorize(e error) {
	log.Println(e)
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

func (s *Server) Serve(c net.Conn) error {
	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	_, chans, reqs, err := ssh.NewServerConn(c, &s.ServerConfig)
	if err != nil {
		return err
	}
	// The incoming Request channel must be serviced.
	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Println("abort: could not accept channel")
			continue
		}

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "shell" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				ok := false
				switch req.Type {
				case "shell":
					ok = true
					if len(req.Payload) > 0 {
						log.Println(c.RemoteAddr(), string(req.Payload))
						// We don't accept any
						// commands, only the
						// default shell.
						ok = false
					}
				case "pty-req":
					ok = true
					log.Println(c.RemoteAddr(), "requested PTY")
				}

				req.Reply(ok, nil)
			}
		}(requests)

		term := terminal.NewTerminal(channel, "$ ")

		go func() {
			defer channel.Close()
			for {
				line, err := term.ReadLine()
				if err != nil {
					break
				}
				log.Println(c.RemoteAddr(), line)
			}
			log.Println(c.RemoteAddr(), "has disconnected")
		}()
	}

	return nil
}

func passwordCallback(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	if c.User() == "root" {
	 	log.Printf("%s login successful for %s with pass '%s'\n", c.RemoteAddr(), c.User(), string(pass))
	 	return nil, nil
	}
	return nil, fmt.Errorf("password rejected for %q '%s'", c.User(), string(pass))

	log.Printf("%s login successful for %s with pass '%s'\n", c.RemoteAddr(), c.User(), string(pass))
	return nil, nil
}

func authLogCallback(conn ssh.ConnMetadata, method string, err error) {
	if err == nil || method == "none" {
		return
	}
	log.Println(conn.RemoteAddr(), err)
}

var logfile string

func init() {
	flag.StringVar(&logfile, "logfile", "-", "The path to the file to be used for logging; the file will be created so it must not exist")
}

func main() {
	flag.Parse()

	if logfile != "-" {
		var f *os.File
		var err error
		if f, err = os.Create(logfile); err != nil {
			panic(err)
		}
		log.SetOutput(f)
	}

	s := NewServer()
	s.ListenAndServe("")
}
