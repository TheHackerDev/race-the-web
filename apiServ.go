package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func StartAPI() error {
	// Set Gin configuration mode
	gin.SetMode(gin.ReleaseMode)

	// Configure & Start the HTTP API server
	router := gin.Default()
	router.GET("/get/config", GetConfig)
	router.POST("/set/config", SetConfig)
	router.POST("/start", APIStart)

	router.Run("127.0.0.1:8000")

	return nil
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
	// Check if the configuration exists
	if len(configuration.Requests) == 0 {
		// No configuration currently exists
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "no configuration set",
		})
		return
	}

	// Send response
	ctx.JSON(http.StatusOK, configuration)
}

// API endpoint to begin the race test using the configuration file already provided.
func APIStart(ctx *gin.Context) {
	// Run race test, returning any initial errors
	err, responses := Start()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("error: %s", err.Error()),
		})
	}

	// Set response values
	ctx.Header("Content-Type", "application/json")
	ctx.Status(http.StatusOK)

	// Manually serialize responses, in order to remove html escaping.
	enc := json.NewEncoder(ctx.Writer)
	enc.SetEscapeHTML(false) // Disable html escaping
	enc.Encode(responses)
}
