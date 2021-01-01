package smtp

type MailReceiver interface {
	Receive(mail *Envelope) error
}
