package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-blog/config"
	"gin-blog/model"
)

func MigrateArticles(db *gorm.DB) error {

	//Copy reference to global database
	if DATABASE == nil {
		DATABASE = db
	}

	// Automigrate the Article model
	err := db.AutoMigrate(&model.Article{})

	if err != nil {
		fmt.Println("Model 'Article': Auto Migration failed")
	}

	return err
}

func RegisterArticleRoutes(engine *gin.Engine, config *config.AppConfig) {

	if PROJECT == "" {
		// Copy the Project Name
		PROJECT = config.Project
	}

	// Article Routes
	engine.GET(config.WebRoot+"articles", DisplayArticles)
	engine.GET(config.WebRoot+"articles/:id", DisplayArticle)
	engine.PUT(config.WebRoot+"articles/:id", AuthorizeRequest(), UpdateArticle)
	engine.POST(config.WebRoot+"articles", AuthorizeRequest(), CreateArticle)
	engine.DELETE(config.WebRoot+"articles/:id", AuthorizeRequest(), DeleteArticle)
}

func DisplayArticle(c *gin.Context) {
	var article *model.Article
	var displayed model.DisplayedArticle
	var user *model.User
	var articleId uint64
	var err error

	articleIdString := c.Params.ByName("id")

	fmt.Printf("Controller 'Articles': Article ID 0: %#v\n", articleIdString)

	if articleId, err = strconv.ParseUint(articleIdString, 10, 64); err != nil {
		c.JSON(http.StatusUnprocessableEntity,
			APIErrorResponse{
				PROJECT + " - Error",
				http.StatusUnprocessableEntity,
				"articles",
				"Unprocessable Content",
				"Article ID: ID is invalid! Message: " + err.Error(),
			})

		return
	}

	fmt.Printf("Controller 'Articles': Article ID 1: %#v\n", articleId)

	if article, err = GetArticleByID(uint(articleId)); article == nil || err != nil {

		fmt.Printf("Controller 'Articles': Article (ID '%d'): %#v; Error: %#v\n", articleId, article, err)

		desc := fmt.Sprintf("Article (ID: '%d'): Article does not exist", articleId)

		if err != nil {
			desc = err.Error()
		}

		c.JSON(http.StatusNotFound,
			APIErrorResponse{
				PROJECT + " - Error",
				http.StatusNotFound,
				"articles",
				"Not Found",
				desc,
			})

		return
	}

	displayed = model.NewDisplayedArticle(article)

	if user, err = GetUserByID(article.UserID); user == nil || err != nil {

		fmt.Printf("Controller 'Articles': User (ID '%d'): %#v; Error: %#v\n", article.UserID, user, err)

		displayed.Author = "Unknown"
	}

	if user != nil {
		displayed.Author = user.Name
		displayed.AuthorSlug = user.Slug
	}

	c.JSON(http.StatusOK, displayed)
}

func DisplayArticles(c *gin.Context) {
	var articles []model.Article
	var userId int
	var err error

	userIdString := c.Params.ByName("userId")

	if userIdString == "" {
		userIdString = c.Params.ByName("user_id")
	}

	if userIdString != "" {
		if userId, err = strconv.Atoi(userIdString); err != nil {
			c.JSON(http.StatusUnprocessableEntity,
				APIErrorResponse{
					PROJECT + " - Error",
					http.StatusUnprocessableEntity,
					"articles",
					"Unprocessable Content",
					"User ID: ID is invalid! Message: " + err.Error(),
				})

			return
		}

		articles = GetArticlesByUserID(uint(userId))
	} else {
		DATABASE.Find(&articles)
	}

	var displayedArticles []model.DisplayedArticle
	var articleMap map[uint]*model.Article = make(map[uint]*model.Article)
	var userMap map[uint]*model.User = make(map[uint]*model.User)
	var userIDs []uint

	for idx, article := range articles {
		displayedArticles = append(displayedArticles, model.NewDisplayedArticle(&article))
		articleMap[article.ID] = &articles[idx]
	}

	for userID, _ := range userMap {
		userIDs = append(userIDs, userID)
	}

	userRes := GetUsersByIDs(userIDs)

	fmt.Printf("Controller 'Articles': Users: %#v\n", userRes)

	if userRes != nil {
		for idx, user := range *userRes {
			userMap[user.ID] = &(*userRes)[idx]
		}
	}

	fmt.Printf("Controller 'Articles': User Map: %#v\n", userMap)

	for idx, displayed := range displayedArticles {
		article := articleMap[displayed.ID]

		if user, ok := userMap[article.UserID]; ok {
			fmt.Printf("Controller 'Articles': User (ID: '%d'): %#v\n", article.UserID, user)

			displayedArticles[idx].Author = user.Name
			displayedArticles[idx].AuthorSlug = user.Slug
		}
	}

	c.JSON(http.StatusOK, displayedArticles)
}

func CreateArticle(c *gin.Context) {
	var article model.Article
	var user *model.User
	var err error

	editor, ok := c.Get("AuthUser")

	fmt.Printf("Controller 'Articles': Editor: %#v; ok: %#v\n", editor, ok)

	if editor == nil || !ok {
		// Exit on missing Authorized User
		return
	}

	c.BindJSON(&article)

	if article.Slug == "" {
		article.Slug = article.Title
	}

	fmt.Printf("Model 'Article': %#v\n", article)

	if article.UserID == 0 {
		article.UserID = editor.(*model.User).ID
	}

	if article.UserID == 0 {
		c.JSON(http.StatusUnprocessableEntity,
			APIErrorResponse{
				PROJECT + " - Error",
				http.StatusUnprocessableEntity,
				"articles",
				"Unprocessable Content",
				"Model 'Article': User ID is missing!",
			})

		return
	} else {
		if user, err = GetUserByID(article.UserID); err != nil {
			c.JSON(http.StatusNotFound,
				APIErrorResponse{
					PROJECT + " - Error",
					http.StatusNotFound,
					"articles",
					"Not Found",
					err.Error(),
				})

			return
		}
	}

	if user != nil {
		DATABASE.Create(&article)

		c.JSON(http.StatusOK, article)
	}
}

func UpdateArticle(c *gin.Context) {
	var article *model.Article
	var updated model.Article
	var user *model.User
	var articleId uint64
	var err error

	editor, ok := c.Get("AuthUser")

	fmt.Printf("Controller 'Articles': Editor: %#v; ok: %#v\n", editor, ok)

	if editor == nil || !ok {
		// Exit on missing Authorized User
		return
	}

	c.BindJSON(&updated)

	fmt.Printf("Model 'Article': %#v\n", updated)

	articleIdString := c.Params.ByName("id")

	fmt.Printf("Controller 'Articles': Article ID 0: %#v\n", articleIdString)

	if articleId, err = strconv.ParseUint(articleIdString, 10, 64); err != nil {
		c.JSON(http.StatusUnprocessableEntity,
			APIErrorResponse{
				PROJECT + " - Error",
				http.StatusUnprocessableEntity,
				"articles",
				"Unprocessable Content",
				"Article ID: ID is invalid! Message: " + err.Error(),
			})

		return
	}

	fmt.Printf("Controller 'Articles': Article ID 1: %#v\n", articleId)

	if article, err = GetArticleByID(uint(articleId)); article == nil || err != nil {

		fmt.Printf("Controller 'Articles': Article (ID '%d'): %#v; Error: %#v\n", articleId, article, err)

		desc := fmt.Sprintf("Article (ID: '%d'): Article does not exist", articleId)

		if err != nil {
			desc = err.Error()
		}

		c.JSON(http.StatusNotFound,
			APIErrorResponse{
				PROJECT + " - Error",
				http.StatusNotFound,
				"articles",
				"",
				desc,
			})

		return
	}

	if updated.UserID != 0 {
		if user, err = GetUserByID(updated.UserID); user == nil || err != nil {
			if err == nil {
				err = errors.New("User ID: User does not exist!")
			}

			c.JSON(http.StatusUnprocessableEntity,
				APIErrorResponse{
					PROJECT + " - Error",
					http.StatusUnprocessableEntity,
					"articles",
					"Unprocessable Content",
					err.Error(),
				})

			return
		}
	}

	article.Update(&updated)

	DATABASE.Save(&article)

	c.JSON(http.StatusOK, article)
}

func DeleteArticle(c *gin.Context) {
	var article *model.Article
	var articleId uint64
	var message string
	var err error

	editor, ok := c.Get("AuthUser")

	fmt.Printf("Controller 'Articles': Editor: %#v; ok: %#v\n", editor, ok)

	if editor == nil || !ok {
		// Exit on missing Authorized User
		return
	}

	articleIdString := c.Params.ByName("id")

	fmt.Printf("Controller 'Articles': Article ID 0: %#v\n", articleIdString)

	if articleId, err = strconv.ParseUint(articleIdString, 10, 64); err != nil {
		c.JSON(http.StatusUnprocessableEntity,
			APIErrorResponse{
				PROJECT + " - Error",
				http.StatusUnprocessableEntity,
				"articles",
				"Unprocessable Content",
				"Article ID: ID is invalid! Message: " + err.Error(),
			})

		return
	}

	fmt.Printf("Controller 'Articles': Article ID 1: %#v\n", articleId)

	if article, err = GetArticleByID(uint(articleId)); article == nil || err != nil {
		fmt.Printf("Controller 'Articles': Article (ID '%d'): %#v; Error: %#v\n", articleId, article, err)

		message = fmt.Sprintf("Article (ID: '%d'): User does not exist", articleId)
	}

	if article != nil {
		DATABASE.Delete(&article, article.ID)

		message = fmt.Sprintf("Article (ID: '%d'): Article was deleted", article.ID)
	}

	c.JSON(http.StatusOK,
		APIDeleteSuccess{
			PROJECT + " - Delete Success",
			http.StatusOK,
			"users",
			"OK",
			message,
		},
	)
}

func GetArticleByID(articleID uint) (*model.Article, error) {
	var match *model.Article
	var err error

	articleRes := GetArticlesByIDs([]uint{articleID})

	if articleRes == nil {
		return nil, fmt.Errorf("Article (ID: '%d'): Article does not exist!", articleID)
	}

	fmt.Printf("Controller 'Articles': GetArticleByID(%d) (Count: '%d'): %#v\n", articleID, len(*articleRes), *articleRes)

	if len(*articleRes) != 0 {
		match = &(*articleRes)[0]
	} else {
		err = fmt.Errorf("Article (ID: '%d'): Article does not exist!", articleID)
	}

	return match, err
}

func GetArticleBySlug(articleSlug string) (*model.Article, error) {
	var articles []model.Article
	var match *model.Article
	var err error

	DATABASE.Find(&articles, "slug = ?", articleSlug)

	fmt.Printf("Controller 'Articles': GetArticleBySlug('%s') (Count: '%d'):: %#v\n", articleSlug, len(articles), articles)

	if len(articles) != 0 {
		match = &articles[0]
	} else {
		err = fmt.Errorf("Article (Slug: '%s'): Article does not exist!", articleSlug)
	}

	return match, err
}

func GetArticlesByIDs(articleIDs []uint) *[]model.Article {
	var articles []model.Article

	DATABASE.Find(&articles, articleIDs)

	return &articles
}

func GetArticlesByUserID(userID uint) []model.Article {
	var articles []model.Article

	DATABASE.Find(&articles, "user_id = ?", userID)

	fmt.Printf("Controller 'Articles': GetArticlesByUserID(%d): %#v\n", userID, articles)

	return articles
}
