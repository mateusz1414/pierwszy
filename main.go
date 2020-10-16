package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

type Studenci struct {
	Imie     string
	Nazwisko string
	Data     string
	Wydzial  string
	Plec     string
}

func connection() (*sql.DB, error) {
	typ := "sqlite3"
	file := "./uczelnia"
	return sql.Open(typ, file)
}

func indexHandler(c *gin.Context) {
	database, err := connection()
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer database.Close()
	zapytanie, err := database.Query("SELECT imie,nazwisko,data_urodzenia,wydzial,plec from studenci")
	if err != nil {
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	var student = Studenci{}
	var studenci = []Studenci{}
	for zapytanie.Next() {
		var imie, nazwisko, data, wydzial string
		var plec int
		err = zapytanie.Scan(&imie, &nazwisko, &data, &wydzial, &plec)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		var p string
		if plec == 0 {
			p = "mężczyzna"
		} else {
			p = "kobieta"
		}

		student.Imie = imie
		student.Nazwisko = nazwisko
		student.Wydzial = wydzial
		student.Data = data
		student.Plec = p
		studenci = append(studenci, student)
	}
	c.HTML(200, "index.html", studenci)
}

func main() {
	fmt.Println("czemy")
	server := gin.Default()
	server.SetHTMLTemplate(tmpl)
	server.Static("/assets", "/css")
	server.GET("/", indexHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.Run(":" + port)
}
