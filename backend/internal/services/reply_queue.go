// ============================================================
// FILE: internal/services/reply_queue.go
// WHAT IT IS:     Background SMTP send queue with a single worker
// WHY IT EXISTS:  Replying via SMTP can take 10–150s for emails with
//                 large attachments on residential connections. Doing
//                 that synchronously in the HTTP handler keeps the
//                 admin staring at a spinner. With this queue, the
//                 handler returns in ~50 ms after persisting the reply
//                 row; a goroutine actually does the SMTP send and
//                 patches the row to sent/failed when it's done. The
//                 dashboard renders the bubble immediately and updates
//                 its tick from SSE events.
// PRODUCERS:      message_handler.Reply, message_handler.RetryReply,
//                 main.go startup recovery loop.
// CONSUMERS:      a single goroutine started in NewReplyQueue.
// LAST UPDATED:   2026-05-10 — initial creation
// ============================================================

package services

import (
	"encoding/json"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

// ReplyJob is what gets pushed onto the queue. We store metadata only —
// attachment bytes are read fresh from disk inside the worker, so a job
// sitting in the buffer doesn't pin tens of MB of memory per reply.
type ReplyJob struct {
	ReplyID   uint   // primary key in message_replies — what we update on send/fail
	ToName    string // visitor's name (for the To: header)
	ToEmail   string // visitor's email
	FromEmail string // which configured Gmail address to send from
	Subject   string
	Body      string
	// AttachmentURLs is the same JSON shape stored in message_replies.attachments.
	// Worker reads the file bytes from disk for entries with /uploads/ paths;
	// link-type entries are skipped (they're already in the body text).
	AttachmentURLs string
}

// ReplyQueue owns a buffered channel of jobs and a single worker goroutine
// that processes them in order. Sequential (not parallel) is fine because
// Gmail SMTP throttles concurrent sends from the same account anyway, and
// keeping it serial means the worker code stays trivial — no rate limiting,
// no per-job context, no fan-out.
type ReplyQueue struct {
	jobs       chan ReplyJob
	emailSvc   *EmailService
	msgSvc     *MessageService
	profileSvc *ProfileService
	broker     *EventBroker
	uploadDir  string

	// closed once Stop() is called so duplicate stops are harmless
	stopOnce sync.Once
	closed   atomic.Bool
}

// NewReplyQueue constructs the queue, starts the worker, and returns a
// pointer the caller can use to enqueue or stop. Buffer size 64 is far
// more than a single admin will ever queue at once; if it fills (would
// only happen if SMTP is broken for a long time), Enqueue marks the
// reply as failed straight away rather than blocking.
func NewReplyQueue(emailSvc *EmailService, msgSvc *MessageService, profileSvc *ProfileService, broker *EventBroker, uploadDir string) *ReplyQueue {
	q := &ReplyQueue{
		jobs:       make(chan ReplyJob, 64),
		emailSvc:   emailSvc,
		msgSvc:     msgSvc,
		profileSvc: profileSvc,
		broker:     broker,
		uploadDir:  uploadDir,
	}
	go q.worker()
	return q
}

// Enqueue pushes a job onto the queue. Non-blocking — if the buffer is
// full (worker stuck or SMTP broken for a long time) the reply is marked
// failed immediately so the dashboard can surface the error rather than
// silently dropping the message.
func (q *ReplyQueue) Enqueue(job ReplyJob) {
	if q.closed.Load() {
		// Server is shutting down — fail fast so the row doesn't sit pending forever.
		_ = q.msgSvc.MarkReplyFailed(job.ReplyID, "server shutting down — please retry")
		q.publish("reply.failed", job.ReplyID, "server shutting down")
		return
	}
	select {
	case q.jobs <- job:
		// queued
	default:
		_ = q.msgSvc.MarkReplyFailed(job.ReplyID, "send queue is full — please retry")
		q.publish("reply.failed", job.ReplyID, "queue full")
	}
}

// Stop closes the channel so the worker drains and exits. Safe to call
// multiple times. Used by graceful shutdown.
func (q *ReplyQueue) Stop() {
	q.stopOnce.Do(func() {
		q.closed.Store(true)
		close(q.jobs)
	})
}

func (q *ReplyQueue) worker() {
	for job := range q.jobs {
		q.process(job)
	}
}

func (q *ReplyQueue) process(job ReplyJob) {
	// Resolve SMTP creds from the profile each send — admin may have rotated
	// the App Password between enqueue and now. Falls back to the env-baked
	// EmailService if the profile has nothing set.
	emailSvc := q.emailSvc
	if profile, err := q.profileSvc.Get(); err == nil && profile.SMTPUser != "" && profile.SMTPPass != "" {
		emailSvc = NewEmailService(
			q.emailSvc.Host(),
			q.emailSvc.Port(),
			profile.SMTPUser,
			profile.SMTPPass,
		)
	}

	attachments := q.loadAttachments(job.AttachmentURLs)

	if err := emailSvc.Send(job.ToName, job.ToEmail, job.FromEmail, job.Subject, job.Body, attachments); err != nil {
		log.Printf("reply queue: send failed for reply %d: %v", job.ReplyID, err)
		_ = q.msgSvc.MarkReplyFailed(job.ReplyID, err.Error())
		q.publish("reply.failed", job.ReplyID, err.Error())
		return
	}

	_ = q.msgSvc.MarkReplySent(job.ReplyID)
	q.publish("reply.sent", job.ReplyID, "")
}

// loadAttachments reads the file bytes for every /uploads/* entry in the
// job's attachment_urls JSON. Mirrors the same logic the synchronous handler
// used to do, just inside the worker so the request handler can return fast.
// Link-type entries are skipped (their URL lives in the body text).
func (q *ReplyQueue) loadAttachments(attUrlsJSON string) []Attachment {
	if attUrlsJSON == "" {
		return nil
	}
	var meta []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(attUrlsJSON), &meta); err != nil {
		return nil
	}

	absUploadDir, _ := filepath.Abs(q.uploadDir)

	var out []Attachment
	for _, m := range meta {
		if m.Type == "link" || m.URL == "" {
			continue
		}
		if !strings.HasPrefix(m.URL, "/uploads/") {
			continue
		}
		rel := strings.TrimPrefix(m.URL, "/uploads/")
		full, err := filepath.Abs(filepath.Join(q.uploadDir, rel))
		if err != nil {
			continue
		}
		// Path-traversal guard: resolved file must still be inside uploadDir.
		if !strings.HasPrefix(full, absUploadDir) {
			continue
		}
		data, err := os.ReadFile(full)
		if err != nil {
			continue
		}
		ct := mime.TypeByExtension(filepath.Ext(m.Name))
		if ct == "" {
			ct = "application/octet-stream"
		}
		out = append(out, Attachment{
			Filename:    m.Name,
			ContentType: ct,
			Data:        data,
		})
	}
	return out
}

// publish wraps the broker call so the worker doesn't have to know about
// event payload shapes. Kept tiny so adding a new field is a one-liner.
func (q *ReplyQueue) publish(eventType string, replyID uint, errMsg string) {
	if q.broker == nil {
		return
	}
	data := map[string]interface{}{
		"reply_id": replyID,
	}
	if errMsg != "" {
		data["error"] = errMsg
	}
	q.broker.Publish(Event{Type: eventType, Data: data})
}
