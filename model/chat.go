package model

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type Chat struct {
	gorm.Model
	ID         uuid.UUID `gorm:"type:char(36);primary_key"`
	SenderID   uint      `gorm:"not null"`
	ReceiverID uint      `gorm:"not null"`
	PostID     uuid.UUID `gorm:"type:char(36);not null"`
	Messages   []Message `gorm:"foreignKey:ChatID"`
}

func (chat *Chat) BeforeCreate(tx *gorm.DB) (err error) {
	chat.ID = uuid.NewV4()
	return
}
