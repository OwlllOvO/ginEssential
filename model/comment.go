package model

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type Comment struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	PostID    uuid.UUID `json:"post_id" gorm:"type:char(36);not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	CreatedAt Time      `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt Time      `json:"updated_at" gorm:"type:timestamp"`
}

func (comment *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	comment.ID = uuid.NewV4() // 直接赋值，不检查错误
	return nil
}
