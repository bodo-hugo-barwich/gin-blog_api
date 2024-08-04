package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"cxcurrency/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MigrateArticles(db *gorm.DB) error {

	//Copy reference to global database
	if DATABASE == nil {
		DATABASE = db
	}

	// Automigrate the User model
	err := db.AutoMigrate(&model.Article{})

	if err != nil {
		fmt.Print("Model 'Article': Auto Migration failed\n")
	}

	return err
}

func RegisterArticleRoutes(engine *gin.Engine) {

	// Article Routes
	engine.GET("/articles", DisplayArticles)
	engine.POST("/articles", AuthorizeRequest(), CreateArticle)
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
					"Blog - Error",
					http.StatusUnprocessableEntity,
					"users",
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
		userMap[article.UserID] = nil
	}

	for userID, _ := range userMap {
		userIDs = append(userIDs, userID)
	}

	userRes := GetUsersByIDs(userIDs)

	if userRes != nil {
		for idx, user := range *userRes {
			userMap[user.ID] = &(*userRes)[idx]
		}
	}

	for idx, displayed := range displayedArticles {
		article := articleMap[displayed.ID]

		if user, ok := userMap[article.UserID]; ok {
			displayedArticles[idx].Author = user.Name
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

	fmt.Printf("Model 'Article': %#v\n", article)

	if article.UserID == 0 {
		article.UserID = editor.(*model.User).ID
	}

	if article.UserID == 0 {
		c.JSON(http.StatusUnprocessableEntity,
			APIErrorResponse{
				"Blog - Error",
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
					"Blog - Error",
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

		displayed := model.NewDisplayedArticle(&article)

		displayed.Author = user.Name

		c.JSON(http.StatusOK, displayed)
	}
}

func GetArticlesByUserID(userID uint) []model.Article {
	var articles []model.Article

	DATABASE.Find(&articles, "user_id = ?", userID)

	fmt.Printf("Controller 'Articles': GetArticlesByUserID(%d): %#v\n", userID, articles)

	return articles
}
