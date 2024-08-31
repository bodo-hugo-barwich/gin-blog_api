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

// loginUser - A editor user which will own the articles
var testLoginUser model.User = model.User{
	Name:     "Test Login No. 1",
	Slug:     "login-1",
	Login:    "login-1",
	Email:    "login-1@email.com",
	Password: "login.pass",
}

func TestLogin(t *testing.T) {
	var appConfig config.AppConfig
	var db *gorm.DB
	var authUser *model.User
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

	//controllers.RegisterLoginRoutes(router, &appConfig)

	//-------------------------------------
	// Create Login User

	loginPassword := testLoginUser.Password

	if !strings.HasPrefix(testLoginUser.Password, "*") {
		testLoginUser.Password = model.EncryptPassword(testLoginUser.Password, model.ENCRYPTIONSALT)
	}

	db.Create(&testLoginUser)

	testLoginUser.Password = loginPassword

	//-------------------------------------
	// Test Article Create Route

	token, err = loginUser(router, &testLoginUser, &appConfig, t)

	if err != nil {
		t.Errorf("Login (%d) '%s': failed! Message: %#v", testLoginUser.ID, testLoginUser.Login, err)
	}

	if token != "" {
		authUser, err = controllers.ValidateToken(token)

		if err != nil {
			t.Errorf("Login (%d) '%s': Token is invalid! Message: %#v", testLoginUser.ID, testLoginUser.Login, err)
		}

		if authUser != nil {
			if authUser.ID != testLoginUser.ID {
				t.Errorf("Token User: ID '%d' but expected '%d'", authUser.ID, testLoginUser.ID)
			}

			if authUser.Login != testLoginUser.Login {
				t.Errorf("Token User: Login '%s' but expected '%s'", authUser.Login, testLoginUser.Login)
			}
		} else {
			t.Errorf("Login (%d) '%s': Authorized User is not set!", testLoginUser.ID, testLoginUser.Login)
		}
	} else {
		t.Errorf("Login (%d) '%s': Token is empty!", testEditor.ID, testEditor.Login)
	}

	//-------------------------------------
	// Clean Up test data

	// Delete Test User
	db.Delete(&testLoginUser, testLoginUser.ID)
}

func loginUser(router *gin.Engine, user *model.User, appConfig *config.AppConfig, t *testing.T) (string, error) {
	var loginJSON []byte
	var err error

	controllers.RegisterLoginRoutes(router, appConfig)

	login := model.Login{user.Login, user.Password}

	loginJSON, err = json.Marshal(&login)

	if err != nil {
		t.Errorf("Login '%s': JSON Encoding failed! Message: %#v", user.Login, err)
	}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", appConfig.WebRoot+"login", strings.NewReader(string(loginJSON)))
	req.Header.Add("Content-Type", "application/json")
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d'; expected 200", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var loginResponse controllers.LoginSuccess

	err = json.Unmarshal(res.Body.Bytes(), &loginResponse)

	if err != nil {
		t.Errorf("Request %s '%s ? %s': Response is invalid JSON! Message: %#v", req.Method, req.URL.Path, req.URL.RawQuery, err)
	}

	return loginResponse.Token, err
}
