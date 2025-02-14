package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/daniel1sender/SolveItHub/config"
	"github.com/daniel1sender/SolveItHub/models"
	"github.com/daniel1sender/SolveItHub/utils"
	"github.com/gin-contrib/cors"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

func main() {

	config.ConnectDB()
	defer config.CloseDB()

	var problems []models.Problem

	// Check environment variables
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Fatal("AWS credentials not set in environment variables")
	}

	// Initialize Gin router
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://solve-it-hub-front.vercel.app/"}, // Your frontend's URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Endpoint to upload a file
	r.POST("/upload", func(c *gin.Context) {
		// Get file from the request
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
			return
		}

		// Save the uploaded file locally (optional)
		err = c.SaveUploadedFile(file, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file locally"})
			return
		}

		// Upload the file to S3
		err = utils.UploadFileToS3(file.Filename, os.Getenv("AWS_BUCKET_NAME"), file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file to S3"})
			return
		}

		// Respond with success
		c.JSON(http.StatusOK, gin.H{
			"message": "File uploaded successfully",
			"file":    file.Filename,
		})
	})

	r.GET("/files", func(c *gin.Context) {
		files, err := utils.ListFilesInS3(os.Getenv("AWS_BUCKET_NAME"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"files": files})
	})

	r.POST("/problems", func(c *gin.Context) {
		// Parse form data
		title := c.PostForm("title")
		description := c.PostForm("description")
		if title == "" || description == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Title and description are required"})
			return
		}

		// Handle file uploads
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
			return
		}
		files := form.File["files"]

		var fileNames []string
		for _, file := range files {
			// Save file locally (optional)
			err := c.SaveUploadedFile(file, file.Filename)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file locally"})
				return
			}

			// Upload file to S3
			err = utils.UploadFileToS3(file.Filename, os.Getenv("AWS_BUCKET_NAME"), file.Filename)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file to S3"})
				return
			}

			fileNames = append(fileNames, file.Filename)
		}

		// Create a new problem
		problem := models.Problem{
			ID:          uuid.New().String(),
			Title:       title,
			Description: description,
			Files:       fileNames,
		}

		// Add the problem to the in-memory list
		problems = append(problems, problem)

		// Respond with the created problem
		c.JSON(http.StatusOK, problem)
	})

	r.GET("/problems", func(c *gin.Context) {
		c.JSON(http.StatusOK, problems)
	})

	r.GET("/download/:filename", func(c *gin.Context) {
		filename := c.Param("filename")

		// Download file from S3
		fileBody, err := utils.DownloadFileFromS3(os.Getenv("AWS_BUCKET_NAME"), filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
			return
		}
		defer fileBody.Close()

		// Stream the file to the client
		c.Writer.Header().Set("Content-Disposition", "attachment; filename="+filename)
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		_, err = io.Copy(c.Writer, fileBody)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream file"})
			return
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port:", port)
	log.Fatal(r.Run(":" + port))
}
