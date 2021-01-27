package departaments

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Results struct {
	Status       int
	TotalResults int64
	Departament  []Departament
	ErrorCode    string
}

type Outs struct {
	Message    string
	UpdateRows int64
	ErrorCode  string
}

type Departament struct {
	IDDepartament int       `json:"iddepartament" gorm:"primary_key"`
	Name          string    `json:"name"`
	Subjects      []Subject `gorm:"many2many:departaments_subjects;association_foreignkey:id_subject;foreignkey:id_departament;association_jointable_foreignkey:id_subject;jointable_foreignkey:id_departament;"`
}

type Subject struct {
	IDSubject int    `json:"idsubject" gorm:"primary_key"`
	Name      string `json:"name"`
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
		var departament []Departament
		database := db.(gorm.DB)
		selectResult := database.Joins("INNER JOIN departaments_subjects on departaments_subjects.id_departament=departaments.id_departament").Joins("INNER JOIN subjects on departaments_subjects.id_subject=subjects.id_subject").Group("departaments.name").Preload("Subjects").Find(&departament)
		result = Results{
			Status:       200,
			TotalResults: selectResult.RowsAffected,
			Departament:  departament,
		}
	}

	c.JSON(200, result)
}
