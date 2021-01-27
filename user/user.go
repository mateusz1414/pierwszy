package user

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/jinzhu/gorm"
)

var secretcode = []byte("mysecretcode")

type Users struct {
	IDuser          int    `gorm:"column:id_user"`
	Login           string `json:"login"`
	Hashpassword    string
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmpassword"`
	Permission      string
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
	hash := md5.Sum([]byte(s.Password))
	s.Hashpassword = hex.EncodeToString(hash[:])
	result := database.Table("users").Where("login=? and hashpassword=?", s.Login, s.Hashpassword).First(&s)
	if result.RowsAffected == 0 {
		return fmt.Errorf("Taki urzytkownik nie istnieje")
	}
	return nil
}

func (s *Users) GetAuthToken() (string, error) {
	claims := jwt.MapClaims{}
	claims["userid"] = s.IDuser
	claims["time"] = time.Now().Unix()
	claims["permission"] = s.Permission
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
	if claims, ok := tok.Claims.(jwt.MapClaims); ok && tok.Valid && claims["time"].(float64)+3000 > float64(time.Now().Unix()) {
		return true, claims
	} else {
		return false, nil
	}
}
