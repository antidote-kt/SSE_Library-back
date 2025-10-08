package dao

import (
	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
)

func GetCourseByID(courseID uint64) (models.Course, error) {
	db := config.GetDB()
	var course models.Course
	err := db.Where("id = ?", courseID).First(&course).Error
	if err != nil {
		return models.Course{}, err
	}
	return course, nil
}
