package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"gin-blog/config"
	"gin-blog/controllers"
)

func ConnectDatabase(config *config.AppConfig) (*gorm.DB, error) {
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

func RegisterRoutes(config *config.AppConfig) *gin.Engine {
	router := gin.Default()

	// Register User Routes
	controllers.RegisterHomeRoute(router, config)
	// Register User Routes
	controllers.RegisterUserRoutes(router, config)
	// Register Article Routes
	controllers.RegisterArticleRoutes(router, config)
	// Register Login Routes
	controllers.RegisterLoginRoutes(router, config)

	return router
}

func Start() error {

	var db *gorm.DB

	appConfig, err := config.ReadConfigFile()

	fmt.Printf("App - Start(): config: %#v; error: %#v\n", appConfig, err)

	if err != nil {
		err = fmt.Errorf("Config is missing! Message: %v\n", err)

		return err
	}

	controllers.PROJECT = appConfig.Project

	if controllers.PROJECT == "" {
		controllers.PROJECT = "Gin Blog API"
	}

	// Create the global Database Connection
	db, err = ConnectDatabase(&appConfig)

	if err != nil {
		err = fmt.Errorf("Database Connection failed! Message: %v\n", err)

		return err
	}

	InitializeDatabase(db)

	router := RegisterRoutes(&appConfig)

	router.Run(":3000")

	return err
}
