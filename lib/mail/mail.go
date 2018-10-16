package mail

import (
	"io"
	"net/smtp"
	"fmt"
	"github.com/and-hom/wwmap/lib/config"
	"crypto/tls"
	"github.com/Sirupsen/logrus"
	"log"
)

func SendMail(conf config.MailSettings, recipients []string, subject string, writeBody func(io.Writer) error) error {
	logrus.Debugf("Settings are %v", conf)
	if conf.Ssl {
		return doSendSSL(conf, recipients, subject, writeBody)
	} else {
		return doSendPlain(conf, recipients, subject, writeBody)
	}
}

func doSendPlain(conf config.MailSettings, recipients []string, subject string, writeBody func(io.Writer) error) error {
	logrus.Debug("Plaintext smtp connect")
	c, err := smtp.Dial(conf.SmtpHostPort())
	if err != nil {
		return err
	}
	defer c.Close()

	if err = auth(c, conf); err != nil {
		return err
	}

	return doSend(c, conf, recipients, subject, writeBody)
}

func doSendSSL(conf config.MailSettings, recipients []string, subject string, writeBody func(io.Writer) error) error {
	logrus.Debug("Ssl smtp connect")
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName: conf.SmtpHost,
	}

	conn, err := tls.Dial("tcp", conf.SmtpHostPort(), tlsconfig)
	if err != nil {
		return err
	}
	defer conn.Close()
	c, err := smtp.NewClient(conn, conf.SmtpHost)
	if err != nil {
		return err
	}
	defer c.Close()

	if err = auth(c, conf); err != nil {
		return err
	}

	return doSend(c, conf, recipients, subject, writeBody)
}

func auth(c *smtp.Client, conf config.MailSettings) error {
	if conf.SmtpUser == "" && conf.SmtpPassword == "" {
		return nil
	}
	logrus.Debug("Authorize")
	auth := smtp.PlainAuth(conf.SmtpIdentity, conf.SmtpUser, conf.SmtpPassword, conf.SmtpHost)
	return c.Auth(auth)
}

func doSend(c *smtp.Client, conf config.MailSettings, recipients []string, subject string, writeBody func(io.Writer) error) error {
	logrus.Debug("Do send message")
	if err := c.Mail(conf.From); err != nil {
		return err
	}

	for _, rcpt := range (recipients) {
		if err := c.Rcpt(rcpt); err != nil {
			log.Panic(err)
		}
	}

	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	_, err = fmt.Fprintf(wc, "Content-Type: text/html; charset=UTF-8\r\nFrom: %s\r\nSubject: %s\r\n\r\n", conf.From, subject)
	if err != nil {
		return err
	}
	return writeBody(wc)
}
