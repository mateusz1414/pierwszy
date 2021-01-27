package loginandregister

import (
	"students/user"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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
		return
	}
	db, DBbool := c.Get("db")
	if DBbool == false {
		outFunc(500, "Nie znaleziono bazy danych", "Database error", c)
		return
	}
	database := db.(gorm.DB)
	err = userData.Authentication(database)
	if err != nil {
		outFunc(400, "Nie udało się zalogować", err.Error(), c)
	} else {
		permission := userData.Permission
		token, err := userData.GetAuthToken()
		if err != nil {
			outFunc(500, "Problem z pobraniem jwt", err.Error(), c)
		} else {
			c.JSON(200, gin.H{
				"status":     200,
				"Message":    "Poprawnie zalogowano",
				"ErrorCode":  "",
				"AuthToken":  token,
				"Permission": permission,
			})
		}
	}
}

func Register(c *gin.Context) {
	userData := user.Users{}
	err := c.ShouldBindJSON(&userData)
	if err != nil {
		outFunc(400, "Niepoprawny format danych", err.Error(), c)
		return
	}
	db, DBbool := c.Get("db")
	if DBbool == false {
		outFunc(500, "Nie znaleziono bazy danych", "Database error", c)
		return
	}
	database := db.(gorm.DB)
	err = userData.RegisterValidate(database)
	if err != nil {
		outFunc(400, "Nie udało się zarejestrować", err.Error(), c)
	} else {
		outFunc(200, "Poprawnie zarjestrowano", "", c)
	}

}
