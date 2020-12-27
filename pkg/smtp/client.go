package smtp

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/smtp"
	"strconv"
	"time"
)

func SendEmail(address, sender, recipient, subject, message string) error {
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

	/*
		Date      string
		From      string
		ReplyTo   string
		Subject   string
		MessageID string
		To        Recipients
		Body      Body
	*/
	err = sendHeader(wc, "Subject", subject)
	if err != nil {
		return err
	}
	err = sendHeader(wc, "From", sender)
	if err != nil {
		return err
	}
	err = sendHeader(wc, "Reply-To", sender)
	if err != nil {
		return err
	}
	err = sendHeader(wc, "To", recipient)
	if err != nil {
		return err
	}
	err = sendHeader(wc, "Date", time.Now().String())
	if err != nil {
		return err
	}
	err = sendHeader(wc, "Message-ID", strconv.Itoa(time.Now().Nanosecond()))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(wc, "\r\n")
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

func SendEmailFromFile(address, sender, recipient, file string) error {
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

	body, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(wc, string(body))
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
func sendHeader(wc io.WriteCloser, key, val string) error {
	_, err := fmt.Fprintf(wc, "%s:%s\r\n", key, val)
	if err != nil {
		return err
	}
	return nil
}
