package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"gin-blog/controllers"
)

var DATABASE *gorm.DB

//var err error

func ConnectDatabase(config *AppConfig) (*gorm.DB, error) {
	// Connect to the PostgreSQL database
	dsn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable",
		config.DB.Host, config.DB.Name, config.DB.User, config.DB.Password)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database")
	}
	return db, err
}

func InitializeDatabase(db *gorm.DB) error {

	// Create Users Structure
	err := controllers.MigrateUsers(db)

	if err == nil {
		// Create Articles Structure
		err = controllers.MigrateArticles(db)
	}

	return err
}

func RegisterRoutes() *gin.Engine {
	router := gin.Default()

	// Register User Routes
	controllers.RegisterUserRoutes(router)
	// Register Article Routes
	controllers.RegisterArticleRoutes(router)
	// Register Login Routes
	controllers.RegisterLoginRoutes(router)

	return router
}

func Start() {

	config, err := ReadConfigFile()

	fmt.Printf("App - Start(): config: %#v; error: %#v\n", config, err)

	if err != nil {
		fmt.Printf("App - Start(): Config is missing! Message: %v\n", err)

		return
	}

	// Create the global Database Connection
	DATABASE, err = ConnectDatabase(&config)

	if err != nil {
		fmt.Printf("App - Start(): Database Connection failed! Message: %v\n", err)

		return
	}

	InitializeDatabase(DATABASE)

	router := RegisterRoutes()

	router.Run(":3000")
}
