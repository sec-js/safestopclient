package models

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net"
	"net/smtp"
	"strings"
)

var auth smtp.Auth

//Request struct
type MailRequest struct {
	From    string
	To      []string
	Subject string
	Body    string
}

func NewMailRequest(to []string, subject string) *MailRequest {
	return &MailRequest{
		To:      to,
		Subject: subject,
	}
}

func (r *MailRequest) SendEmail() (bool, error) {

	email_server := viper.GetString("email_server")
	email_server_without_port, _, _ := net.SplitHostPort(email_server)

	email_username := viper.GetString("email_username")
	email_password := viper.GetString("email_password")
	auth = smtp.PlainAuth("", email_username, email_password, email_server_without_port)
	from_address := ""


	if viper.GetString("env") == "development" {
		from_address = "support@ssc.local"

		headers := make(map[string]string)
		headers["Subject"] = r.Subject
		headers["From"] = from_address
		headers["To"] = strings.Join(r.To, ",")
		headers["Content-Type"] = "text/html"

		var msg bytes.Buffer
		for k, v := range headers {
			msg.WriteString(k + ": " + v + "\r\n")
		}

		msg.WriteString("\r\n")
		msg.WriteString(r.Body)

		c, err := smtp.Dial(email_server)
		if err != nil {
			log.Println(err)
			return false, err
		}

		// To && From
		if err := c.Mail(from_address); err != nil {
			log.Println(err)
			return false, err
		}

		if err := c.Rcpt(strings.Join(r.To, ",")); err != nil {
			log.Println(err)
			return false, err
		}

		// Data
		w, err := c.Data()
		if err != nil {
			log.Panic(err)
			return false, err
		}

		_, err = w.Write(msg.Bytes())
		if err != nil {
			log.Println(err)
			return false, err
		}

		err = w.Close()
		if err != nil {
			log.Println(err)
			return false, err
		}

		c.Quit()
		return true, nil

	} else if viper.GetString("env") == "production" {
		from_address = fmt.Sprintf("support@%s", viper.GetString("domain"))

		headers := make(map[string]string)
		headers["Subject"] = r.Subject
		headers["From"] = from_address
		headers["To"] = strings.Join(r.To, ",")
		headers["Content-Type"] = "text/html"

		var msg bytes.Buffer
		for k, v := range headers {
			msg.WriteString(k + ": " + v + "\r\n")
		}

		msg.WriteString("\r\n")
		msg.WriteString(r.Body)

		tlsconfig := &tls.Config {
			InsecureSkipVerify: true,
			ServerName: email_server,
		}

		c, err := smtp.Dial(email_server)
		if err != nil {
			log.Println(err)
			return false, err
		}

		c.StartTLS(tlsconfig)

		if err := c.Auth(auth); err != nil {
			log.Println(err)
			return false, err
		}

		// To && From
		if err := c.Mail(from_address); err != nil {
			log.Println(err)
			return false, err
		}

		if err := c.Rcpt(strings.Join(r.To, ",")); err != nil {
			log.Println(err)
			return false, err
		}

		// Data
		w, err := c.Data()
		if err != nil {
			log.Panic(err)
			return false, err
		}

		_, err = w.Write(msg.Bytes())
		if err != nil {
			log.Println(err)
			return false, err
		}

		err = w.Close()
		if err != nil {
			log.Println(err)
			return false, err
		}

		c.Quit()
		return true, nil

	}

	return true, nil
}