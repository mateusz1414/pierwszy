package grades

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Results struct {
	Status       int
	TotalResults int64
	Grades       []Grades
	ErrorCode    string
}

type Outs struct {
	Message    string
	UpdateRows int64
	ErrorCode  string
}

type Grades struct {
	IDGrade   string `json:"idgrade"`
	Grade     string `json:"grade"`
	IDStudent int    `json:"idstudent"`
	IDSubject int
	Subjects  []Subject `gorm:"foreignKey:id_subject;association_foreignkey:id_subject"`
	Teachers  []Teacher `gorm:"foreignKey:id_subject;association_foreignkey:id_subject"`
}

type Student struct {
	IDStudent int
	Name      string  `json:"name"`
	Surname   string  `json:"surname"`
	Grades    []Grade `gorm:"foreignKey:id_student;association_foreignkey:id_student"`
}

type Subject struct {
	IDSubject int    `gorm:"primary_key"`
	Name      string `json:"name"`
}

type Grade struct {
	IDGrade   int `gorm:"primary_key"`
	Grade     int
	IDStudent int
}

type Teacher struct {
	IDTeacher int
	Name      string
	Surname   string
	IDSubject int
}

func outFunc(status int, mess string, rows int64, errc string, c *gin.Context) {
	outs := Outs{
		Message:    mess,
		UpdateRows: rows,
		ErrorCode:  errc,
	}
	c.JSON(status, outs)
}

func GetStudentsGradesFromSubject(c *gin.Context) {
	var subject = Teacher{}
	teacherIDInt, isEmpty := c.Get("userid")
	if !isEmpty {
		outFunc(400, "Nie znlaziono id danego nauczyciela", 0, "The teacher's id was not found", c)
		return
	}
	subject.IDTeacher = teacherIDInt.(int)
	db, dbBool := c.Get("db")
	if dbBool == false {
		outFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(gorm.DB)
	var grade = []Student{}
	fmt.Println()
	result := database.Joins("INNER JOIN grades on students.id_student=grades.id_student").Joins("INNER JOIN teachers on grades.id_subject=teachers.id_subject").Where("teachers.id_teacher=?", subject.IDTeacher).Group("id_student").Preload("Grades", "grades.id_subject=(SELECT id_subject FROM `teachers` WHERE id_teacher=?)", subject.IDTeacher).Find(&grade)
	if result.Error != nil || result.RowsAffected == 0 {
		outFunc(400, "Problem z wyświetleniem ocen danego przedmiotu", result.RowsAffected, result.Error.Error(), c)
		return
	}
	c.JSON(200, result)
}

func GetGradesForOneStudent(c *gin.Context) {
	var student = Grades{}
	studentIDInt, isEmpty := c.Get("userid")
	if !isEmpty {
		outFunc(400, "Nie podano id studenta", 0, "incorrect student id", c)
		return
	}
	student.IDStudent = studentIDInt.(int)
	db, dbBool := c.Get("db")
	if dbBool == false {
		outFunc(500, "Nie znaleziono bazy danych", 0, "Database error", c)
		return
	}
	database := db.(gorm.DB)
	var grade = []Grades{}
	result := database.Debug().Joins("INNER JOIN teachers on grades.id_subject=teachers.id_subject").Joins("INNER JOIN subjects on grades.id_subject=subjects.id_subject").Where("id_student=?", student.IDStudent).Preload("Subjects").Preload("Teachers").Find(&grade)
	if result.Error != nil || result.RowsAffected == 0 {
		outFunc(400, "Problem z wyświetleniem ocen danego studenta", result.RowsAffected, result.Error.Error(), c)
		return
	}
	c.JSON(200, result)
}

func AddGrade(c *gin.Context) {
	var grade = Teacher{}
	teacherIDInt, isEmpty := c.Get("userid")
	if !isEmpty {
		outFunc(400, "Nie znaleziono id nauczyciela", 0, "not find id teacher", c)
		return
	}
	grade.IDTeacher = teacherIDInt.(int)
	err := c.ShouldBindJSON(&grade)
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
	database.Table("teachers").Where("id_teacher=?", grade.IDTeacher).Find(&grade)
	fmt.Println(grade)
	result := database.Select("grade", "id_student", "id_subject").Create(&grade)
	if result.Error != nil {
		outFunc(400, "Problem z dodaniem oceny", result.RowsAffected, result.Error.Error(), c)
	}
	outFunc(200, "Dodano ocene", result.RowsAffected, "", c)

}
