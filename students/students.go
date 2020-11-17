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
	IDStudenta    int    `json:"idstudenta"`
	Imie          string `json:"imiestudenta"`
	Nazwisko      string `json:"nazwiskostudenta"`
	DataUrodzenia string `json:"datastudenta"`
	Wydzial       string `json:"wydzialstudenta"`
	Plec          string `json:"plecstudenta"`
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

func (s *Students) compare(nowy *Students) {
	if nowy.Imie != "" && nowy.Imie != s.Imie {
		s.Imie = nowy.Imie
	}
	if nowy.Nazwisko != "" && nowy.Nazwisko != s.Nazwisko {
		s.Nazwisko = nowy.Nazwisko
	}
	if nowy.DataUrodzenia != "" && nowy.DataUrodzenia != s.DataUrodzenia {
		s.DataUrodzenia = nowy.DataUrodzenia
	}
	if nowy.Wydzial != "" && nowy.Wydzial != s.Wydzial {
		s.Wydzial = nowy.Wydzial
	}
	if nowy.Plec != "" && nowy.Plec != s.Plec {
		s.Plec = nowy.Plec
	}

}

func IndexHandler(c *gin.Context) {
	db, dbBool := c.Get("db")
	result := Results{}
	if dbBool != true {
		result = Results{
			Status:    500,
			ErrorCode: "Database error",
		}
	} else {
		var studenci []Students
		database := db.(gorm.DB)
		selectResult := database.Find(&studenci)
		result = Results{
			Status:       200,
			TotalResults: selectResult.RowsAffected,
			Student:      studenci,
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
	student.IDStudenta = studentIdInt
	db, dbBool := c.Get("db")
	if dbBool == false {
		outFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(gorm.DB)
	result := database.Where("id_studenta=?", student.IDStudenta).Delete(&Students{})
	if result.Error != nil || result.RowsAffected == 0 {
		outFunc(400, "Problem z usunięciem studenta", result.RowsAffected, result.Error.Error(), c)
		return
	}
	outFunc(200, "Usunięto studenta", result.RowsAffected, "", c)
}

func StudentChange(c *gin.Context) {
	newStudent := Students{}
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
	newStudent.IDStudenta = studentIdInt
	db, dbBool := c.Get("db")
	if dbBool == false {
		outFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(gorm.DB)
	oldStudent := Students{}
	database.First(&oldStudent, newStudent.IDStudenta)
	oldStudent.compare(&newStudent)

	result := database.Model(oldStudent).Where("id_studenta=?", oldStudent.IDStudenta).Save(&oldStudent)
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
	result := database.Select("imie", "nazwisko", "data_urodzenia", "wydzial", "plec").Create(&student)
	if result.Error != nil {
		outFunc(400, "Problem z dodaniem studenta", result.RowsAffected, result.Error.Error(), c)
	}
	outFunc(200, "Dodano studenta", result.RowsAffected, "", c)

}
