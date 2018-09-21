package models

import (
	"bytes"
	"github.com/spf13/viper"
	"log"
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

func NewMailRequest(to []string, subject, body string) *MailRequest {
	return &MailRequest{
		To:      to,
		Subject: subject,
		Body:    body,
	}
}

func (r *MailRequest) SendEmail() (bool, error) {

	host_address := "127.0.0.1:25"
	from := "support@ssc.local"
	if viper.GetString("ENV") == "development" {
		auth = smtp.PlainAuth("", "", "", host_address)
	} else if viper.GetString("ENV") == "production" {
		host_address = viper.GetString("SES_ADDRESS")
		auth = smtp.PlainAuth("", viper.GetString("SES_USERNAME"), viper.GetString("SES_PASSWORD"), host_address)
		if viper.GetString("domain") == "safestopapp.ca" {
			from = "support@safestopapp.com"
		} else {
			from = "support@safestopapp.ca"
		}
	}

	headers := make(map[string]string)
	headers["Subject"] = r.Subject
	headers["From"] = from
	headers["To"] = strings.Join(r.To, ",")
	headers["Content-Type"] = "text/html"

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(k + ": " + v + "\r\n")
	}

	msg.WriteString("\r\n")
	msg.WriteString(r.Body)

	if err := smtp.SendMail(host_address, auth, from, r.To, msg.Bytes()); err != nil {
		log.Println(err)
		return false, err

	}
	return true, nil
}




//func (r *MailRequest) ParseMailTemplate(templateFileName string, data interface{}) error {
//
//	funcMap := template.FuncMap{"t": T}
//
//	t, err := template.ParseFiles("views/mail/" + templateFileName + ".html")
//	if err != nil {
//		log.Println(err)
//		return err
//	}
//
//	t.Funcs(funcMap)
//
//	buf := new(bytes.Buffer)
//	if err = t.Execute(buf, data); err != nil {
//		log.Println(err)
//		return err
//	}
//	r.body = buf.String()
//	return nil
//}

//func T(locale string, key string, value string, args ...interface{}) template.HTML {
//	return i18n.GetI18n().Default(value).T(locale, key, args...)
//}
