package model

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	UserID  uint   `json:"user_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type DisplayedArticle struct {
	ID         uint   `json:"id"`
	Author     string `json:"author"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

func NewDisplayedArticle(article *Article) DisplayedArticle {
	return DisplayedArticle{
		article.ID,
		"",
		article.Title,
		article.Content,
		article.CreatedAt.Format(time.RFC3339),
		article.UpdatedAt.Format(time.RFC3339),
	}
}
