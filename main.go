package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Main entry function for the program
func main() {
	// Run from command-line if arguments are provided- this means that a configuration file has been provided
	if len(os.Args) == 2 {
		Start()
		os.Exit(0)
	}

	// Set Gin configuration mode
	gin.SetMode(gin.DebugMode)

	// Configure & Start the HTTP API server
	r := gin.Default()
	r.GET("/get/config", GetConfig)
	r.POST("/set/config", SetConfig)

	r.Run("127.0.0.1:8000")
}

// API endpoint to set the configuration options
func SetConfig(ctx *gin.Context) {
	// Validate input
	var config Configuration
	if ctx.BindJSON(&config) != nil {
		// Invalid JSON sent
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid JSON",
		})
		return
	}

	// Send response
	ctx.JSON(http.StatusOK, gin.H{
		"message": "configuration saved",
	})
}

// API endpoint to retrieve the high-level configuration
func GetConfig(ctx *gin.Context) {
	// Validate input

	// Send response
	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("configuration: %v", "TODO"),
	})
}

// TODO: Write unit tests for all endpoints
