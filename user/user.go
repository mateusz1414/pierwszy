package user

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"gorm.io/gorm"
)

type Users struct {
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

func (s *Users) Authentication() {

}
