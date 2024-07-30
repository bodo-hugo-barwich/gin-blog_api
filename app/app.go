package app

import (
	//"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"cxcurrency/controllers"
)

var DATABASE *gorm.DB
//var err error

func ConnectDatabase() (*gorm.DB, error) {
	// Connect to the PostgreSQL database
	dsn := "host=db user=cxcurrency password=secret dbname=cxcurrency sslmode=disable"

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

func Start() {

	// Create the global Database Connection
	DATABASE, _ = ConnectDatabase()

	InitializeDatabase(DATABASE)

	router := gin.Default()

	// Register User Routes
	controllers.RegisterUserRoutes(router)
	// Register Article Routes
	controllers.RegisterArticleRoutes(router)

	router.Run(":3000")
}