package smtp

import (
	"fmt"
	"net/smtp"
)

func SendEmail(address, sender, recipient, message string) error {
	// Connect to the remote SMTP server.
	c, err := smtp.Dial(address)
	if err != nil {
		return err
	}

	// Set the sender and recipient first
	if err := c.Mail(sender); err != nil {
		return err
	}

	if err := c.Rcpt(recipient); err != nil {
		return err
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(wc, message)
	if err != nil {
		return err
	}

	err = wc.Close()
	if err != nil {
		return err
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		return err
	}

	return nil
}
