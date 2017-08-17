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
	if len(os.Args) >= 2 {
		if err := Start(); err != nil {
			fmt.Println(usage)
			outError("[ERROR] %s\n", err)
		}
		os.Exit(0)
	}

	// Set Gin configuration mode
	gin.SetMode(gin.DebugMode)

	// Configure & Start the HTTP API server
	router := gin.Default()
	router.GET("/get/config", GetConfig)
	router.POST("/set/config", SetConfig)
	router.POST("/start", APIStart)

	router.Run("127.0.0.1:8000")
}

// API endpoint to set the configuration options
func SetConfig(ctx *gin.Context) {
	// Validate input
	var config Configuration // temporary variable required for proper validation
	if ctx.BindJSON(&config) != nil {
		// Invalid JSON sent
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid JSON data",
		})
		return
	}

	// Set defaults
	SetDefaults(&config)

	// Assign to global configuration object
	configuration = config

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

func APIStart(ctx *gin.Context) {
	// Run race test, returning any initial errors
	if err := Start(); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("error: %s", err.Error()),
		})
	}

	// TODO: Implement error channel to return errors back to caller immediately.

	// Send response
	ctx.JSON(http.StatusOK, gin.H{
		"message": "test complete, check console for output",
	})

	// TODO: return output as JSON response (be mindful of timeouts)
}

// TODO: Write unit tests for all endpoints
