package common

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"errors"
	"net/smtp"
	"strings"
)
type loginAuth struct {
    username, password string
}

func LoginAuth(username, password string) smtp.Auth {
    return &loginAuth{username, password}
}


func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
    return "LOGIN", []byte(a.username), nil
}


func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
    if more {
        switch string(fromServer) {
        case "Username:":
            return []byte(a.username), nil
        case "Password:":
            return []byte(a.password), nil
        default:
            return nil, errors.New("Unknown from server")
        }
    }
    return nil, nil
}

func SendEmail(subject string, receiver string, content string) error {
	if SMTPFrom == "" { // for compatibility
		SMTPFrom = SMTPAccount
	}
	
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
	mail := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s<%s>\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n",
		receiver, SystemName, SMTPFrom, encodedSubject, content))

	//auth := smtp.PlainAuth("", SMTPAccount, SMTPToken, SMTPServer)
	auth := LoginAuth(SMTPAccount, SMTPToken) 
	addr := fmt.Sprintf("%s:%d", SMTPServer, SMTPPort)
	to := strings.Split(receiver, ";")
	var err error
	if SMTPPort == 465 {
		tlsConfig := &tls.Config{
			//InsecureSkipVerify: true,
			ServerName:         SMTPServer,
		}
		//conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", SMTPServer, SMTPPort), tlsConfig)
		conn, err := net.Dial("tcp", "smtp.office365.com:587")
		if err != nil {
			return err
		}
		client, err := smtp.NewClient(conn, SMTPServer)
		if err != nil {
			return err
		}
		if err = client.StartTLS(tlsConfig); err != nil {
    		return err
		}
	

	if err = client.Auth(auth); err != nil {
		return err
	}
		defer client.Close()
		if err = client.Auth(auth); err != nil {
			return err
		}
		if err = client.Mail(SMTPFrom); err != nil {
			return err
		}
		receiverEmails := strings.Split(receiver, ";")
		for _, receiver := range receiverEmails {
			if err = client.Rcpt(receiver); err != nil {
				return err
			}
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(mail)
		if err != nil {
			return err
		}
		err = w.Close()
		if err != nil {
			return err
		}
	} else {
		err = smtp.SendMail(addr, auth, SMTPAccount, to, mail)
	}
	return err
}
