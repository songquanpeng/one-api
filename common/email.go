package common

import "gopkg.in/gomail.v2"

func SendEmail(subject string, receiver string, content string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", SMTPAccount)
	m.SetHeader("To", receiver)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)
	d := gomail.NewDialer(SMTPServer, 587, SMTPAccount, SMTPToken)
	err := d.DialAndSend(m)
	return err
}
