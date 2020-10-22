package main

import (
	"fmt"
	"log"
	"os"
	"pierwszy/logowanierejestracja"
	"pierwszy/studenci"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	database, err := connection()
	if err != nil {
		log.Fatal("Problem z baza")
		return
	}
	server := gin.Default()
	server.Use(dbMiddleware(*database))
	student := server.Group("student")
	{
		student.GET("/", studenci.IndexHandler)
		student.DELETE("/", studenci.StudentDelete)
		student.PUT("/", studenci.StudentChange)
		student.POST("/", studenci.StudentAdd)
	}
	user := server.Group("user")
	{
		user.POST("login", logowanierejestracja.Login)
		user.POST("register", logowanierejestracja.Register)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.Run(":" + port)
}

func connection() (*gorm.DB, error) {
	file := "uczelnia"
	db, err := gorm.Open(sqlite.Open(file), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("Blad polaczenia z baza danych: %v", err.Error())
	}
	return db, nil
}

func dbMiddleware(db gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}
