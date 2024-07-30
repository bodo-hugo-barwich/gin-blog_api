package controllers

import (
	"fmt"
	"strconv"
	"net/http"

	"cxcurrency/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DATABASE *gorm.DB

func MigrateUsers(db *gorm.DB) error {

	if DATABASE == nil {

		//Copy reference to global database
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

	// Define routes that interact with the database
	engine.GET("/users", DisplayUsers)
	engine.GET("/users/:id", DisplayUser)
	engine.POST("/users", CreateUser)
}

func DisplayUser(c *gin.Context) {
	var user *model.User
	var userId uint64
	var err error

	userIdString := c.Params.ByName("id")

	fmt.Printf("Controller 'Users': User ID 0: %#v\n", userIdString)

	if userId, err = strconv.ParseUint(userIdString, 10, 64); err != nil {
		c.JSON(http.StatusUnprocessableEntity, struct {
			Title            string
			StatusCode       uint
			Page             string
			ErrorMessage     string
			ErrorDescription string
		}{
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

		c.JSON(http.StatusNotFound, struct {
			Title            string
			StatusCode       uint
			Page             string
			ErrorMessage     string
			ErrorDescription string
		}{
			"Blog - Error",
			http.StatusNotFound,
			"users",
			"Not Found",
			desc,
		})

		return
	}

	c.JSON(200, *user)
}

func DisplayUsers(c *gin.Context) {
	var users []model.User

	DATABASE.Find(&users)

	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var user model.User

	c.BindJSON(&user)

	DATABASE.Create(&user)

	c.JSON(http.StatusOK, user)
}

func GetUserByID(userID uint) (*model.User, error) {
	var users []model.User
	var match *model.User
	var err error

	DATABASE.Find(&users, []uint{userID})

	fmt.Printf("Controller 'Users': GetUserByID(%d) (Count: '%d'): %#v\n", userID, len(users), users)

	if len(users) != 0 {
		match = &users[0]
	} else {
		err = fmt.Errorf("User (ID: '%d'): User does not exist", userID)
	}

	return match, err
}
