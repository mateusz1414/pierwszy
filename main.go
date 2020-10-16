package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

type Studenci struct {
	Imie      string
	Nazwisko  string
	Data      string
	Wydzial   string
	Plec      string
	Id        int
	CzyEdycja bool
}

func connection() (*sql.DB, error) {
	typ := "sqlite3"
	file := "./uczelnia"
	return sql.Open(typ, file)
}

func pobierzStudentow(idDoZmiany int) []Studenci {
	database, err := connection()
	if err != nil {
		fmt.Errorf("Problem z polaczeniem %s", err.Error())
		return nil
	}
	defer database.Close()
	zapytanie, err := database.Query("SELECT imie,nazwisko,data_urodzenia,wydzial,plec,id_studenta from studenci")
	if err != nil {
		fmt.Errorf("Problem z polaczeniem %s", err.Error())
		return nil
	}
	var student = &Studenci{}
	var studenci = []Studenci{}
	for zapytanie.Next() {
		var imie, nazwisko, data, wydzial string
		var plec, id int
		err = zapytanie.Scan(&imie, &nazwisko, &data, &wydzial, &plec, &id)
		if err != nil {
			fmt.Errorf("Problem z polaczeniem %s", err.Error())
			return nil
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
		student.Id = id
		if idDoZmiany == student.Id {
			student.CzyEdycja = true
		} else {
			student.CzyEdycja = false
		}
		studenci = append(studenci, *student)

	}
	return studenci
}

func indexHandler(c *gin.Context) {
	studenci := pobierzStudentow(0)
	c.HTML(200, "index.html", studenci)
}

func studentEdit(c *gin.Context) {
	id, iftrue := c.GetQuery("id")
	idNumber, err := strconv.Atoi(id)
	if err != nil {
		log.Fatal("Problem z zamiana")
		return
	}
	if iftrue {
		studenci := pobierzStudentow(idNumber)
		c.HTML(200, "index.html", studenci)
	}
}

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
	fmt.Println("czy to tu!!!!!!!!!!!!!!!!!!!!!!!!!")
	if err != nil {
		//fmt.Errorf("Problem z zapytaniem %s", err.Error())
		return
	}
	c.Redirect(301, "/")

}

func main() {
	fmt.Println("czemy")
	server := gin.Default()
	server.SetHTMLTemplate(tmpl)
	server.Static("/assets", "./css")
	server.GET("/", indexHandler)
	server.GET("/edit", studentEdit)
	server.POST("/changestudent", studentChange)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.Run(":" + port)
}
