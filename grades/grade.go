package grades

import (
	"students/students"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type StudentResult struct {
	GradesCount int64     `json:"gradeCount"`
	StudentID   int64     `json:"studentID"`
	Subjects    []Subject `json:"subjects"`
	ErrorCode   string    `json:"errorCode"`
}

type TeacherResult struct {
	GradesCount int64     `json:"gradeCount"`
	Name        string    `json:"subjectName"`
	Students    []Student `json:"students"`
	ErrorCode   string    `json:"errorCode"`
}

type Teacher struct {
	TeacherID int
	SubjectID int
}

type Student struct {
	StudentID int     `json:"studentID"`
	Name      string  `json:"name"`
	Surname   string  `json:"surname"`
	Grades    []Grade `gorm:"foreignKey:student_id;association_foreignkey:student_id" json:"grades"`
}

type Subject struct {
	Name      string  `json:"name"`
	SubjectID int     `json:"subjectID"`
	Grades    []Grade `gorm:"foreignKey:subject_id;association_foreignkey:subject_id" json:"grades"`
}

type Grade struct {
	Value     int `json:"value"`
	StudentID int `json:"-"`
	SubjectID int `json:"-"`
}

type GradeAdd struct {
	Value     int `json:"value"`
	StudentID int `json:"studentID"`
	SubjectID int `json:"subjectID"`
}

func GetStudentGrades(c *gin.Context) {
	db, ok := c.Get("db")
	status := 200
	result := StudentResult{}
	if !ok {
		status = 500
		result.ErrorCode = "Database error"
		c.JSON(status, result)
		return
	}
	claims, ok := c.Get("userid")
	if !ok {
		status = 500
		result.ErrorCode = "Server error"
		c.JSON(status, result)
		return
	}
	result.StudentID = claims.(int64)
	database := db.(*gorm.DB)
	selectResult := database.Joins("inner join Grades on Grades.subject_id=Subjects.subject_id").Where("student_id=?", result.StudentID).Group("Subjects.subject_id").Preload("Grades", "student_id=?", result.StudentID).Find(&result.Subjects)
	result.GradesCount = selectResult.RowsAffected
	c.JSON(status, result)

}

func GetAllGrades(c *gin.Context) {
	db, ok := c.Get("db")
	status := 200
	result := TeacherResult{}
	if !ok {
		status = 500
		result.ErrorCode = "Database error"
		c.JSON(status, result)
		return
	}
	claims, ok := c.Get("userid")
	if !ok {
		status = 500
		result.ErrorCode = "Server error"
		c.JSON(status, result)
		return
	}
	userID := claims.(int64)
	database := db.(*gorm.DB)
	selectResult := database.Joins("inner join Departament_subject ds on Students.departament_id=ds.departament_id").Joins("left join Grades on Students.student_id=Grades.student_id AND ds.subject_id=Grades.subject_id").Joins("left join Subjects on Grades.subject_id=Subjects.subject_id").Joins("inner join Teachers on ds.subject_id=Teachers.subject_id").Where("Teachers.teacher_id=?", userID).Group("Students.student_id").Preload("Grades", "Grades.subject_id=(SELECT subject_id FROM Teachers WHERE teacher_id=?)", userID).Find(&result.Students)
	result.GradesCount = selectResult.RowsAffected
	database.Table("Subjects").Select("name").Where("subject_id=(SELECT subject_id FROM Teachers WHERE teacher_id=?)", userID).First(&result)
	c.JSON(status, result)

}

func AddGrade(c *gin.Context) {
	db, ok := c.Get("db")
	if !ok {
		students.OutFunc(500, "", 0, "database error", c)
		return
	}
	claims, ok := c.Get("userid")
	if !ok {
		students.OutFunc(500, "", 0, "server error", c)
		return
	}
	grade := GradeAdd{}
	err := c.ShouldBindJSON(&grade)
	if err != nil {
		students.OutFunc(500, "", 0, "server error", c)
		return
	}

	userID := claims.(int64)
	database := db.(*gorm.DB)
	var teacher Teacher
	count := 0
	database.Where("teacher_id=?", userID).First(&teacher)
	grade.SubjectID = teacher.SubjectID
	database.Table("Students").Joins("inner join Departament_subject on Students.departament_id=Departament_subject.departament_id").Where("subject_id=? AND student_id=?", grade.SubjectID, grade.StudentID).Count(&count)
	if count == 0 {
		students.OutFunc(400, "Student not found", 0, "not found", c)
		return
	}
	selectResult := database.Table("Grades").Create(&grade)
	if selectResult.RowsAffected == 0 {
		students.OutFunc(400, "", 0, "database error", c)
		return
	}
	grades := []Grade{}
	database.Where("student_id=? AND subject_id=?", grade.StudentID, grade.SubjectID).Find(&grades)
	c.JSON(200, gin.H{
		"message":       "success",
		"updateRows":    selectResult.RowsAffected,
		"studentID":     grade.StudentID,
		"errorCode":     "",
		"studentGrades": grades,
	})
}