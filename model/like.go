package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Like struct {
	ID        uint      `gorm:"primarykey"`
	UserId    uint      `gorm:"primaryKey"`
	PostId    uuid.UUID `gorm:"type:char(36);primaryKey"`
	CreatedAt time.Time
}
