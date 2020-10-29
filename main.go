package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mateusz1414/pierwszy/logowanierejestracja"
	"github.com/mateusz1414/pierwszy/studenci"
	"github.com/mateusz1414/pierwszy/user"

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
		student.DELETE("/", authMiddleware(), studenci.StudentDelete)
		student.PUT("/", authMiddleware(), studenci.StudentChange)
		student.POST("/", authMiddleware(), studenci.StudentAdd)
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

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bearer := c.GetHeader("Authorization")
		split := strings.Split(bearer, "Bearer ")
		if len(split) < 2 {
			c.JSON(401, gin.H{
				"error": "unauthenticated",
			})
			c.Abort()
			return
		}
		token := split[1]
		isValid, userId := user.IsTokenValid(token)

		if isValid == false {
			c.JSON(401, gin.H{
				"error": "unauthenticated",
			})
			c.Abort()
		} else {
			c.Set("userid", userId)
			c.Next()
		}

	}
}
