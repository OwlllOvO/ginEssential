package model

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// Message represents a single message in a chat.
type Message struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	ChatID    uuid.UUID `json:"chat_id"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	SenderID  uint      `json:"sender_id"`
	CreatedAt Time      `json:"created_at" gorm:"type:timestamp"`
}

func (message *Message) BeforeCreate(tx *gorm.DB) (err error) {
	message.ID = uuid.NewV4()
	return nil
}
