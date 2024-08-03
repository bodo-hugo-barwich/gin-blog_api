package model

import (
	//"fmt"

	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	UserID  uint   `json:"user_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
