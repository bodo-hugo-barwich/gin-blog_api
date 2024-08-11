package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gin-blog/model"
)

func TestArticleRoutes(t *testing.T) {

	db, _ := ConnectDatabase()

	InitializeDatabase(db)

	router := RegisterRoutes()

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/articles", nil)
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d'; expected 200", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var displayedArticles []model.DisplayedArticle

	err := json.Unmarshal(res.Body.Bytes(), &displayedArticles)

	if err != nil {
		t.Fatalf("URL '%s': Response is invalid JSON! Message: %#v", req.RequestURI, err)
	}
}
