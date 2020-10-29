package studenci

import (
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
	var usunSe = Students{}
	err := c.ShouldBindJSON(&usunSe)
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
	result := database.Delete(&Students{}, usunSe.IDStudenta)
	if result.Error != nil || result.RowsAffected == 0 {
		outFunc(400, "Problem z usunięciem studenta", result.RowsAffected, result.Error.Error(), c)
		return
	}
	outFunc(200, "Usunięto studenta", result.RowsAffected, "", c)
}

func StudentChange(c *gin.Context) {
	edytujSe := Students{}
	err := c.ShouldBindJSON(&edytujSe)
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
	student := Students{}
	if edytujSe.IDStudenta == 0 {
		outFunc(400, "Niepoprawne idstudenta", 0, "Invalid primarykey", c)
		return
	}
	database.First(&student, edytujSe.IDStudenta)
	student.compare(&edytujSe)

	result := database.Where("id_studenta=?", student.IDStudenta).Save(&student)
	outFunc(200, "Zmieniono dane studenta", result.RowsAffected, "", c)

}

func StudentAdd(c *gin.Context) {
	var dodajSe = Students{}
	err := c.ShouldBindJSON(&dodajSe)
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
	result := database.Select("imie", "nazwisko", "data_urodzenia", "wydzial", "plec").Create(&dodajSe)
	if result.Error != nil {
		//sprawdz 400 !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		outFunc(400, "Problem z dodaniem studenta", result.RowsAffected, result.Error.Error(), c)
	}
	outFunc(200, "Dodano studenta", result.RowsAffected, "", c)

}
