package lib

import (
	"fmt"
	"net/smtp"
)

func SendMyMail(subject string) error {
	auth := smtp.PlainAuth(
		"",
		Myconfig.Email.Address,
		Myconfig.Email.Password,
		Myconfig.Email.Server)

	to := []string{Myconfig.Email.Address}

	msg := []byte("To: " + Myconfig.Email.Address + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n")

	err := smtp.SendMail(Myconfig.Email.Server+":587", auth, Myconfig.Email.Address, to, msg)
	if err != nil {
		fmt.Println(err)
	}
	return err
}
