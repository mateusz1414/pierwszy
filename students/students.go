package students

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Result struct {
	TotalResults int64     `json:"totalResults"`
	Students     []Student `json:"students"`
	ErrorCode    string    `json:"errorCode"`
}

type Outs struct {
	Message    string `json:"message"`
	UpdateRows int64  `json:"updateRows"`
	ErrorCode  string `json:"errorCode"`
}

type Student struct {
	StudentID     int         `json:"studentID"`
	Name          string      `json:"name"`
	Surname       string      `json:"surname"`
	DateOfBrith   string      `json:"dateOfBrith"`
	DepartamentID int         `json:"-"`
	Departament   Departament `gorm:"foreignKey:departament_id;association_foreignkey:departament_id" json:"departament"`
	Sex           string      `json:"sex"`
}

type Departament struct {
	DepartamentID int    `json:"departamentID"`
	Name          string `json:"name"`
}

func OutFunc(status int, mess string, rows int64, errc string, c *gin.Context) {
	outs := Outs{
		Message:    mess,
		UpdateRows: rows,
		ErrorCode:  errc,
	}
	c.JSON(status, outs)
}

func getAll(c *gin.Context, database *gorm.DB) {
	var studenci []Student
	selectResult := database.Joins("inner join Departaments on Departaments.departament_id=Students.departament_id").Preload("Departament").Find(&studenci)
	result := Result{
		TotalResults: selectResult.RowsAffected,
		Students:     studenci,
	}
	//fmt.Println(studenci)
	c.JSON(200, result)
}

func GetStudent(c *gin.Context) {
	db, dbBool := c.Get("db")
	result := Result{}
	status := 200
	studentIDString := c.Param("studentID")
	if !dbBool {
		status = 500
		result.ErrorCode = "Database error"
		c.JSON(status, result)
		return
	}
	database := db.(*gorm.DB)
	if studentIDString == "getAll" {
		getAll(c, database)
		return
	}
	studentIDInt, err := strconv.Atoi(studentIDString)
	if err != nil {
		status = 500
		result.ErrorCode = "Server error"
	}
	var student Student
	selectResult := database.Joins("inner join Departaments on Departaments.departament_id=Students.departament_id").Where("student_id=?", studentIDInt).Preload("Departaments").First(&student)
	fmt.Println(student)
	if selectResult.RowsAffected == 0 {
		status = 404
		result.ErrorCode = "Student not found"
	}
	result.TotalResults = selectResult.RowsAffected
	result.Students = []Student{student}
	c.JSON(status, result)

}

func StudentDelete(c *gin.Context) {
	var student = Student{}
	studentIDString, isEmpty := c.Params.Get("studentID")
	if !isEmpty {
		OutFunc(400, "Nie podano id studenta", 0, "incorrect student id", c)
		return
	}
	studentIdInt, err := strconv.Atoi(studentIDString)
	if err != nil {
		OutFunc(500, "Nieoczekiwany błąd serwera", 0, err.Error(), c)
		return
	}
	student.StudentID = studentIdInt
	db, dbBool := c.Get("db")
	if dbBool == false {
		OutFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(*gorm.DB)
	result := database.Where("student_id=?", student.StudentID).Delete(&Student{})
	if result.Error != nil {
		OutFunc(400, "Problem z usunięciem studenta", result.RowsAffected, result.Error.Error(), c)
		return
	}
	if result.RowsAffected == 0 {
		OutFunc(400, "Problem z usunięciem studenta", result.RowsAffected, "Nie znaleziono studenta", c)
		return
	}
	OutFunc(200, "Usunięto studenta", result.RowsAffected, "", c)
}

func StudentChange(c *gin.Context) {
	newStudent := Student{}
	studentIDString, isEmpty := c.Params.Get("studentID")
	if !isEmpty {
		OutFunc(400, "Nie podano id studenta", 0, "incorrect student id", c)
		return
	}
	studentIdInt, err := strconv.Atoi(studentIDString)
	if err != nil {
		OutFunc(500, "Nieoczekiwany błąd serwera", 0, err.Error(), c)
		return
	}
	err = c.ShouldBindJSON(&newStudent)
	if err != nil {
		OutFunc(400, "Podaj poprawny format danych", 0, err.Error(), c)
		return
	}
	newStudent.StudentID = studentIdInt
	db, dbBool := c.Get("db")
	if dbBool == false {
		OutFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(*gorm.DB)
	result := database.Model(newStudent).Where("student_id=?", newStudent.StudentID).Updates(newStudent)
	if result.Error != nil {
		OutFunc(400, "Problem z usunięciem studenta", result.RowsAffected, result.Error.Error(), c)
		return
	}
	if result.RowsAffected == 0 {
		OutFunc(400, "Problem z usunięciem studenta", result.RowsAffected, "Nie znaleziono studenta", c)
		return
	}
	OutFunc(200, "Zmieniono dane studenta", result.RowsAffected, "", c)

}

func StudentAdd(c *gin.Context) {
	var student = Student{}
	err := c.ShouldBindJSON(&student)
	if err != nil {
		OutFunc(400, "Podaj poprawny format danych", 0, err.Error(), c)
		return
	}
	db, dbBool := c.Get("db")
	if dbBool == false {
		OutFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(*gorm.DB)
	result := database.Select("name", "surname", "date_of_brith", "departament", "sex").Create(&student)
	if result.Error != nil {
		OutFunc(400, "Problem z dodaniem studenta", result.RowsAffected, result.Error.Error(), c)
	}
	OutFunc(200, "Dodano studenta", result.RowsAffected, "", c)

}
