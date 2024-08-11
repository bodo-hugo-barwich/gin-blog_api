package model

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	gorm.Model
	UserID  uint   `json:"user_id"`
	Title   string `json:"title"`
	Slug    string `json:"slug"`
	Content string `json:"content"`
}

type DisplayedArticle struct {
	ID         uint   `json:"id"`
	Author     string `json:"author"`
	AuthorSlug string `json:"author_slug"`
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Content    string `json:"content"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

func NewDisplayedArticle(article *Article) DisplayedArticle {
	return DisplayedArticle{
		article.ID,
		"",
		"",
		article.Title,
		article.Slug,
		article.Content,
		article.CreatedAt.Format(time.RFC3339),
		article.UpdatedAt.Format(time.RFC3339),
	}
}

func (article *Article) Update(update *Article) {
	if update.UserID != 0 {
		article.UserID = update.UserID
	}

	if update.Title != "" {
		article.Title = update.Title
	}

	if update.Slug != "" {
		article.Slug = update.Slug
	}

	if update.Content != "" {
		article.Content = update.Content
	}
}
