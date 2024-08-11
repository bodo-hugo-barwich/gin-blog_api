package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gin-blog/model"
)

func MigrateUsers(db *gorm.DB) error {

	if DATABASE == nil {
		DATABASE = db
	}

	// Automigrate the User model
	err := db.AutoMigrate(&model.User{})

	if err != nil {
		fmt.Print("Model 'User': Auto Migration failed\n")
	}

	return err
}

func RegisterUserRoutes(engine *gin.Engine) {

	// User Routes
	engine.GET("/users", AuthorizeRequest(), DisplayUsers)
	engine.GET("/users/:id", AuthorizeRequest(), DisplayUser)
	engine.POST("/users", AuthorizeRequest(), CreateUser)
}

func DisplayUser(c *gin.Context) {
	var user *model.User
	var userId uint64
	var err error

	admin, ok := c.Get("AuthUser")

	fmt.Printf("Controller 'Users': Admin: %#v; ok: %#v\n", admin, ok)

	if admin == nil || !ok {
		// Exit on missing Authorized User
		return
	}

	userIdString := c.Params.ByName("id")

	fmt.Printf("Controller 'Users': User ID 0: %#v\n", userIdString)

	if userId, err = strconv.ParseUint(userIdString, 10, 64); err != nil {
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

	fmt.Printf("Controller 'Users': User ID 1: %#v\n", userId)

	if user, err = GetUserByID(uint(userId)); user == nil || err != nil {

		fmt.Printf("Controller 'Users': User (ID '%d'): %#v; Error: %#v\n", userId, user, err)

		desc := "User (ID: '" + userIdString + "'): User does not exist"

		if err != nil {
			desc = err.Error()
		}

		c.JSON(http.StatusNotFound,
			APIErrorResponse{
				"Blog - Error",
				http.StatusNotFound,
				"users",
				"Not Found",
				desc,
			})

		return
	}

	c.JSON(http.StatusOK, *user)
}

func DisplayUsers(c *gin.Context) {
	var users []model.User

	admin, ok := c.Get("AuthUser")

	fmt.Printf("Controller 'Users': Admin: %#v; ok: %#v\n", admin, ok)

	if admin == nil || !ok {
		// Exit on missing Authorized User
		return
	}

	DATABASE.Model(&model.User{}).Select("id, name, login, email").Find(&users)

	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var user model.User

	admin, ok := c.Get("AuthUser")

	fmt.Printf("Controller 'Users': Admin: %#v; ok: %#v\n", admin, ok)

	if admin == nil || !ok {
		// Exit on missing Authorized User
		return
	}

	c.BindJSON(&user)

	if user.Slug == "" {
		user.Slug = user.Name
	}

	if !strings.HasPrefix(user.Password, "*") {
		user.Password = model.EncryptPassword(user.Password, model.ENCRYPTIONSALT)
	}

	DATABASE.Create(&user)

	c.JSON(http.StatusOK, user)
}

func GetUserByID(userID uint) (*model.User, error) {
	var match *model.User
	var err error

	userRes := GetUsersByIDs([]uint{userID})

	if userRes == nil {
		return nil, fmt.Errorf("User (ID: '%d'): User does not exist!", userID)
	}

	fmt.Printf("Controller 'Users': GetUserByID(%d) (Count: '%d'): %#v\n", userID, len(*userRes), *userRes)

	if len(*userRes) != 0 {
		match = &(*userRes)[0]
	} else {
		err = fmt.Errorf("User (ID: '%d'): User does not exist!", userID)
	}

	return match, err
}

func GetUsersByIDs(userIDs []uint) *[]model.User {
	var users []model.User

	DATABASE.Find(&users, userIDs)

	return &users
}

func GetUserByLogin(userLogin string) (*model.User, error) {
	var users []model.User
	var match *model.User
	var err error

	DATABASE.Find(&users, "login = ?", userLogin)

	fmt.Printf("Controller 'Users': GetUserByLogin(%s) (Count: '%d'): %#v\n", userLogin, len(users), users)

	if len(users) != 0 {
		match = &users[0]
	} else {
		err = fmt.Errorf("User (Login: '%s'): User does not exist", userLogin)
	}

	return match, err
}
