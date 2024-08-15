package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-blog/model"
)

func TestArticleRoutes(t *testing.T) {
	var config AppConfig
	var db *gorm.DB
	var err error

	gin.SetMode("test")

	if config, err = ReadConfigFile(); err != nil {
		t.Fatalf("Application Configuration: Configuration is missing! Message: %#v", err)
	}

	if db, err = ConnectDatabase(&config); err != nil {
		t.Fatalf("Database Connection: Connection failed! Message: %#v", err)
	}

	if err = InitializeDatabase(db); err != nil {
		t.Fatalf("Database Setup: Setup failed! Message: %#v", err)
	}

	router := RegisterRoutes()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/articles", nil)
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d'; expected 200", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var displayedArticles []model.DisplayedArticle

	err = json.Unmarshal(res.Body.Bytes(), &displayedArticles)

	if err != nil {
		t.Fatalf("URL '%s': Response is invalid JSON! Message: %#v", req.RequestURI, err)
	}
}
