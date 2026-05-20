package models

import "time"

// JobKeyword stores a single user-defined priority keyword for job scoring.
type JobKeyword struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Keyword   string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"keyword"`
	CreatedAt time.Time `json:"created_at"`
}

func (JobKeyword) TableName() string {
	return "job_keywords"
}
