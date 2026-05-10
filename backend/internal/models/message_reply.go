package models

import "time"

// MessageReply stores every reply sent for a contact message.
// One message can have many replies — shown as a chat history.
//
// Delivery is async: the HTTP handler saves a row with status="pending"
// and returns immediately; a background worker picks it up, calls SMTP,
// and updates the row to "sent" or "failed" with the wrapped error.
type MessageReply struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID   uint   `gorm:"not null;index"           json:"message_id"`
	ReplyText   string `gorm:"type:text;not null"       json:"reply_text"`
	FromEmail   string `gorm:"type:varchar(255)"        json:"from_email"`
	Attachments string `gorm:"type:text"                json:"attachments"` // JSON array of {name,url,type,size}

	// Delivery state. "pending" right after enqueue, "sent" once SMTP returns
	// 250 OK, "failed" if SMTP errored. Frontend renders one tick / two ticks /
	// warning icon based on this.
	//
	// Deliberately NO column-level default: the AutoMigrate that adds this
	// column would otherwise stamp every legacy reply as "pending", and the
	// startup recovery loop would then re-send all of them. New rows get the
	// value set explicitly by the Reply handler. Legacy rows stay empty and
	// the frontend treats empty as "sent" (the synchronous-send era invariant).
	DeliveryStatus string     `gorm:"type:varchar(20);index" json:"delivery_status"`
	DeliveryError  string     `gorm:"type:text"              json:"delivery_error,omitempty"`
	DeliveredAt    *time.Time `                              json:"delivered_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}
