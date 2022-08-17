package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"time"

	"github.com/go-mail/mail/v2"
	"github.com/google/uuid"
)

//go:embed "templates"
var templateFS embed.FS

type MailConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	Sender   string
}

type Mailer interface {
	SendFromTemplate(recipient, templateFile string, data interface{}) error
}

type ProductionMailer struct {
	dialer *mail.Dialer
	sender string
}

func New(config *MailConfig) ProductionMailer {
	dialer := mail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	dialer.Timeout = 5 * time.Second

	return ProductionMailer{
		dialer: dialer,
		sender: config.Sender,
	}
}

func (m ProductionMailer) SendFromTemplate(recipient, templateFile string, data interface{}) error {
	msg, err := prepareTemplateMessage(m.sender, recipient, templateFile, data)
	if err != nil {
		return err
	}

	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}
	return nil
}

type MockMailer struct {
	messages []string
	sender   string
}

func NewMock(config *MailConfig) MockMailer {
	return MockMailer{
		sender: config.Sender,
	}
}

func (m MockMailer) SendFromTemplate(recipient, templateFile string, data interface{}) error {
	msg, err := prepareTemplateMessage(m.sender, recipient, templateFile, data)
	if err != nil {
		return err
	}

	dialer := mail.SendFunc(m.sendFunc)

	err = mail.Send(dialer, msg)
	return err
}

func (m *MockMailer) sendFunc(from string, to []string, msg io.WriterTo) error {
	message := new(bytes.Buffer)
	_, err := msg.WriteTo(message)
	if err != nil {
		return err
	}
	m.messages = append(m.messages, message.String())
	return nil
}

func prepareTemplateMessage(sender, recipient, templateFile string, data interface{}) (*mail.Message, error) {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return nil, err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return nil, err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return nil, err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return nil, err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())
	msg.SetHeader("Message-ID", uuid.New().String())

	return msg, nil
}
