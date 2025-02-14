package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/daniel1sender/SolveItHub/config"
	"github.com/daniel1sender/SolveItHub/models"
	"github.com/gin-contrib/cors"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

var problems []models.Problem

func main() {

	config.ConnectDB()
	defer config.CloseDB()

	// Initialize Gin router
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://solve-it-hub-front.vercel.app/"}, // Your frontend's URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.POST("/problems", func(c *gin.Context) {
		// Parse form data
		title := c.PostForm("title")
		description := c.PostForm("description")
		if title == "" || description == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Title and description are required"})
			return
		}

		// Create a new problem
		problem := models.Problem{
			ID:          uuid.New().String(),
			Title:       title,
			Description: description,
		}

		// Add the problem to the in-memory list
		problems = append(problems, problem)

		// Respond with the created problem
		c.JSON(http.StatusOK, problem)
	})

	r.GET("/problems", func(c *gin.Context) {
		c.JSON(http.StatusOK, problems)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port:", port)
	log.Fatal(r.Run(":" + port))
}