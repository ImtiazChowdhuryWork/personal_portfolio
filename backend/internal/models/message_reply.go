package models

import "time"

// MessageReply stores every reply sent for a contact message.
// One message can have many replies — shown as a chat history.
type MessageReply struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	MessageID   uint      `gorm:"not null;index"           json:"message_id"`
	ReplyText   string    `gorm:"type:text;not null"       json:"reply_text"`
	FromEmail   string    `gorm:"type:varchar(255)"        json:"from_email"`
	Attachments string    `gorm:"type:text"                json:"attachments"` // JSON array of {name,url,type,size}
	CreatedAt   time.Time `                                json:"created_at"`
}
