package courier

type EmailTemplate interface {
	EmailSubject() (string, error)
	EmailBody() (string, error)
	EmailRecipient() (string, error)
}
