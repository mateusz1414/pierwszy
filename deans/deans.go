package deans

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Results struct {
	Status       int
	TotalResults int64
	User         []Users
	ErrorCode    string
}

type Outs struct {
	Status     int
	Message    string
	UpdateRows int64
	ErrorCode  string
}

type Users struct {
	IDUser     int    `json:"iduser" gorm:"primary_key"`
	Login      string `json:"login"`
	Permission string `json:"permission"`
}

func outFunc(status int, mess string, rows int64, errc string, c *gin.Context) {
	outs := Outs{
		Status:     status,
		Message:    mess,
		UpdateRows: rows,
		ErrorCode:  errc,
	}
	c.JSON(status, outs)
}

func GetAllUsersWithoutPermission(c *gin.Context) {
	db, dbBool := c.Get("db")
	result := Results{}
	if dbBool != true {
		result = Results{
			Status:    500,
			ErrorCode: "Database error",
		}
	} else {
		var users []Users
		database := db.(gorm.DB)
		selectResult := database.Where("permission is NULL or permission=''").Find(&users)
		result = Results{
			Status:       200,
			TotalResults: selectResult.RowsAffected,
			User:         users,
		}
	}
	c.JSON(200, result)
}

func AddPermission(c *gin.Context) {
	var permission = Users{}

	err := c.ShouldBindJSON(&permission)
	if err != nil {
		outFunc(400, "Podaj poprawny format danych", 0, err.Error(), c)
		return
	}
	db, dbBool := c.Get("db")
	if dbBool == false {
		outFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	if permission.IDUser == 0 {
		outFunc(400, "Nie podano id użytkownika", 0, "Not given user id", c)
		return
	}
	database := db.(gorm.DB)
	result := database.Model(&permission).Update(&permission)
	if result.Error != nil {
		outFunc(400, "Problem z dodaniem permisji", result.RowsAffected, result.Error.Error(), c)
	}
	outFunc(200, "Dodano permisje", result.RowsAffected, "", c)

}
