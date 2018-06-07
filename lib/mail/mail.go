package mail

import (
	"io"
	"net/smtp"
	"fmt"
)

func SendMail(from string, to []string, subject string, writeBody func(io.Writer) error) error {
	c, err := smtp.Dial("localhost:25")
	if err != nil {
		return err
	}
	defer c.Close()

	c.Mail(from)
	for _, rcpt := range (to) {
		c.Rcpt(rcpt)
	}

	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	fmt.Fprintf(wc, "Content-Type: text/html; charset=UTF-8\r\nSubject: %s\r\n", subject)
	return writeBody(wc)
}