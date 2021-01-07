package user

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"github.com/jinzhu/gorm"
)

var secretcode = []byte("mysecretcode")

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
	token := make([]byte, 10)
	rand.Read(token)
	s.Code = generateToken()
	err := database.Select("email", "hashpassword", "permissions", "active", "code").Create(&s)
	if err.Error != nil {
		return fmt.Errorf("Server error")
	}

	return nil

}

func (s *User) Authentication(database *gorm.DB) error {
	hash := md5.Sum([]byte(s.Password))
	s.Hashpassword = hex.EncodeToString(hash[:])
	result := database.Where("email=? and hashpassword=?", s.Email, s.Hashpassword).First(&s)
	if result.RowsAffected == 0 {
		return fmt.Errorf("Invalid")
	}
	return nil
}

func (s *User) GetAuthToken() (string, error) {
	claims := jwt.MapClaims{}
	claims["userid"] = s.UserID
	claims["permissions"] = s.Permissions
	claims["time"] = time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	authToken, err := token.SignedString(secretcode)
	return authToken, err
}

func IsTokenValid(token string) (bool, jwt.MapClaims) {
	tok, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); ok == false {
			return nil, fmt.Errorf("Token nie jest walidowany %v", tok.Header["alg"])
		}
		return secretcode, nil
	})
	if err != nil {
		return false, nil
	}
	if claims, ok := tok.Claims.(jwt.MapClaims); ok && tok.Valid && claims["time"].(float64)+1800 > float64(time.Now().Unix()) {
		return true, claims
		//return true, strconv.Itoa(int(userid.(float64)))
	} else {
		return false, nil
	}
}

func generateToken() string {
	token := make([]byte, 10)
	rand.Read(token)
	return fmt.Sprintf("%x", token)
}

func Activation(c *gin.Context) {
	userLogin := c.Param("userID")
	code := c.Param("code")
	db, ok := c.Get("db")
	if !ok {
		c.JSON(400, gin.H{
			"errorCode": "Database error",
		})
		return
	}
	database := db.(*gorm.DB)
	user := User{}
	database.Where("email=?", userLogin).First(&user)
	if user.Active == 1 {
		c.JSON(400, gin.H{
			"errorCode": "Already activated",
		})
		return
	}
	if user.Code != code {
		c.JSON(400, gin.H{
			"errorCode": "Invalid request",
		})
		return
	}
	user.Active = 1
	database.Model(user).Select("active").Update(&user)
	c.JSON(200, gin.H{
		"errorCode": "",
	})
}
