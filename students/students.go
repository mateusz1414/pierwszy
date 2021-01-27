package students

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Results struct {
	Status       int
	TotalResults int64
	Student      []Students
	ErrorCode    string
}

type Outs struct {
	Status     int
	Message    string
	UpdateRows int64
	ErrorCode  string
}

type Students struct {
	IDStudent     int    `json:"idstudent" gorm:"primary_key"`
	Name          string `json:"name"`
	Surname       string `json:"surname"`
	Dob           string `json:"dob"`
	IDDepartament int
	Departament   Departament `gorm:"foreignKey:id_departament;association_foreignkey:id_departament"`
	Sex           string      `json:"sex"`
}

type Departament struct {
	IDDepartament int
	Name          string `json:"departament"`
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

/*func IndexHandler(c *gin.Context) {
	db, dbBool := c.Get("db")
	result := Results{}
	if dbBool != true {
		result = Results{
			Status:    500,
			ErrorCode: "Database error",
		}
	} else {
		var students []Students
		database := db.(gorm.DB)
		selectResult := database.Find(&students)
		result = Results{
			Status:       200,
			TotalResults: selectResult.RowsAffected,
			Student:      students,
		}
	}
	c.JSON(200, result)
}*/
func GetAll(c *gin.Context) {
	db, dbBool := c.Get("db")
	result := Results{}
	if dbBool != true {
		result = Results{
			Status:    500,
			ErrorCode: "Database error",
		}
	} else {
		var students []Students
		database := db.(gorm.DB)
		selectResult := database.Joins("INNER JOIN departaments on students.id_departament=departaments.id_departament").Preload("Departament").Find(&students)
		result = Results{
			Status:       200,
			TotalResults: selectResult.RowsAffected,
			Student:      students,
		}
	}
	c.JSON(200, result)
}

func StudentDelete(c *gin.Context) {
	var student = Students{}
	studentIDString, isEmpty := c.Params.Get("studentID")
	if !isEmpty {
		outFunc(400, "Nie podano id studenta", 0, "incorrect student id", c)
		return
	}
	studentIdInt, err := strconv.Atoi(studentIDString)
	if err != nil {
		outFunc(500, "Nieoczekiwany błąd serwera", 0, err.Error(), c)
		return
	}
	err = c.ShouldBindJSON(&student)
	if err != nil {
		outFunc(400, "Podaj poprawny format danych", 0, err.Error(), c)
		return
	}
	student.IDStudent = studentIdInt
	db, dbBool := c.Get("db")
	if dbBool == false {
		outFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(gorm.DB)
	result := database.Where("id_student=?", student.IDStudent).Delete(&Students{})
	if result.Error != nil || result.RowsAffected == 0 {
		outFunc(400, "Problem z usunięciem studenta", result.RowsAffected, result.Error.Error(), c)
		return
	}
	outFunc(200, "Usunięto studenta", result.RowsAffected, "", c)
}

func StudentChange(c *gin.Context) {
	Student := Students{}
	studentIDString, isEmpty := c.Params.Get("studentID")
	if !isEmpty {
		outFunc(400, "Nie podano id studenta", 0, "incorrect student id", c)
		return
	}
	studentIdInt, err := strconv.Atoi(studentIDString)
	if err != nil {
		outFunc(500, "Nieoczekiwany błąd serwera", 0, err.Error(), c)
		return
	}
	err = c.ShouldBindJSON(&Student)
	if err != nil {
		outFunc(400, "Podaj poprawny format danych", 0, err.Error(), c)
		return
	}
	Student.IDStudent = studentIdInt
	db, dbBool := c.Get("db")
	if dbBool == false {
		outFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(gorm.DB)
	result := database.Model(&Student).Update(&Student)
	outFunc(200, "Zmieniono dane studenta", result.RowsAffected, "", c)

}

func StudentAdd(c *gin.Context) {
	var student = Students{}
	err := c.ShouldBindJSON(&student)
	if err != nil {
		outFunc(400, "Podaj poprawny format danych", 0, err.Error(), c)
		return
	}
	db, dbBool := c.Get("db")
	if dbBool == false {
		outFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(gorm.DB)
	result := database.Select("name", "surname", "dob", "id_departament", "sex").Create(&student)
	if result.Error != nil {
		outFunc(400, "Problem z dodaniem studenta", result.RowsAffected, result.Error.Error(), c)
	}
	outFunc(200, "Dodano studenta", result.RowsAffected, "", c)

}
