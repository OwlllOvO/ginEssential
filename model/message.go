package model

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type Message struct {
	ID         uuid.UUID `gorm:"type:char(36);primary_key"`
	SenderID   uint      `json:"sender_id"`
	ReceiverID uint      `json:"receiver_id"`
	Content    string    `json:"content" gorm:"type:text;not null"`
	CreatedAt  Time      `json:"created_at" gorm:"type:timestamp"`
}

func (message *Message) BeforeCreate(tx *gorm.DB) (err error) {
	message.ID = uuid.NewV4()
	return nil
}
