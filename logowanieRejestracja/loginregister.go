package logowanierejestracja

import (
	"pierwszy/user"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Outs struct {
	Status    int
	Message   string
	ErrorCode string
}

func outFunc(status int, mess string, errc string, c *gin.Context) {
	outs := Outs{
		Status:    status,
		Message:   mess,
		ErrorCode: errc,
	}
	c.JSON(status, outs)
}

func Login(c *gin.Context) {
	userData := user.Users{}
	err := c.ShouldBindJSON(&userData)
	if err != nil {
		outFunc(400, "Niepoprawny format danych", err.Error(), c)
	}
	db, DBbool := c.Get("db")
	if DBbool == false {
		outFunc(500, "Nie znaleziono bazy danych", "Database error", c)
	}
	database := db.(gorm.DB)
	err = userData.Authentication(database)
	if err != nil {
		outFunc(400, "Nie udało się zalogować", err.Error(), c)
	} else {
		outFunc(200, "Poprawnie zalogowano", "", c)
	}
}

func Register(c *gin.Context) {
	userData := user.Users{}
	err := c.ShouldBindJSON(&userData)
	if err != nil {
		outFunc(400, "Niepoprawny format danych", err.Error(), c)
	}
	db, DBbool := c.Get("db")
	if DBbool == false {
		outFunc(500, "Nie znaleziono bazy danych", "Database error", c)
	}
	database := db.(gorm.DB)
	err = userData.RegisterValidate(database)
	if err != nil {
		outFunc(400, "Nie udało się zarejestrować", err.Error(), c)
	} else {
		outFunc(200, "Poprawnie zarjestrowano", "", c)
	}

}
