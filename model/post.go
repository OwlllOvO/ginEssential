package model

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type Post struct {
	ID         uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	UserId     uint      `json:"user_id" gorm:"not null"`
	User       *User
	CategoryId uint `json:"category_id" gorm:"not null"`
	Category   *Category
	Title      string    `json:"title" gorm:"type:varchar(50); not null"`
	HeadImg    string    `json:"head_img"`
	Content    string    `json:"content" gorm:"type:text;not null"`
	CreatedAt  Time      `json:"created_at" gorm:"type:timestamp"`
	UpdatedAt  Time      `json:"updated_at" gorm:"type:timestamp"`
	Comments   []Comment `json:"comments"`                                          // 关联评论
	Status     string    `json:"status" gorm:"type:varchar(20); default:'Pending'"` // Added status field with a default value of 'Pending'
}

func (post *Post) BeforeCreate(tx *gorm.DB) (err error) {
	post.ID = uuid.NewV4() // 直接赋值，不检查错误
	return nil
}
