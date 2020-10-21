package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Results struct {
	Status       int
	TotalResults int
	Student      []Students
}

type Outs struct {
	Status     int
	Message    string
	UpdateRows int64
	ErrorValue string
}

type Students struct {
	IDStudenta    int    `json:"idstudenta"`
	Imie          string `json:"imiestudenta"`
	Nazwisko      string `json:"nazwiskostudenta"`
	DataUrodzenia string `json:"datastudenta"`
	Wydzial       string `json:"wydzialstudenta"`
	Plec          string `json:"plecstudenta"`
}

func connection() (*gorm.DB, error) {
	file := "uczelnia"
	return gorm.Open(sqlite.Open(file), &gorm.Config{})
}

func pobierzStudentow() []Students {
	database, err := connection()
	if err != nil {
		panic("failed to connect database")
	}
	//To Twoszr tabele w bazie danych wg struktury
	//database.AutoMigrate(&Students{})
	//defer database.Close()
	var studenci []Students
	database.Find(&studenci)

	return studenci
}

func indexHandler(c *gin.Context) {
	studenci := pobierzStudentow()
	RowsAffected := len(studenci)
	result := Results{
		Status:       200,
		TotalResults: RowsAffected,
		Student:      studenci,
	}
	c.JSON(200, result)
}

func studentDelete(c *gin.Context) {
	var usunSe = Students{}
	err := c.ShouldBindJSON(&usunSe)
	if err != nil {
		outs := Outs{
			Status:     400,
			Message:    "Podaj poprawny format danych",
			UpdateRows: 0,
			ErrorValue: err.Error(),
		}
		c.JSON(400, outs)
		return
	}
	database, err := connection()
	if err != nil {
		fmt.Errorf("Problem z polaczeniem %s", err.Error())
		return
	}
	result := database.Delete(&Students{}, &usunSe)
	if result.Error != nil {
		c.JSON(400, Outs{
			Status:     400,
			Message:    "Problem z usunieciem użytkownika",
			UpdateRows: result.RowsAffected,
			ErrorValue: result.Error.Error(),
		})
	}
	c.JSON(200, Outs{
		Status:     200,
		Message:    "Usunieto użytkownika",
		UpdateRows: result.RowsAffected,
		ErrorValue: "",
	})
}

func studentChange(c *gin.Context) {
	//var edytujSe := Students{}
	//err :=
	//database, err := connection()
	//if err != nil {
	//	fmt.Errorf("Problem z polaczeniem %s", err.Error())
	//	return
	//}

	//database.Where("id_studenta=?", id).Save(&user)
	//c.Redirect(301, "/")

}

func studentAdd(c *gin.Context) {
	var dodajSe = Students{}
	err := c.ShouldBindJSON(&dodajSe)
	if err != nil {
		outs := Outs{
			Status:     400,
			Message:    "Podaj poprawny format danych",
			UpdateRows: 0,
			ErrorValue: err.Error(),
		}
		c.JSON(400, outs)
		return
	}
	database, err := connection()
	if err != nil {
		fmt.Errorf("Problem z polaczeniem %s", err.Error())
		return
	}
	result := database.Select("imie", "nazwisko", "data_urodzenia", "wydzial", "plec").Create(&dodajSe)
	if result.Error != nil {
		//sprawdz 400 !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
		c.JSON(400, Outs{
			Status:     400,
			Message:    "Problem z dodaniem użytkownika",
			UpdateRows: result.RowsAffected,
			ErrorValue: result.Error.Error(),
		})
	}
	c.JSON(200, Outs{
		Status:     200,
		Message:    "Dodano użytkownika",
		UpdateRows: result.RowsAffected,
		ErrorValue: "",
	})

}

func main() {
	server := gin.Default()
	server.GET("/", indexHandler)
	server.DELETE("/", studentDelete)
	server.PUT("/", studentChange)
	server.POST("/", studentAdd)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.Run(":" + port)
}
