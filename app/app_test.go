package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-blog/config"
	"gin-blog/model"
)

// testUser - A editor user which will own the articles
var testUser model.User = model.User{
	Name:  "Test User No. 1",
	Slug:  "user-1",
	Login: "user-1",
	Email: "user-1@email.com",
}

// testArticles - Articles that will be created as test data
var testArticles = []model.Article{
	{
		UserID:  0,
		Title:   "Test Article No. 1",
		Slug:    "article-1",
		Content: "Test Article No. 1 Content",
	},
	{
		UserID:  0,
		Title:   "Test Article No. 2",
		Slug:    "article-2",
		Content: "Test Article No. 2 Content",
	},
	{
		UserID:  0,
		Title:   "Test Article No. 3",
		Slug:    "article-3",
		Content: "Test Article No. 3 Content",
	},
}

var resDisplayedArticles map[uint]*model.DisplayedArticle = make(map[uint]*model.DisplayedArticle)

func TestArticleRoutes(t *testing.T) {
	var appConfig config.AppConfig
	var db *gorm.DB
	var err error

	gin.SetMode(gin.TestMode)

	if appConfig, err = config.ReadConfigFile(); err != nil {
		t.Fatalf("Application Configuration: Configuration is missing! Message: %#v", err)
	}

	if db, err = ConnectDatabase(&appConfig); err != nil {
		t.Fatalf("Database Connection: Connection failed! Message: %#v", err)
	}

	if err = InitializeDatabase(db); err != nil {
		t.Fatalf("Database Setup: Setup failed! Message: %#v", err)
	}

	router := RegisterRoutes(&appConfig)

	//-------------------------------------
	// Create test data

	// Create 1 test user
	db.Create(&testUser)

	// Create 3 articles
	for idx, article := range testArticles {
		// Set the user id of the test user
		article.UserID = testUser.ID

		db.Create(&article)

		// Set assigned article id
		testArticles[idx].ID = article.ID
	}

	//-------------------------------------
	// Test Article List

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", appConfig.WebRoot+"articles", nil)
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d'; expected 200", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var displayedArticles []model.DisplayedArticle

	err = json.Unmarshal(res.Body.Bytes(), &displayedArticles)

	if err != nil {
		t.Errorf("Request %s '%s ? %s': Response is invalid JSON! Message: %#v", req.Method, req.URL.Path, req.URL.RawQuery, err)
	}

	for idx, displayed := range displayedArticles {
		// Build the Article Lookup Map
		resDisplayedArticles[displayed.ID] = &displayedArticles[idx]
	}

	for _, article := range testArticles {
		if displayed, ok := resDisplayedArticles[article.ID]; !ok || displayed == nil {
			t.Errorf("Article (%d) '%s': Article is not in Response!", article.ID, article.Slug)
		}
	}

	//-------------------------------------
	// Test one Article

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("%sarticles/%d", appConfig.WebRoot, testArticles[1].ID), nil)
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d'; expected 200", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var displayedArticle model.DisplayedArticle

	err = json.Unmarshal(res.Body.Bytes(), &displayedArticle)

	if err != nil {
		t.Errorf("Request %s '%s ? %s': Response is invalid JSON! Message: %#v", req.Method, req.URL.Path, req.URL.RawQuery, err)
	}

	if displayedArticle.ID != testArticles[1].ID || displayedArticle.Slug != testArticles[1].Slug {
		t.Errorf("Article (%d) '%s': Article is not in Response!", testArticles[1].ID, testArticles[1].Slug)
	}

	//-------------------------------------
	// Clean up test data

	// Delete Test Articles
	for _, article := range testArticles {
		db.Delete(&article, article.ID)
	}

	// Delete Test User
	db.Delete(&testUser, testUser.ID)
}
