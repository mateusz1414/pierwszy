package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"students/grades"
	"students/loginandregister"
	"students/students"
	"students/teachers"
	"students/user"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	database, err := connection()
	if err != nil {
		log.Fatal("Problem z baza")
		return
	}
	server := gin.Default()

	//cors policy
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	server.Use(cors.New(config))

	//session
	store := cookie.NewStore([]byte("ThisIsGoLanguage"))
	store.Options(sessions.Options{
		Path:   "/",
		MaxAge: 1800,
	})
	server.Use(sessions.Sessions("go_session_id", store))

	//add database
	server.Use(dbMiddleware(database))

	//read configs
	file, err := ioutil.ReadFile("./config/credentials.json")
	if err != nil {
		log.Fatal("Read config problem")
		return
	}

	//directores
	student := server.Group("student")
	{
		student.GET("/:studentID", students.GetStudent)
		student.POST("/sendRequest", authMiddleware("user"), students.RequestStudent)
		student.DELETE("/:studentID", authMiddleware("dean"), students.StudentDelete)
		student.PUT("/:studentID", authMiddleware("dean"), students.StudentChange)
	}
	departament := server.Group("departament")
	{
		departament.GET("/getAll", teachers.GetDepartaments)
	}
	management := server.Group("management")
	{
		management.GET("/applicationList", authMiddleware("dean"), students.ApplicationList)
		management.PUT("/:studentID", authMiddleware("dean"), students.StudentAdd)
		management.DELETE("/:studentID", authMiddleware("dean"), students.WaitingDiscard)
	}
	teacher := server.Group("teacher")
	{
		teacher.GET("/:teacherID", teachers.GetTeacher)
	}
	userGroup := server.Group("user")
	{
		userGroup.POST("login", loginandregister.Login)
		userGroup.GET("oauth/login/:provider", addConfig("credentials", file), loginandregister.OauthLogin)
		userGroup.GET("oauth/authorize/:provider", addConfig("credentials", file), loginandregister.OauthAuthorize)
		userGroup.POST("register", loginandregister.Register)
		userGroup.GET("activation/:jwt", user.Activation)
	}
	grade := server.Group("grade")
	{
		grade.GET("/myGrades", authMiddleware("student"), grades.GetStudentGrades)
		grade.GET("/getAll", authMiddleware("teacher"), grades.GetAllGrades)
		grade.POST("/", authMiddleware("teacher"), grades.AddGrade)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	server.Run(":" + port)
}

func connection() (*gorm.DB, error) {
	serverSQL := "34.107.70.144"
	user := "apiuser"
	password := "apistudents"
	dbname := "college"
	db, err := gorm.Open("mysql", user+":"+password+"@tcp("+serverSQL+")/"+dbname+"?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		return nil, fmt.Errorf("Blad polaczenia z baza danych: %v", err.Error())
	}
	return db, nil
}

func dbMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

func addConfig(name string, config []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(name, config)
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
		isValid, claims := user.TokenValidation(token, func(claims jwt.MapClaims) error {
			if claims["time"] != nil && claims["userid"] != nil && claims["permissions"] != nil && claims["time"].(float64)+1800 > float64(time.Now().Unix()) {
				return nil
			}
			return fmt.Errorf("Invalid token")
		})

		if isValid == false {
			c.JSON(401, gin.H{
				"error": "unauthenticated",
			})
			c.Abort()
			return
		}
		if claims["permissions"] != permission {
			c.JSON(403, gin.H{
				"error": "You dont have permission",
			})
			c.Abort()
			return
		}
		c.Set("userid", int64(claims["userid"].(float64)))
		c.Next()

	}
}
