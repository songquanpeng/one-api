package common

import (
	"fmt"
	"net/smtp"
	"strings"
)

func SendEmail(subject string, receiver string, content string) error {
	mail := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s<%s>\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n",
		receiver, SystemName, SMTPFrom, subject, content))
	auth := smtp.PlainAuth("", SMTPAccount, SMTPToken, SMTPServer)
	addr := fmt.Sprintf("%s:%d", SMTPServer, SMTPPort)
	to := strings.Split(receiver, ";")
	err := smtp.SendMail(addr, auth, SMTPAccount, to, mail)
	return err
}
