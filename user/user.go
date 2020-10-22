package user

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"

	"gorm.io/gorm"
)

var secretcode = []byte("mysecretcode")

type Users struct {
	Iduser          int
	Login           string `json:"login"`
	Hashpassword    string
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmpassword"`
}

func (s *Users) RegisterValidate(database gorm.DB) error {
	if len(s.Login) <= 4 {
		return fmt.Errorf("Login jest zbyt krótki")
	}
	if len(s.Password) <= 4 {
		return fmt.Errorf("Hasło jest zbyt krótkie")
	}
	if s.Password != s.ConfirmPassword {
		return fmt.Errorf("Hasła nie są jednakowe")
	}
	var count int64
	database.Table("users").Where("login=?", s.Login).Count(&count)
	if count != 0 {
		return fmt.Errorf("Taki urzytkownik już istnieje")
	}
	hash := md5.Sum([]byte(s.Password))
	s.Hashpassword = hex.EncodeToString(hash[:])
	err := database.Select("login", "hashpassword").Create(&s)
	if err.Error != nil {
		return fmt.Errorf("Nie udało się dodać urzytkownika")
	}

	return nil

}

func (s *Users) Authentication(database gorm.DB) error {
	var count int64
	hash := md5.Sum([]byte(s.Password))
	s.Hashpassword = hex.EncodeToString(hash[:])
	database.Table("users").Where("login=? and hashpassword=?", s.Login, s.Hashpassword).Count(&count)
	if count == 0 {
		return fmt.Errorf("Taki urzytkownik nie istnieje")
		fmt.Println(s.Password)
	}
	return nil
}

func (s *Users) GetAuthToken() (string, error) {
	claims := jwt.MapClaims{}
	claims["userid"] = s.Iduser
	claims["time"] = time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	authToken, err := token.SignedString(secretcode)
	return authToken, err
}

func IsTokenValid(token string) (bool, string) {
	tok, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); ok == false {
			return nil, fmt.Errorf("Token nie jest walidowany %v", tok.Header["alg"])
		}
		return secretcode, nil
	})
	if err != nil {
		return false, ""
	}
	if claims, ok := tok.Claims.(jwt.MapClaims); ok && tok.Valid && claims["time"].(int64)+300 < time.Now().Unix() {
		userid := claims["userid"]
		return true, userid.(string)
	} else {
		return false, ""
	}
}
