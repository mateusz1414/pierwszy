package user

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"github.com/jinzhu/gorm"
)

var secretcode = []byte("mysecretcode")

//APIAdress have adress server API
var APIAdress = "http://192.168.0.20:8081/"

//ServerAdress have adress web application
var ServerAdress = "http://192.168.0.20:8080/"

type User struct {
	UserID          int    `gorm:"primary_key"`
	Email           string `json:"email"`
	Hashpassword    string
	Permissions     string `json:"permissions"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmpassword"`
	Active          int
	Code            string
}

func (s *User) RegisterValidate(database *gorm.DB) error {
	match, _ := regexp.Match(`^([a-zA-Z0-9_.])+\@+([a-z-])+\.+([a-z]{2,4})+$`, []byte(s.Email))
	if !match {
		return fmt.Errorf("Invalid email")
	}
	if len(s.Password) <= 4 {
		return fmt.Errorf("Password is to short")
	}
	if s.Password != s.ConfirmPassword {
		return fmt.Errorf("Password do not match")
	}
	var count int64
	database.Table("Users").Where("email=?", s.Email).Count(&count)
	if count != 0 {
		return fmt.Errorf("Email taken")
	}
	hash := md5.Sum([]byte(s.Password))
	s.Hashpassword = hex.EncodeToString(hash[:])
	s.Active = 0
	s.Permissions = "user"
	err := database.Select("email", "hashpassword", "permissions", "active").Create(&s)
	if err.Error != nil {
		return fmt.Errorf("Server error")
	}
	var erro error
	s.Code, erro = CreateJWTToken(jwt.MapClaims{
		"userid": s.UserID,
	})
	return erro

}

func (s *User) Authentication(database *gorm.DB) error {
	hash := md5.Sum([]byte(s.Password))
	s.Hashpassword = hex.EncodeToString(hash[:])
	result := database.Where("email=? and hashpassword=?", s.Email, s.Hashpassword).First(&s)
	if result.RowsAffected == 0 {
		return fmt.Errorf("Invalid")
	}
	var err error = nil
	if s.Active == 0 {
		s.Code, err = CreateJWTToken(jwt.MapClaims{
			"userid": s.UserID,
		})
	}
	return err
}

func CreateJWTToken(c jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	authToken, err := token.SignedString(secretcode)
	return authToken, err
}

func TokenValidation(tokenString string, function func(claims jwt.MapClaims) error) (bool, jwt.MapClaims) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok == false {
			return nil, fmt.Errorf("Token is not valid %v", token.Header["alg"])
		}
		return secretcode, nil
	})
	if err != nil {
		return false, nil
	}
	claims := token.Claims.(jwt.MapClaims)
	err = function(claims)
	if err != nil && token.Valid {
		return false, nil
	}
	return true, claims
}

func generateToken() string {
	token := make([]byte, 10)
	rand.Read(token)
	return fmt.Sprintf("%x", token)
}

func Activation(c *gin.Context) {
	token := c.Param("jwt")
	db, ok := c.Get("db")
	if !ok {
		c.JSON(400, gin.H{
			"errorCode": "Database error",
		})
		return
	}
	database := db.(*gorm.DB)
	isValid, claims := TokenValidation(token, func(claims jwt.MapClaims) error {
		if claims["userid"] != nil {
			return nil
		}
		return fmt.Errorf("Invalid token")
	})
	if isValid == false {
		c.Redirect(307, ServerAdress+"login/invalidToken")
		return
	}
	user := User{
		UserID: int(claims["userid"].(float64)),
	}
	result := database.First(&user)
	if result.RowsAffected == 0 {
		c.Redirect(307, ServerAdress+"login/notFound")
		return
	}
	user.Active = 1
	database.Model(user).Select("active").Update(&user)
	c.Redirect(307, ServerAdress+"login/activated")
}
