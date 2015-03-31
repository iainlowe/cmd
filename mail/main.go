/* The mail command is a drop-in replacement for bsd-mailx(1) that only supports sending email.
 */
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/smtp"
	"os"
	"os/user"
	"strings"
	"time"
)

var from string
var rcpt []string
var subject string
var verbose bool
var skipempty bool

func init() {
	flag.StringVar(&from, "from", "", "The email address that the message is being sent from")
	flag.StringVar(&subject, "s", "", "The subject of the email")
	flag.BoolVar(&verbose, "v", false, "Be as verbose as possible (enable all logging)")
	flag.BoolVar(&skipempty, "e", false, "Don't send empty mails. If the body is empty skip the mail.")
}

func (m *Message) send(to string) error {
	var mxhost string
	var c *smtp.Client
	var err error

	if mxhost, err = getMXHost(to); err != nil {
		return fmt.Errorf("failed to find MX record for %s: %s", to, err)
	}

	if c, err = smtp.Dial(mxhost + ":25"); err != nil {
		return fmt.Errorf("failed to connect to host (%s:25): %s", mxhost, err)
	}

	log.Printf("sending via %s\n", mxhost)

	if err = c.Mail(m.From); err != nil {
		return fmt.Errorf("error occurred on %s while sending MAIL FROM for \"%s\": %s", mxhost, m.From, err)
	}

	if err = c.Rcpt(to); err != nil {
		return fmt.Errorf("error occurred on %s while sending RCPT TO for \"%s\": %s", mxhost, to, err)
	}

	var w io.WriteCloser

	if w, err = c.Data(); err != nil {
		return fmt.Errorf("error occurred on %s while sending start of DATA: %s", mxhost, err)
	}

	defer w.Close()
	defer c.Quit()

	if _, err = fmt.Fprintf(w, "%s", strings.Replace(m.Body, "@@@TO@@@", to, -1)); err != nil {
		return fmt.Errorf("error occurred on %s while writing body of DATA: %s", mxhost, err)

	}

	return nil
}

func (m *Message) Send() {
	for _, to := range m.Rcpt {
		if err := m.send(to); err != nil {
			log.Printf("error delivering mail to %s: %s\n", to, err)
		}
	}
}

type Message struct {
	From string
	Rcpt []string
	Body string
}

func NewMessage(from string, rcpt []string, subject, body string) *Message {
	return &Message{
		From: from,
		Rcpt: rcpt,
		Body: formatBody(from, subject, body),
	}
}

func main() {
	flag.Parse()

	if from == "" {
		u, err := user.Current()

		if err != nil {
			log.Fatalln(err)
		}

		h, err := os.Hostname()

		if err != nil {
			log.Fatalln(err)
		}

		from = u.Username + "@" + h
	}

	rcpt = flag.Args()

	if !verbose {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetFlags(0)
	}

	if len(rcpt) == 0 {
		log.Fatalln("abort: you must specify at least one recipient")
	}

	var body []byte
	var err error

	if body, err = ioutil.ReadAll(os.Stdin); err != nil && err != io.EOF {
		log.Fatalln("error while reading message body from STDIN:", err)
	}

	if skipempty && len(strings.TrimSpace(string(body))) == 0 {
		return
	}

	NewMessage(from, rcpt, subject, string(body)).Send()
}

// Creates a minimally RFC822-compliant email body
func formatBody(from string, subject string, message string) string {
	var body = ""

	body += fmt.Sprintf("Message-ID: <%d%s@localhost>\r\n", time.Now().Unix(), "abc")
	body += fmt.Sprintf("From: %s\r\n", from)
	body += "To: @@@TO@@@\r\n"
	body += fmt.Sprintf("Subject: %s\r\n", subject)
	body += fmt.Sprintf("Date: %s\r\n\r\n", time.Now().Format(time.RFC822))
	body += message

	return body
}

// Returns an MX host for sending mail to the supplied email address
func getMXHost(email string) (mxhost string, err error) {
	mx, err := net.LookupMX(strings.Split(email, "@")[1])

	if err != nil {
		return
	}

	mxhost = mx[0].Host

	return
}
