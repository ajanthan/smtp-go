package smtp

import (
	"fmt"
	"log"
	"net/smtp"
)

func SendEmail(address, sender, recipient, message string) error {
	// Connect to the remote SMTP server.
	log.Println("connecting to ", address)
	c, err := smtp.Dial(address)
	if err != nil {
		return err
	}

	log.Println("Sender .. ")
	// Set the sender and recipient first
	if err := c.Mail(sender); err != nil {
		return err
	}

	log.Println("Recipient .. ")
	if err := c.Rcpt(recipient); err != nil {
		return err
	}

	log.Println("Content[0] .. ")
	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(wc, message)
	if err != nil {
		return err
	}

	log.Println("Content[1] .. ")
	err = wc.Close()
	if err != nil {
		return err
	}
	log.Println("Content[2] .. ")
	log.Println("QUIT .. ")
	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		return err
	}
	log.Println("End .. ")
	return nil
}
