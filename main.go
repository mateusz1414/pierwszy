package main

import (
	"fmt"
	"html/template"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

type Results struct {
	Status       int
	TotalResults int
	Student      []Students
}

type Students struct {
	IDStudenta    int "gorm:primaryKey;autoIncrement"
	Imie          string
	Nazwisko      string
	DataUrodzenia string
	Wydzial       string
	Plec          string
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
	id, _ := c.GetQuery("id")
	database, err := connection()
	if err != nil {
		fmt.Errorf("Problem z polaczeniem %s", err.Error())
		return
	}
	database.Delete(&Students{}, id)
	c.Redirect(301, "/")
}

func studentChange(c *gin.Context) {
	database, err := connection()
	if err != nil {
		fmt.Errorf("Problem z polaczeniem %s", err.Error())
		return
	}
	id := c.GetHeader("id")
	imie := c.GetHeader("imie")
	nazwisko := c.GetHeader("nazwisko")
	wydzial := c.GetHeader("wydzial")
	data := c.GetHeader("data")
	plec := c.GetHeader("plec")
	if id == "" {
		c.Redirect(301, "/")
	}
	var user Students
	database.First(&user, id)
	if imie != "" {
		user.Imie = imie
	}
	if nazwisko != "" {
		user.Nazwisko = nazwisko
	}
	if wydzial != "" {
		user.Wydzial = wydzial
	}
	if data != "" {
		user.DataUrodzenia = data
	}
	if plec != "" {
		user.Plec = plec
	}
	database.Where("id_studenta=?", id).Save(&user)
	c.Redirect(301, "/")

}

func studentAdd(c *gin.Context) {
	imie := c.GetHeader("imie")
	nazwisko := c.GetHeader("nazwisko")
	wydzial := c.GetHeader("wydzial")
	data := c.GetHeader("data")
	plec := c.GetHeader("plec")
	user := Students{
		Imie:          imie,
		Nazwisko:      nazwisko,
		Wydzial:       wydzial,
		DataUrodzenia: data,
		Plec:          plec,
	}
	database, err := connection()
	if err != nil {
		fmt.Errorf("Problem z polaczeniem %s", err.Error())
		return
	}

	database.Select("imie", "nazwisko", "data_urodzenia", "wydzial", "plec").Create(&user)
	fmt.Println(user.IDStudenta)
	c.Redirect(301, "/")

}

func main() {
	server := gin.Default()
	server.SetHTMLTemplate(tmpl)
	server.Static("/assets", "./css")
	server.GET("/", indexHandler)
	server.GET("/delete", studentDelete)
	server.POST("/changestudent", studentChange)
	server.POST("/addstudent", studentAdd)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.Run(":" + port)
}
