package teachers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Results struct {
	Status       int
	TotalResults int64
	Teacher      []Teachers
	ErrorCode    string
}

type Outs struct {
	Message    string
	UpdateRows int64
	ErrorCode  string
}

type Subject struct {
	IDSubject int
	Name      string `json:"subject"`
}

type Teachers struct {
	IDTeacher int     `json:"idteacher"`
	Name      string  `json:"name"`
	Surname   string  `json:"surname"`
	IDSubject int     `json:"idsubject"`
	Subject   Subject `gorm:"foreignKey:id_subject;association_foreignkey:id_subject"`
}

func outFunc(status int, mess string, rows int64, errc string, c *gin.Context) {
	outs := Outs{
		Message:    mess,
		UpdateRows: rows,
		ErrorCode:  errc,
	}
	c.JSON(status, outs)
}

func GetAll(c *gin.Context) {
	db, dbBool := c.Get("db")
	result := Results{}
	if dbBool != true {
		result = Results{
			Status:    500,
			ErrorCode: "Database error",
		}
	} else {
		var teachers []Teachers
		database := db.(gorm.DB)
		selectResult := database.Joins("INNER JOIN subjects on teachers.id_subject=subjects.id_subject").Preload("Subject").Find(&teachers)
		result = Results{
			Status:       200,
			TotalResults: selectResult.RowsAffected,
			Teacher:      teachers,
		}
	}
	c.JSON(200, result)
}

func GetOnce(c *gin.Context) {
	var teacher = Teachers{}
	teacherIDString, isEmpty := c.Params.Get("teacherID")
	if !isEmpty {
		outFunc(400, "Nie podano id nauczyciela", 0, "incorrect teacher id", c)
		return
	}
	if teacherIDString == "getAll" {
		GetAll(c)
	} else {
		teacherIdInt, err := strconv.Atoi(teacherIDString)
		if err != nil {
			outFunc(500, "Nieoczekiwany błąd serwera", 0, err.Error(), c)
			return
		}
		teacher.IDTeacher = teacherIdInt
		db, dbBool := c.Get("db")
		if dbBool == false {
			outFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
			return
		}
		database := db.(gorm.DB)
		var teachers = Teachers{}
		result := database.Joins("INNER JOIN subjects on teachers.id_subject=subjects.id_subject").Preload("Subject").Where("id_teacher=?", teacher.IDTeacher).Find(&teachers)
		if result.Error != nil || result.RowsAffected == 0 {
			outFunc(400, "Problem z wyświetleniem nauczyciela", result.RowsAffected, result.Error.Error(), c)
			return
		}
		c.JSON(200, result)
	}

}
