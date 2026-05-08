package services

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net"
	"net/smtp"
	"net/textproto"
	"time"
)

type EmailService struct {
	host string
	port string
	user string
	pass string
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

func NewEmailService(host, port, user, pass string) *EmailService {
	return &EmailService{host: host, port: port, user: user, pass: pass}
}

func (s *EmailService) Host() string { return s.host }
func (s *EmailService) Port() string { return s.port }

// Verify connects to the configured SMTP server, performs STARTTLS if offered,
// and attempts authentication with the supplied credentials. It does NOT send
// any mail. Returns nil if Gmail accepts the credentials, a descriptive error
// otherwise (auth failure, network error, TLS error).
func (s *EmailService) Verify(user, pass string) error {
	addr := s.host + ":" + s.port
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("could not reach %s: %v", addr, err)
	}
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp handshake failed: %v", err)
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
			return fmt.Errorf("STARTTLS failed: %v", err)
		}
	}
	auth := smtp.PlainAuth("", user, pass, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}
	return client.Quit()
}

// Send sends a plain-text email, optionally with file attachments.
func (s *EmailService) Send(toName, toEmail, fromEmail, subject, body string, attachments []Attachment) error {
	if s.user == "" || s.pass == "" {
		return fmt.Errorf("SMTP not configured — set Sending Gmail Address and App Password in Profile settings")
	}
	if fromEmail == "" {
		fromEmail = s.user
	}

	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	var raw []byte
	if len(attachments) == 0 {
		// Plain text email — no multipart needed
		raw = []byte(fmt.Sprintf(
			"From: Imtiaz <%s>\r\nTo: %s <%s>\r\nReply-To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
			fromEmail, toName, toEmail, fromEmail, subject, body,
		))
	} else {
		// Multipart/mixed email with attachments
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		headers := fmt.Sprintf(
			"From: Imtiaz <%s>\r\nTo: %s <%s>\r\nReply-To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n\r\n",
			fromEmail, toName, toEmail, fromEmail, subject, writer.Boundary(),
		)

		// Text body part
		textHeader := make(textproto.MIMEHeader)
		textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
		textPart, _ := writer.CreatePart(textHeader)
		textPart.Write([]byte(body))

		// Attachment parts
		for _, att := range attachments {
			attHeader := make(textproto.MIMEHeader)
			attHeader.Set("Content-Type", att.ContentType)
			attHeader.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Filename))
			attHeader.Set("Content-Transfer-Encoding", "base64")
			attPart, _ := writer.CreatePart(attHeader)
			encoded := base64.StdEncoding.EncodeToString(att.Data)
			// Write in 76-char lines as per MIME spec
			for i := 0; i < len(encoded); i += 76 {
				end := i + 76
				if end > len(encoded) {
					end = len(encoded)
				}
				attPart.Write([]byte(encoded[i:end] + "\r\n"))
			}
		}
		writer.Close()

		raw = []byte(headers + buf.String())
	}

	return smtp.SendMail(s.host+":"+s.port, auth, s.user, []string{toEmail}, raw)
}
