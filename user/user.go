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

type User struct {
	UserID          int
	Login           string `json:"login"`
	Hashpassword    string
	Permissions     string
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmpassword"`
}

func (s *User) RegisterValidate(database *gorm.DB) error {
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
	database.Table("User").Where("login=?", s.Login).Count(&count)
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

func (s *User) Authentication(database *gorm.DB) error {
	hash := md5.Sum([]byte(s.Password))
	s.Hashpassword = hex.EncodeToString(hash[:])
	result := database.Where("login=? and hashpassword=?", s.Login, s.Hashpassword).First(&s)
	if result.RowsAffected == 0 {
		return fmt.Errorf("invalid")
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
