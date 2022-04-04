package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pianisimo/jwt/routes"
	"log"
	"os"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Could not read .env file")
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1", api1)
	router.GET("/api-2", api2)

	router.Run(":" + port)
}

func api1(context *gin.Context) {
	context.JSON(200, gin.H{"success": "Access granted for api-1"})
}

func api2(context *gin.Context) {
	context.JSON(200, gin.H{"success": "Access granted for api-2"})
}
