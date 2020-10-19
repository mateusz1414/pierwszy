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
	IDStudenta    int "gorm:primaryKey"
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
	database.AutoMigrate(&Students{})
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

/*
func studentChange(c *gin.Context) {
	id := c.PostForm("id")
	imie := c.PostForm("imie")
	nazwisko := c.PostForm("nazwisko")
	wydzial := c.PostForm("wydzial")
	data := c.PostForm("data")
	plec := c.PostForm("plec")
	database, err := connection()
	if err != nil {
		fmt.Errorf("Problem z polaczeniem %s", err.Error())
		return
	}
	zapytanie, _ := database.Prepare("UPDATE studenci SET imie=?,nazwisko=?,wydzial=?,data_urodzenia=?,plec=? WHERE id_studenta=?")
	_, err = zapytanie.Exec(imie, nazwisko, wydzial, data, plec, id)
	if err != nil {
		fmt.Errorf("Problem z zapytaniem %s", err.Error())
		return
	}
	c.Redirect(301, "/")

}

//co tu wiele mówić  tworzysz studenta wg structury Students patrz linijka 52 nie deklarujesz id samo doda
//funkcja na dodawanie to database.Create(&student)
//i przekierowanie

//edycja robisz tak samo studenta ale pustą zmienną typu Student var student =&Student{} powinno zadzialać jak nie to kombinuj z & i klamrami
//potem pobierasz dane do tej zmiennej database.First(&student,id) First działa troche jak fetch
//zmieniasz co chcesz student.Imie="Darek"
//i zapisujesz db.Save(&student)

func studentAdd(c *gin.Context) {
	imie := c.PostForm("imie")
	nazwisko := c.PostForm("nazwisko")
	wydzial := c.PostForm("wydzial")
	data := c.PostForm("data")
	plec := c.PostForm("plec")
	database, err := connection()
	if err != nil {
		fmt.Errorf("Problem z polaczeniem %s", err.Error())
		return
	}
	zapytanie, _ := database.Prepare("INSERT INTO studenci(imie, nazwisko, data_urodzenia, wydzial, plec) VALUES(?,?,?,?,?)")
	_, err = zapytanie.Exec(imie, nazwisko, data, wydzial, plec)
	if err != nil {
		fmt.Errorf("Problem z zapytaniem %s", err.Error())
		return
	}
	c.Redirect(301, "/")

}*/

func main() {
	server := gin.Default()
	server.SetHTMLTemplate(tmpl)
	server.Static("/assets", "./css")
	server.GET("/", indexHandler)
	server.GET("/delete", studentDelete)
	/*
		server.POST("/changestudent", studentChange)
		server.POST("/addstudent", studentAdd)*/
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.Run(":" + port)
}
