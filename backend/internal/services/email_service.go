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
//
// Uses a hand-rolled SMTP session (instead of net/smtp.SendMail) so every
// step has an explicit deadline. Without these timeouts a hung Gmail server
// or a network blip causes the call to block indefinitely — which in turn
// makes the dashboard's reply request hang forever and the admin sees a
// permanent "Emailing…" spinner.
func (s *EmailService) Send(toName, toEmail, fromEmail, subject, body string, attachments []Attachment) error {
	if s.user == "" || s.pass == "" {
		return fmt.Errorf("SMTP not configured — set Sending Gmail Address and App Password in Profile settings")
	}
	if fromEmail == "" {
		fromEmail = s.user
	}

	var raw []byte
	if len(attachments) == 0 {
		raw = []byte(fmt.Sprintf(
			"From: Imtiaz <%s>\r\nTo: %s <%s>\r\nReply-To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
			fromEmail, toName, toEmail, fromEmail, subject, body,
		))
	} else {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		headers := fmt.Sprintf(
			"From: Imtiaz <%s>\r\nTo: %s <%s>\r\nReply-To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n\r\n",
			fromEmail, toName, toEmail, fromEmail, subject, writer.Boundary(),
		)

		textHeader := make(textproto.MIMEHeader)
		textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
		textPart, _ := writer.CreatePart(textHeader)
		textPart.Write([]byte(body))

		for _, att := range attachments {
			attHeader := make(textproto.MIMEHeader)
			attHeader.Set("Content-Type", att.ContentType)
			attHeader.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Filename))
			attHeader.Set("Content-Transfer-Encoding", "base64")
			attPart, _ := writer.CreatePart(attHeader)
			encoded := base64.StdEncoding.EncodeToString(att.Data)
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

	// Adaptive deadline: 30s base + 1s per 100 KB of message body. A reply
	// with no attachments stays at ~30s (catches stuck handshakes fast); a
	// 13 MB reply gets ~160s (room for slow residential uploads to Gmail
	// without the deadline biting mid-DATA). Capped at 5 minutes to match
	// the HTTP server's WriteTimeout.
	timeout := 30*time.Second + time.Duration(len(raw)/100_000)*time.Second
	if timeout > 5*time.Minute {
		timeout = 5 * time.Minute
	}
	return s.sendWithTimeout(toEmail, raw, timeout)
}

// sendWithTimeout dials, hand-shakes, authenticates and writes the message
// with a hard deadline on every blocking syscall. Returns a wrapped error
// that names which step timed out so logs are useful.
func (s *EmailService) sendWithTimeout(toEmail string, raw []byte, total time.Duration) error {
	addr := s.host + ":" + s.port
	deadline := time.Now().Add(total)

	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("smtp dial %s: %w", addr, err)
	}
	// Apply the overall deadline to every read/write that follows. The SMTP
	// client's STARTTLS / Auth / Mail / Rcpt / Data calls all bottom out
	// here, so a hung Gmail server returns an error instead of blocking.
	if err := conn.SetDeadline(deadline); err != nil {
		conn.Close()
		return fmt.Errorf("smtp set deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp handshake: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
			return fmt.Errorf("smtp STARTTLS: %w", err)
		}
		// New deadline after the TLS upgrade replaced the underlying conn
		_ = conn.SetDeadline(deadline)
	}

	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(s.user); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	if err := client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("smtp RCPT TO %s: %w", toEmail, err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}
	if _, err := w.Write(raw); err != nil {
		return fmt.Errorf("smtp body write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp body close: %w", err)
	}
	return client.Quit()
}
