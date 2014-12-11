package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/kr/pty"

	"code.google.com/p/go.crypto/ssh"
)

func copyAndClose(r io.ReadCloser, w io.WriteCloser) error {
	io.Copy(w, r)
	once.Do(close)
}

func rejectOrAccept(c ssh.NewChannel) (ssh.Channel, <-chan *ssh.Request, error) {
	// Channels have a type, depending on the application level
	// protocol intended. In the case of a shell, the type is
	// "session" and ServerShell may be used to present a simple
	// terminal interface.
	if c.ChannelType() != "session" {
		c.Reject(ssh.UnknownChannelType, "unknown channel type: "+c.ChannelType())
		return
	}

	channel, in, err := c.Accept()

	if err != nil {
		return nil, nil, err
	}

	return channel, in, err
}

type NewChannelHandler struct {
	SessionHandler func(ssh.Channel, <-chan *ssh.Request, error)
}

func (s *Server) Serve(c net.Conn) error {
	var err error

	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	_, chans, reqs, err := ssh.NewServerConn(c, &s.ServerConfig)

	if err != nil {
		return err
	}

	select {
	case req := <-reqs:
		log.Println(req)
	case c := <-chans:
		if ch, reqs, err := rejectOrAccept(c); err == nil {

		}
	}

	// Service the incoming Channel channel.
	for newChannel := range chans {
		channel, requests, err := rejectOrAccept(newChannel)

		// allocate a terminal for this channel
		log.Print("creating pty...")

		f, tty, err := pty.Open()

		if err != nil {
			log.Printf("could not start pty (%s)", err)
			continue
		}

		f.Write([]byte("this is cool"))

		//teardown session
		var once sync.Once
		close := func() {
			channel.Close()

			log.Printf("session closed")
		}

		log.Println("OK now")

		copyAndClose(channel, f)
		copyAndClose(f, channel)

		//pipe session to bash and visa-versa
		go func() {
			log.Println("starting S->C")
			tr := io.TeeReader(channel, f)
			br := bufio.NewReader(tr)
			var e error
			var line []byte

			for e != io.EOF {
				line, _, e = br.ReadLine()
				log.Printf("%s server replied: %s", c.RemoteAddr(), string(line))
			}

			b, err := ioutil.ReadAll(br)
			if err != nil {
				log.Println("err:", err)
			}
			log.Printf("%s server replied: %s", c.RemoteAddr(), string(b))

			once.Do(close)
		}()
		go func() {
			log.Println("starting C->S")
			tr := io.TeeReader(f, channel)
			br := bufio.NewReader(tr)
			line, _, e := br.ReadLine()
			if e != nil {
				log.Println(e)
			}
			log.Printf("%s client: %s\n", c.RemoteAddr(), string(line))
			ioutil.ReadAll(br)
			once.Do(close)
		}()

		go handleSSHRequests(f, c, requests)
	}

	return nil
}

func handleSSHRequests(f *os.File, c net.Conn, in <-chan *ssh.Request) {
	for req := range in {
		ok := false
		switch req.Type {
		case "env":
			ok = true
			log.Println(c.RemoteAddr(), string(req.Payload))
		case "shell":
			ok = true
			if len(req.Payload) > 0 {
				log.Println(c.RemoteAddr(), "PAYLOAD", string(req.Payload))
				// We don't accept any
				// commands, only the
				// default shell.
				ok = false
			}
		case "exec":
			log.Println(c.RemoteAddr(), "exec request:", string(req.Payload))
			ok = true
			fakeShellCommand(f, string(req.Payload))
		case "pty-req":
			// Responding 'ok' here will let the client
			// know we have a pty ready for input
			ok = true
			// Parse body...
			termLen := req.Payload[3]
			termEnv := string(req.Payload[4 : termLen+4])
			w, h := parseDims(req.Payload[termLen+4:])
			SetWinsize(f.Fd(), w, h)
			log.Printf("pty-req '%s'", termEnv)
		case "window-change":
			w, h := parseDims(req.Payload)
			SetWinsize(f.Fd(), w, h)
			continue //no response

		default:
			log.Println(c.RemoteAddr(), "unkown SSH request;", req.Type, string(req.Payload))
		}
		if !ok {
			log.Printf("declining %s request...\n", req.Type)
		}
		req.Reply(ok, nil)
	}
}

var pass = map[string]func(s string) string{
	"uname": func(s string) (r string) {
		r = strings.Replace(s, "helium", "everbor", -1)
		if rp := strings.Split(r, " "); len(rp) > 2 {
			rp[2] = "2.6.1-64-generic"
			r = strings.Join(rp, " ")
		}
		return r
	},
	"echo": func(s string) string {
		return s
	},
}

func fakeShellCommand(f *os.File, cmd string) {
	log.Println(cmd)
	cmdparts := strings.Split(cmd, " ")

	log.Printf("cmd (%s)\n", cmdparts[0])

	if fn, ok := pass[cmdparts[0]]; ok {
		c := exec.Command(cmdparts[0], cmdparts[1:]...)
		b, _ := c.CombinedOutput()
		s := fn(string(b))
		s = strings.Replace(s, "\n", "\r\n", -1)
		sb := []byte(s)
		f.Write(sb)
	} else {
		f.Write([]byte("I'm sorry, I can't do that (" + strings.Join(cmdparts, ", ") + ")right now.\r\n"))
	}
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
