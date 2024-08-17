package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gin-blog/config"
)

func RegisterHomeRoute(engine *gin.Engine, config *config.AppConfig) {

	if PROJECT == "" {
		// Copy the Project Name
		PROJECT = config.Project
	}

	// Copy Project Description
	PROJECTDESCRIPTION = config.Description

	// Home Route
	engine.GET(config.WebRoot, DisplayHome)
}

func DisplayHome(c *gin.Context) {
	// Dispatch Home Page
	c.JSON(http.StatusOK,
		HomeSuccess{
			PROJECT + " - Home",
			http.StatusOK,
			"home",
			"OK",
			PROJECTDESCRIPTION,
		})
}
