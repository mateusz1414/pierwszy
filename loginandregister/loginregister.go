package loginandregister

import (
	"students/user"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var serverAdress = "192.168.0.20:8081/"

type Outs struct {
	Message       string `json:"message"`
	ErrorCode     string `json:"errorCode"`
	ActivationURL string `json:"activationURL"`
}

func outFunc(status int, mess string, errc string, c *gin.Context) {
	outs := Outs{
		Message:   mess,
		ErrorCode: errc,
	}
	c.JSON(status, outs)
}

func Login(c *gin.Context) {
	userData := user.User{}
	err := c.ShouldBindJSON(&userData)
	if err != nil {
		outFunc(400, "Invalid data format", err.Error(), c)
		return
	}
	db, DBbool := c.Get("db")
	if DBbool == false {
		outFunc(500, "Database not found", "Database error", c)
		return
	}
	database := db.(*gorm.DB)
	err = userData.Authentication(database)
	if err != nil {
		outFunc(400, "Invalid login or password", err.Error(), c)
	} else {
		if userData.Active == 0 {
			c.JSON(400, gin.H{
				"message":       "Accont is not active",
				"errorCode":     "Not active",
				"activationURL": serverAdress + "user/active/" + userData.Login + "/" + userData.Code,
			})
			return
		}
		token, err := userData.GetAuthToken()
		if err != nil {
			outFunc(500, "Server error", err.Error(), c)
		} else {
			c.JSON(200, gin.H{
				"Message":    "Logged",
				"ErrorCode":  "",
				"Permission": userData.Permissions,
				"AuthToken":  token,
			})
		}
	}
}

func Register(c *gin.Context) {
	userData := user.User{}
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
	database := db.(*gorm.DB)
	err = userData.RegisterValidate(database)
	if err != nil {
		outFunc(400, "Register failed", err.Error(), c)
	} else {
		result := Outs{}
		result.Message = "Poprawnie zarjestrowano"
		result.ActivationURL = serverAdress + "user/active/" + userData.Login + "/" + userData.Code
		c.JSON(200, result)
	}

}
