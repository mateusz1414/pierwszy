package students

import (
	"fmt"
	"strconv"
	"students/user"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

//Result list students
type Result struct {
	TotalResults int64     `json:"totalResults"`
	Students     []Student `json:"students"`
	ErrorCode    string    `json:"errorCode"`
}

//Outs return information
type Outs struct {
	Message    string `json:"message"`
	UpdateRows int64  `json:"updateRows"`
	ErrorCode  string `json:"errorCode"`
}

//Student is student with data
type Student struct {
	StudentID     int         `json:"studentID" gorm:"primary_key"`
	Name          string      `json:"name"`
	Surname       string      `json:"surname"`
	Dob           string      `json:"dob"`
	DepartamentID int         `json:"-"`
	Departaments  Departament `gorm:"foreignKey:departament_id;association_foreignkey:departament_id" json:"departament"`
	Sex           *int        `json:"sex"`
}

//Waiting is user who wait on add to student list
type Waiting struct {
	StudentID     int64  `json:"studentID" gorm:"primary_key"`
	Name          string `json:"name"`
	Surname       string `json:"surname"`
	Dob           string `json:"dob"`
	DepartamentID int    `json:"departamentID"`
	Sex           int    `json:"sex"`
}

//Departament in database
type Departament struct {
	DepartamentID int    `json:"departamentID"`
	Name          string `json:"name"`
}

//OutFunc return information
func OutFunc(status int, mess string, rows int64, errc string, c *gin.Context) {
	outs := Outs{
		Message:    mess,
		UpdateRows: rows,
		ErrorCode:  errc,
	}
	c.JSON(status, outs)
}

func getAll(c *gin.Context, database *gorm.DB) {
	var studenci []Student
	selectResult := database.Joins("inner join departaments on departaments.departament_id=students.departament_id").Order("surname,name").Preload("Departaments").Find(&studenci)
	result := Result{
		TotalResults: selectResult.RowsAffected,
		Students:     studenci,
	}
	c.JSON(200, result)
}

//GetStudent show data of student
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
	var student Student
	var err error
	student.StudentID, err = strconv.Atoi(studentIDString)
	if err != nil {
		status = 500
		result.ErrorCode = "Server error"
	}
	selectResult := database.Joins("inner join departaments on departaments.departament_id=students.departament_id").Where("student_id=?", student.StudentID).Order("surname,name").Preload("Departaments").First(&student)
	if selectResult.RowsAffected == 0 {
		status = 404
		result.ErrorCode = "Student not found"
	}
	result.TotalResults = selectResult.RowsAffected
	result.Students = []Student{student}
	c.JSON(status, result)
}

//StudentDelete delete student
func StudentDelete(c *gin.Context) {
	var student = Student{}
	var err error
	studentIDString := c.Param("studentID")
	student.StudentID, err = strconv.Atoi(studentIDString)
	if err != nil {
		OutFunc(500, err.Error(), 0, "Server error", c)
		return
	}
	db, dbBool := c.Get("db")
	if dbBool == false {
		OutFunc(500, "", 0, "Database error", c)
		return
	}
	database := db.(*gorm.DB)
	err = database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("Grades").Where("student_id=?", student.StudentID).Delete(&Student{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&user.User{}).Where("user_id=?", student.StudentID).Update("permissions", "user").Error; err != nil {
			return err
		}
		if result := tx.Delete(&student, student.StudentID); result.RowsAffected == 0 {
			return fmt.Errorf("Not found")
		}
		return nil
	})
	if err != nil {
		OutFunc(400, "", 0, "Not found", c)
		return
	}
	OutFunc(200, "Success", 1, "", c)
}

//StudentChange change student
func StudentChange(c *gin.Context) {
	newStudent := Student{}
	studentIDString := c.Param("studentID")
	studentIDInt, err := strconv.Atoi(studentIDString)
	if err != nil {
		OutFunc(500, err.Error(), 0, "Server error", c)
		return
	}
	if err = c.ShouldBindJSON(&newStudent); err != nil {
		OutFunc(400, err.Error(), 0, "Invalid data", c)
		return
	}
	newStudent.StudentID = studentIDInt
	db, dbBool := c.Get("db")
	if !dbBool {
		OutFunc(500, "", 0, "Database error", c)
		return
	}
	database := db.(*gorm.DB)
	if result := database.Model(newStudent).Where("student_id=?", newStudent.StudentID).Updates(newStudent); result.RowsAffected == 0 {
		OutFunc(400, "", result.RowsAffected, "Not found", c)
		return
	}
	OutFunc(200, "Zmieniono dane studenta", 1, "", c)

}

//StudentAdd user to student list
func StudentAdd(c *gin.Context) {
	var (
		student   = Waiting{}
		db        interface{}
		studentID int
		err       error
		ok        bool
	)
	if studentID, err = strconv.Atoi(c.Param("studentID")); err != nil {
		OutFunc(400, "", 0, "Invalid data", c)
		return
	}
	student.StudentID = int64(studentID)
	if db, ok = c.Get("db"); !ok {
		OutFunc(500, "", 0, "Database error", c)
		return
	}
	database := db.(*gorm.DB)
	err = database.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&student, student.StudentID).Error; err != nil {
			return err
		}
		if err := tx.Table("Students").Create(&student).Error; err != nil {
			return err
		}
		if err := tx.Delete(&student, student.StudentID).Error; err != nil {
			return err
		}
		if err := tx.Model(user.User{}).Where("user_id=?", student.StudentID).Updates(&user.User{Permissions: "student"}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		OutFunc(400, err.Error(), 0, "Add failed", c)
		return
	}
	OutFunc(200, "Success", 1, "", c)
}

//WaitingDiscard dalete user with waiting list
func WaitingDiscard(c *gin.Context) {
	var student = Waiting{}
	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		OutFunc(400, "", 0, "Invalid data", c)
		return
	}
	student.StudentID = int64(studentID)
	db, ok := c.Get("db")
	if !ok {
		OutFunc(500, "", 0, "Database error", c)
		return
	}
	database := db.(*gorm.DB)
	deleteResult := database.Delete(student, student.StudentID)
	if deleteResult.RowsAffected == 0 {
		OutFunc(400, "User not found", 0, "Delete failed", c)
		return
	}
	OutFunc(200, "Succes", deleteResult.RowsAffected, "", c)
}

//RequestStudent add student on waiting list
func RequestStudent(c *gin.Context) {
	var student = Waiting{}
	if err := c.ShouldBindJSON(&student); err != nil {
		OutFunc(400, err.Error(), 0, "Invalid data format", c)
		return
	}
	db, dbBool := c.Get("db")
	if dbBool == false {
		OutFunc(500, "Database not found", 0, "Database error", c)
		return
	}
	database := db.(*gorm.DB)
	claims, ok := c.Get("userid")
	if !ok {
		OutFunc(500, "", 0, "Server error", c)
		return
	}
	student.StudentID = claims.(int64)
	var count int64
	database.Table("waitings").Where("student_id=?", student.StudentID).Count(&count)
	if count != 0 {
		OutFunc(400, "", 0, "On list", c)
		return
	}
	result := database.Select("student_id", "name", "surname", "dob", "departament_id", "sex").Create(&student)
	if result.Error != nil {
		OutFunc(500, result.Error.Error(), result.RowsAffected, "Send request problem", c)
		return
	}
	OutFunc(200, "Succes", result.RowsAffected, "", c)

}

//ApplicationList show list users witings on add to students
func ApplicationList(c *gin.Context) {
	result := Result{}
	db, ok := c.Get("db")
	status := 200
	if !ok {
		result.ErrorCode = "Database error"
		status = 500
		c.JSON(status, result)
		return
	}
	database := db.(*gorm.DB)
	selectResult := database.Table("waitings").Joins("left join departaments on departaments.departament_id=waitings.departament_id").Order("surname,name").Preload("Departaments").Find(&result.Students)
	result.TotalResults = selectResult.RowsAffected
	c.JSON(status, result)
}
