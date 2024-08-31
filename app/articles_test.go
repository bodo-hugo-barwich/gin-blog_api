package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-blog/config"
	"gin-blog/controllers"
	"gin-blog/model"
)

// testEditor - A editor user which will own the articles
var testEditor model.User = model.User{
	Name:     "Test Editor No. 1",
	Slug:     "editor-1",
	Login:    "editor-1",
	Email:    "editor-1@email.com",
	Password: "editor.pass",
}

// testArticle - Articles that will be created as test data
var testArticle = model.Article{
	UserID:  0,
	Title:   "Test Article No. 1",
	Slug:    "article-1",
	Content: "Test Article No. 1 Content",
}

func TestCreateArticle(t *testing.T) {
	var appConfig config.AppConfig
	var db *gorm.DB
	var articleJSON []byte
	var token string
	var err error

	gin.SetMode(gin.TestMode)

	if appConfig, err = config.ReadConfigFile(); err != nil {
		t.Fatalf("Application Configuration: Configuration is missing! Message: %#v", err)
	}

	if db, err = ConnectDatabase(&appConfig); err != nil {
		t.Fatalf("Database Connection: Connection failed! Message: %#v", err)
	}

	if err = controllers.MigrateArticles(db); err != nil {
		t.Fatalf("Articles Migration: Migration failed! Message: %#v", err)
	}

	router := gin.Default()

	controllers.RegisterArticleRoutes(router, &appConfig)

	//-------------------------------------
	// Create Editor User

	loginPassword := testEditor.Password

	if !strings.HasPrefix(testEditor.Password, "*") {
		testEditor.Password = model.EncryptPassword(testEditor.Password, model.ENCRYPTIONSALT)
	}

	db.Create(&testEditor)

	testEditor.Password = loginPassword

	testArticle.UserID = testEditor.ID

	//-------------------------------------
	// Test Article Create Route

	token, err = loginUser(router, &testEditor, &appConfig, t)

	if err != nil {
		t.Errorf("Login (%d) '%s': failed! Message: %#v", testEditor.ID, testEditor.Login, err)
	}

	if token == "" {
		t.Errorf("Login (%d) '%s': Token is empty!", testEditor.ID, testEditor.Login)
	}

	articleJSON, err = json.Marshal(&testArticle)

	if err != nil {
		t.Errorf("Article '%s / %s': JSON Encoding failed! Message: %#v", testArticle.Slug, testArticle.Title, err)
	}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", appConfig.WebRoot+"articles", strings.NewReader(string(articleJSON)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d'; expected 200", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var createdArticle model.Article

	err = json.Unmarshal(res.Body.Bytes(), &createdArticle)

	if err != nil {
		t.Errorf("Request %s '%s ? %s': Response is invalid JSON! Message: %#v", req.Method, req.URL.Path, req.URL.RawQuery, err)
	}

	testArticle.ID = createdArticle.ID

	if createdArticle.ID == 0 {
		t.Errorf("Create Article: ID '%d' but not expected '0'", createdArticle.ID)
	}

	if createdArticle.Slug != testArticle.Slug {
		t.Errorf("Create Article: Slug '%s' but expected '%s'", createdArticle.Slug, testArticle.Slug)
	}

	if createdArticle.Title != testArticle.Title {
		t.Errorf("Create Article: Title '%s' but expected '%s'", createdArticle.Title, testArticle.Title)
	}

	//-------------------------------------
	// Clean Up test data

	// Delete Test Article
	db.Delete(&testArticle, testArticle.ID)

	// Delete Test User
	db.Delete(&testEditor, testEditor.ID)
}
