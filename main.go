package main

import (
	"fmt"
	"os"
	"strings"

	"students/deans"
	"students/departaments"
	"students/grades"
	"students/loginandregister"
	"students/students"
	"students/teachers"
	"students/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	database, err := connection()
	if err != nil {
		//log.Fatal("Problem z baza")
		fmt.Println(err.Error(), "d")
		return
	}
	server := gin.Default()
	server.Use(dbMiddleware(*database))
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Authorization"}
	server.Use(cors.New(config))
	student := server.Group("student")
	{
		student.GET("/getAll", students.GetAll)
		student.DELETE("/:studentID", authMiddleware("dean"), students.StudentDelete)
		student.PUT("/:studentID", authMiddleware("dean"), students.StudentChange)
		student.POST("/", authMiddleware("dean"), students.StudentAdd)
	}
	user := server.Group("user")
	{
		user.POST("login", loginandregister.Login)
		user.POST("register", loginandregister.Register)
	}
	teacher := server.Group("teacher")
	{
		teacher.GET("/:teacherID", teachers.GetOnce)
	}
	subject := server.Group("subject")
	{
		subject.GET("/", authMiddleware("teacher"), grades.GetStudentsGradesFromSubject)
	}
	grade := server.Group("grade")
	{
		grade.GET("/getAllOfStudent", authMiddleware("student"), grades.GetGradesForOneStudent)
		grade.GET("/getAllOfStudents", authMiddleware("teacher"), grades.GetStudentsGradesFromSubject)
		grade.POST("/getAll", authMiddleware("teacher"), grades.AddGrade)
	}
	departament := server.Group("departament")
	{
		departament.GET("/", departaments.GetAll)
	}
	dean := server.Group("dean")
	{
		dean.GET("/", deans.GetAllUsersWithoutPermission)
		dean.PUT("/", deans.AddPermission)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	server.Run(":" + port)
}

func connection() (*gorm.DB, error) {
	db, err := gorm.Open("mysql", "user:zaq1@WSX@tcp(34.107.48.244:3306)/uczelnia?charset=utf8&parseTime=True&loc=Local")
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

func authMiddleware(permission string) gin.HandlerFunc {
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
		isValid, claims := user.IsTokenValid(token)

		if isValid == false {
			c.JSON(401, gin.H{
				"error": "unauthenticated",
			})
			c.Abort()
		}
		if claims["permission"] != permission {
			c.JSON(401, gin.H{
				"error": "no access",
			})
			c.Abort()
		} else {
			c.Set("permission", claims["permission"])
			c.Set("userid", int(claims["userid"].(float64)))
			c.Next()
		}

	}
}
