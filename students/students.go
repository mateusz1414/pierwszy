package students

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Result struct {
	TotalResults int64
	Students     []Student
	ErrorCode    string
}

type Outs struct {
	Message    string
	UpdateRows int64
	ErrorCode  string
}

type Student struct {
	StudentID   int    `json:"studentID"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	DateOfBrith string `json:"dateofbrith"`
	Departament string `json:"departament"`
	Sex         string `json:"sex"`
}

func outFunc(status int, mess string, rows int64, errc string, c *gin.Context) {
	outs := Outs{
		Message:    mess,
		UpdateRows: rows,
		ErrorCode:  errc,
	}
	c.JSON(status, outs)
}

func (s *Student) compare(nowy *Student) {
	if nowy.Name != "" && nowy.Name != s.Name {
		s.Name = nowy.Name
	}
	if nowy.Surname != "" && nowy.Surname != s.Surname {
		s.Surname = nowy.Surname
	}
	if nowy.DateOfBrith != "" && nowy.DateOfBrith != s.DateOfBrith {
		s.DateOfBrith = nowy.DateOfBrith
	}
	if nowy.Departament != "" && nowy.Departament != s.Departament {
		s.Departament = nowy.Departament
	}
	if nowy.Sex != "" && nowy.Sex != s.Sex {
		s.Sex = nowy.Sex
	}

}

func getAll(c *gin.Context, database *gorm.DB) {
	var studenci []Student
	selectResult := database.Find(&studenci)
	result := Result{
		TotalResults: selectResult.RowsAffected,
		Students:     studenci,
	}
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
	selectResult := database.Where("student_id=?", studentIDInt).Find(&student)
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
		outFunc(400, "Nie podano id studenta", 0, "incorrect student id", c)
		return
	}
	studentIdInt, err := strconv.Atoi(studentIDString)
	if err != nil {
		outFunc(500, "Nieoczekiwany błąd serwera", 0, err.Error(), c)
		return
	}
	student.StudentID = studentIdInt
	db, dbBool := c.Get("db")
	if dbBool == false {
		outFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(gorm.DB)
	result := database.Where("student_id=?", student.StudentID).Delete(&Student{})
	if result.Error != nil || result.RowsAffected == 0 {
		outFunc(400, "Problem z usunięciem studenta", result.RowsAffected, result.Error.Error(), c)
		return
	}
	outFunc(200, "Usunięto studenta", result.RowsAffected, "", c)
}

func StudentChange(c *gin.Context) {
	newStudent := Student{}
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
	err = c.ShouldBindJSON(&newStudent)
	if err != nil {
		outFunc(400, "Podaj poprawny format danych", 0, err.Error(), c)
		return
	}
	newStudent.StudentID = studentIdInt
	db, dbBool := c.Get("db")
	if dbBool == false {
		outFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(gorm.DB)

	result := database.Model(newStudent).Where("student_id=?", newStudent.StudentID).Updates(newStudent)
	outFunc(200, "Zmieniono dane studenta", result.RowsAffected, "", c)

}

func StudentAdd(c *gin.Context) {
	var student = Student{}
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
	result := database.Select("imie", "nazwisko", "data_urodzenia", "wydzial", "plec").Create(&student)
	if result.Error != nil {
		outFunc(400, "Problem z dodaniem studenta", result.RowsAffected, result.Error.Error(), c)
	}
	outFunc(200, "Dodano studenta", result.RowsAffected, "", c)

}
