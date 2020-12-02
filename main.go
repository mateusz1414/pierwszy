package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"students/loginandregister"
	"students/students"
	"students/user"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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
		student.GET("/", students.IndexHandler)
		student.DELETE("/:studentID", authMiddleware(), students.StudentDelete)
		student.PUT("/:studentID", authMiddleware(), students.StudentChange)
		student.POST("/", authMiddleware(), students.StudentAdd)
	}
	user := server.Group("user")
	{
		user.POST("login", loginandregister.Login)
		user.POST("register", loginandregister.Register)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.Run(":" + port)
}

func connection() (*gorm.DB, error) {
	file := "uczelnia"
	db, err := gorm.Open("sqlite3", file)
	if err != nil {
		return nil, fmt.Errorf("Blad polaczenia z baza danych: %v", err.Error())
	}
	return db, nil
}

func dbMiddleware(db gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
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
