package utils

import (
	"fmt"
	"net/smtp"
	"os"

)


func SendInviteEmail(toEmail, inviteLink string)error{
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	subject := "You have been invited to join a team"

		body := fmt.Sprintf("Click here to accept the invite:\n%s", inviteLink)

		message := []byte("Subject: " + subject + "\r\n\r\n" + body)

		auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)
	return err

}