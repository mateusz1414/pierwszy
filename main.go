package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var tmpl = template.Must(template.ParseFiles("index.html"))

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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	database, err := connection()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer database.Close()
	zapytanie, err := database.Query("SELECT imie,nazwisko,data_urodzenia,wydzial,plec from studenci")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var student = Studenci{}
	var studenci = []Studenci{}
	for zapytanie.Next() {
		var imie, nazwisko, data, wydzial string
		var plec int
		err = zapytanie.Scan(&imie, &nazwisko, &data, &wydzial, &plec)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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
	tmpl.Execute(w, studenci)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	fs := http.FileServer(http.Dir("css"))
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.Handle("/css/", http.StripPrefix("/css/", fs))
	http.ListenAndServe(":"+port, mux)
}
