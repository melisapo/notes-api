package model

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID        string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"  json:"id"`
	Content   string         `gorm:"not null"                                        json:"content"`
	Drawing   *string        `gorm:"type:text"                                       json:"drawing,omitempty"`
	BgColor   string         `gorm:"type:char(7);default:'#f7f19e'"					 json:"bg"` // pastel yellow
	FgColor   string         `gorm:"type:char(7);default:'#1e1f24'"					 json:"fg"` // dark gray
	Likes     int            `gorm:"default:0"                                       json:"likes"`
	Status    string         `gorm:"default:pending"                                 json:"status,omitempty"`
	CreatedAt time.Time      `                                                       json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                           json:"-"`
}

type PostLike struct {
	PostID    string    `gorm:"type:uuid;primaryKey"`
	IPAddress string    `gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
}

type PostReport struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PostID    string    `gorm:"type:uuid;not null"`
	Reason    *string   `gorm:"type:text"`
	CreatedAt time.Time `json:"created_at"`
}
