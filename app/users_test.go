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

// testAdmin - An admin user which will create the users
var testAdmin model.User = model.User{
	Name:     "Test Admin No. 1",
	Slug:     "admin-1",
	Login:    "admin-1",
	Email:    "admin-1@email.com",
	Password: "admin.pass",
}

// testUsers - Users that will be created as test data
var testUsers = []model.User{
	{
		Name:  "Test User No. 1",
		Slug:  "user-1",
		Login: "user-1",
		Email: "user-1@email.com",
	},
	{
		Name:  "Test User No. 2",
		Slug:  "user-2",
		Login: "user-2",
		Email: "user-2@email.com",
	},
	{
		Name:  "Test User No. 3",
		Slug:  "user-3",
		Login: "user-3",
		Email: "user-3@email.com",
	},
}

var resListUsers map[uint]*model.User = make(map[uint]*model.User)

func TestDisplayUsers(t *testing.T) {
	var appConfig config.AppConfig
	var db *gorm.DB
	var token string
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

	router := gin.Default()

	controllers.RegisterUserRoutes(router, &appConfig)

	//-------------------------------------
	// Create Admin User

	// Reset Admin User ID
	testAdmin.ID = 0

	loginPassword := testAdmin.Password

	if !strings.HasPrefix(testAdmin.Password, "*") {
		testAdmin.Password = model.EncryptPassword(testAdmin.Password, model.ENCRYPTIONSALT)
	}

	createRestoreUser(db, &testAdmin)

	testAdmin.Password = loginPassword

	// Create 3 Users
	for idx, user := range testUsers {
		db.Create(&user)

		// Set assigned user id
		testUsers[idx].ID = user.ID
	}

	//-------------------------------------
	// Get Admin Login Token

	token, err = loginUser(router, &testAdmin, &appConfig, t)

	if err != nil {
		t.Errorf("Login (%d) '%s': failed! Message: %#v", testAdmin.ID, testAdmin.Login, err)
	}

	if token == "" {
		t.Errorf("Login (%d) '%s': Token is empty!", testAdmin.ID, testAdmin.Login)
	}

	//-------------------------------------
	// Test User List

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", appConfig.WebRoot+"users", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d'; expected 200", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var resUsers []model.User

	err = json.Unmarshal(res.Body.Bytes(), &resUsers)

	if err != nil {
		t.Errorf("Request %s '%s ? %s': Response is invalid JSON! Message: %#v", req.Method, req.URL.Path, req.URL.RawQuery, err)
	}

	for idx, displayed := range resUsers {
		// Build the Article Lookup Map
		resListUsers[displayed.ID] = &resUsers[idx]
	}

	for _, user := range testUsers {
		if displayed, ok := resListUsers[user.ID]; !ok || displayed == nil {
			t.Errorf("User (%d) '%s': User is not in Response!", user.ID, user.Slug)
		}
	}

	//-------------------------------------
	// Test one User

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("%susers/%d", appConfig.WebRoot, testUsers[1].ID), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d' but expected '200'", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var resUser model.User

	err = json.Unmarshal(res.Body.Bytes(), &resUser)

	if err != nil {
		t.Errorf("Request %s '%s ? %s': Response is invalid JSON! Message: %#v", req.Method, req.URL.Path, req.URL.RawQuery, err)
	}

	if resUser.ID != testUsers[1].ID || resUser.Slug != testUsers[1].Slug {
		t.Errorf("User (%d) '%s': User is not in Response!", testUsers[1].ID, testUsers[1].Slug)
	}

	//-------------------------------------
	// Clean up test data

	// Delete Test Articles
	for _, user := range testUsers {
		db.Delete(&user, user.ID)
	}

	// Delete Test Admin
	db.Delete(&testAdmin, testAdmin.ID)
}

func TestCreateUser(t *testing.T) {
	var appConfig config.AppConfig
	var db *gorm.DB
	var testUser *model.User = &testUsers[0]
	var userJSON []byte
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

	controllers.RegisterUserRoutes(router, &appConfig)

	//-------------------------------------
	// Create Admin User

	// Reset Admin User ID
	testAdmin.ID = 0

	// Reset Test User ID
	testUser.ID = 0

	loginPassword := testAdmin.Password

	if !strings.HasPrefix(testAdmin.Password, "*") {
		testAdmin.Password = model.EncryptPassword(testAdmin.Password, model.ENCRYPTIONSALT)
	}

	createRestoreUser(db, &testAdmin)

	testAdmin.Password = loginPassword

	//-------------------------------------
	// Test User Create Route

	token, err = loginUser(router, &testAdmin, &appConfig, t)

	if err != nil {
		t.Errorf("Login (%d) '%s': failed! Message: %#v", testAdmin.ID, testAdmin.Login, err)
	}

	if token == "" {
		t.Errorf("Login (%d) '%s': Token is empty!", testAdmin.ID, testAdmin.Login)
	}

	userJSON, err = json.Marshal(testUser)

	if err != nil {
		t.Errorf("User '%s / %s': JSON Encoding failed! Message: %#v", testAdmin.Slug, testAdmin.Name, err)
	}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", appConfig.WebRoot+"users", strings.NewReader(string(userJSON)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d'; expected 200", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var createdUser model.User

	err = json.Unmarshal(res.Body.Bytes(), &createdUser)

	if err != nil {
		t.Errorf("Request %s '%s ? %s': Response is invalid JSON! Message: %#v", req.Method, req.URL.Path, req.URL.RawQuery, err)
	}

	testUser.ID = createdUser.ID

	if createdUser.ID == 0 {
		t.Errorf("Create User: ID '%d' but not expected '0'", createdUser.ID)
	}

	if createdUser.Slug != testUser.Slug {
		t.Errorf("Create User: Slug '%s' but expected '%s'", createdUser.Slug, testUser.Slug)
	}

	if createdUser.Name != testUser.Name {
		t.Errorf("Create User: Name '%s' but expected '%s'", createdUser.Name, testUser.Name)
	}

	//-------------------------------------
	// Clean Up test data

	// Delete Test User
	db.Delete(&testUser, testUser.ID)

	// Delete Test Admin
	db.Delete(&testAdmin, testAdmin.ID)
}

func TestUpdateUser(t *testing.T) {
	var appConfig config.AppConfig
	var db *gorm.DB
	var testUser *model.User = &testUsers[1]
	var userJSON []byte
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

	controllers.RegisterUserRoutes(router, &appConfig)

	//-------------------------------------
	// Create Test Data

	// Reset Admin User ID
	testAdmin.ID = 0

	// Reset Test User ID
	testUser.ID = 0

	loginPassword := testAdmin.Password

	if !strings.HasPrefix(testAdmin.Password, "*") {
		testAdmin.Password = model.EncryptPassword(testAdmin.Password, model.ENCRYPTIONSALT)
	}

	createRestoreUser(db, &testAdmin)

	testAdmin.Password = loginPassword

	// Create Test User
	db.Create(&testUser)

	// Change Test User
	testUser.Slug += "-updated"
	testUser.Name += " - Updated"

	//-------------------------------------
	// Test User Update Route

	token, err = loginUser(router, &testAdmin, &appConfig, t)

	if err != nil {
		t.Errorf("Login (%d) '%s': failed! Message: %#v", testAdmin.ID, testAdmin.Login, err)
	}

	if token == "" {
		t.Errorf("Login (%d) '%s': Token is empty!", testAdmin.ID, testAdmin.Login)
	}

	userJSON, err = json.Marshal(testUser)

	if err != nil {
		t.Errorf("User '%s / %s': JSON Encoding failed! Message: %#v", testUser.Slug, testUser.Name, err)
	}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("%susers/%d", appConfig.WebRoot, testUser.ID), strings.NewReader(string(userJSON)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d'; expected 200", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var updatedUser model.User

	err = json.Unmarshal(res.Body.Bytes(), &updatedUser)

	if err != nil {
		t.Errorf("Request %s '%s ? %s': Response is invalid JSON! Message: %#v", req.Method, req.URL.Path, req.URL.RawQuery, err)
	}

	if updatedUser.Slug != testUser.Slug {
		t.Errorf("Create User: Slug '%s' but expected '%s'", updatedUser.Slug, testUser.Slug)
	}

	if updatedUser.Name != testUser.Name {
		t.Errorf("Create User: Name '%s' but expected '%s'", updatedUser.Name, testUser.Name)
	}

	//-------------------------------------
	// Clean Up test data

	// Delete Test User
	db.Delete(&testUser, testUser.ID)

	// Delete Test Admin
	db.Delete(&testAdmin, testAdmin.ID)
}

func TestDeleteUser(t *testing.T) {
	var appConfig config.AppConfig
	var db *gorm.DB
	var testUser *model.User = &testUsers[2]
	var userJSON []byte
	var token string
	var err error

	gin.SetMode(gin.TestMode)

	if appConfig, err = config.ReadConfigFile(); err != nil {
		t.Fatalf("Application Configuration: Configuration is missing! Message: %#v\n", err)
	}

	if db, err = ConnectDatabase(&appConfig); err != nil {
		t.Fatalf("Database Connection: Connection failed! Message: %#v\n", err)
	}

	if err = controllers.MigrateArticles(db); err != nil {
		t.Fatalf("Articles Migration: Migration failed! Message: %#v\n", err)
	}

	router := gin.Default()

	controllers.RegisterUserRoutes(router, &appConfig)

	//-------------------------------------
	// Create Test Data

	// Reset Admin User ID
	testAdmin.ID = 0

	// Reset Test User ID
	testUser.ID = 0

	loginPassword := testAdmin.Password

	if !strings.HasPrefix(testAdmin.Password, "*") {
		testAdmin.Password = model.EncryptPassword(testAdmin.Password, model.ENCRYPTIONSALT)
	}

	createRestoreUser(db, &testAdmin)

	testAdmin.Password = loginPassword

	// Create Test User
	db.Create(&testUser)

	// Change Test User
	testUser.Slug += "-updated"
	testUser.Name += " - Updated"

	//-------------------------------------
	// Test User Update Route

	token, err = loginUser(router, &testAdmin, &appConfig, t)

	if err != nil {
		t.Errorf("Login (%d) '%s': failed! Message: %#v", testAdmin.ID, testAdmin.Login, err)
	}

	if token == "" {
		t.Errorf("Login (%d) '%s': Token is empty!", testAdmin.ID, testAdmin.Login)
	}

	userJSON, err = json.Marshal(testUser)

	if err != nil {
		t.Errorf("User '%s / %s': JSON Encoding failed! Message: %#v", testUser.Slug, testUser.Name, err)
	}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%susers/%d", appConfig.WebRoot, testUser.ID), strings.NewReader(string(userJSON)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("Request %s '%s ? %s': HTTP Status Code '%d'; expected 200", req.Method, req.URL.Path, req.URL.RawQuery, res.Code)
	}

	fmt.Printf("Request %s '%s ? %s' - Body:\n'%#v'\n", req.Method, req.URL.Path, req.URL.RawQuery, res.Body.String())

	var deleteResponse controllers.APIDeleteSuccess
	expectedDescription := fmt.Sprintf("User (ID: '%d'): User was deleted", testUser.ID)

	err = json.Unmarshal(res.Body.Bytes(), &deleteResponse)

	if err != nil {
		t.Errorf("Request %s '%s ? %s': Response is invalid JSON! Message: %#v", req.Method, req.URL.Path, req.URL.RawQuery, err)
	}

	if deleteResponse.Description != expectedDescription {
		t.Errorf("Delete User (%d) '%s': Description '%s' but expected '%s'", testUser.ID, testUser.Login, deleteResponse.Description, expectedDescription)
	}

	//-------------------------------------
	// Clean Up test data

	// Delete Test User
	db.Delete(&testUser, testUser.ID)

	// Delete Test Admin
	db.Delete(&testAdmin, testAdmin.ID)
}

func createRestoreUser(db *gorm.DB, searchUser *model.User) {
	var resUser *model.User
	var err error

	if controllers.DATABASE == nil {
		controllers.DATABASE = db
	}

	if resUser, err = controllers.GetUserByLogin(searchUser.Login); resUser == nil || err != nil {
		var restore model.User

		db.Unscoped().Where("login = ?", searchUser.Login).Where("deleted_at IS NOT NULL").Find(&restore)

		fmt.Printf("Login '%s / %s': Restore: %#v\n", searchUser.Login, searchUser.Name, restore)

		if restore.ID != 0 {
			// Deleted account become new user
			resUser = &restore

			// Re-enable user account
			db.Model(&restore).Unscoped().Where("id = ?", restore.ID).Update("deleted_at", nil)
		}
	}

	if resUser == nil {
		db.Create(searchUser)

		resUser = searchUser
	}

	// Update original with the fetched or created ID
	searchUser.ID = resUser.ID
}
