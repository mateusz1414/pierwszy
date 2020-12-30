package teachers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

//Teacher in datatabase
type Teacher struct {
	TeacherID int     `json:"teacherID"`
	Name      string  `json:"name"`
	Surname   string  `json:"surname"`
	SubjectID int     `json:"-"`
	Subject   Subject `gorm:"foreignKey:subject_id" json:"subject"`
}

//Subject in database
type Subject struct {
	SubjectID int    `gorm:"primary_key" json:"subjectID"`
	Name      string `json:"name"`
}

//Result return data in JSON
type Result struct {
	TotalResults int64     `json:"totalResults"`
	Teachers     []Teacher `json:"teachers"`
	ErrorCode    string    `json:"errorCode"`
}

func getAll(c *gin.Context, database *gorm.DB) {
	status := 200
	result := Result{}
	teachers := []Teacher{}
	selectResult := database.Joins("inner join Subjects on Subjects.subject_id=Teachers.subject_id").Preload("Subject").Find(&teachers)
	if selectResult.RowsAffected == 0 {
		status = 400
		result.ErrorCode = "Teachers not found"
	}
	result.TotalResults = selectResult.RowsAffected
	result.Teachers = teachers
	c.JSON(status, result)
}

//GetTeacher show selected teachers
func GetTeacher(c *gin.Context) {
	db, ok := c.Get("db")
	status := 200
	result := Result{}
	teacherIDString := c.Param("teacherID")
	if !ok {
		status = 500
		result.ErrorCode = "Database error"
		c.JSON(status, result)
		return
	}
	database := db.(*gorm.DB)
	if teacherIDString == "getAll" {
		getAll(c, database)
		return
	}
	teacherIDInt, err := strconv.Atoi(teacherIDString)
	if err != nil {
		status = 500
		result.ErrorCode = "Server error"
	}
	teacher := Teacher{
		TeacherID: teacherIDInt,
	}
	selectResult := database.Joins("inner join Subjects on Subjects.subject_id=Teachers.subject_id").Where("teacher_id=?", teacher.TeacherID).Preload("Subjects").First(&teacher)
	if selectResult.RowsAffected == 0 {
		status = 400
		result.ErrorCode = "Teacher not found"
	}
	result.TotalResults = selectResult.RowsAffected
	result.Teachers = []Teacher{teacher}
	c.JSON(status, result)
}
