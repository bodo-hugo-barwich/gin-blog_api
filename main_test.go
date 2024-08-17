package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-blog/app"
	"gin-blog/config"
	"gin-blog/controllers"
)

func TestHomePage(t *testing.T) {
	var appConfig config.AppConfig
	var db *gorm.DB
	var err error

	gin.SetMode(gin.TestMode)

	if appConfig, err = config.ReadConfigFile(); err != nil {
		t.Fatalf("Application Configuration: Configuration is missing! Message: %#v", err)
	}

	if db, err = app.ConnectDatabase(&appConfig); err != nil {
		t.Fatalf("Database Connection: Connection failed! Message: %#v", err)
	}

	if err = app.InitializeDatabase(db); err != nil {
		t.Fatalf("Database Setup: Setup failed! Message: %#v", err)
	}

	router := app.RegisterRoutes(&appConfig)

	//-------------------------------------
	// Test Home Page

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", appConfig.WebRoot, nil)
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d'; expected 200", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var home controllers.HomeSuccess

	err = json.Unmarshal(res.Body.Bytes(), &home)

	if err != nil {
		t.Errorf("Request %s '%s ? %s': Response is invalid JSON! Message: %#v", req.Method, req.URL.Path, req.URL.RawQuery, err)
	}

	if home.Title != appConfig.Project+" - Home" {
		t.Errorf("Home Title: Title '%s' but expected '%s - Home'", home.Title, appConfig.Project)
	}

	if home.Description != appConfig.Description {
		t.Errorf("Home Description: Description '%s' but expected '%s'", home.Description, appConfig.Description)
	}
}
