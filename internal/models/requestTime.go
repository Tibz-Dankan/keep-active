package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (rt *RequestTime) BeforeCreate(tx *gorm.DB) error {
	uuid := uuid.New().String()
	tx.Statement.SetColumn("ID", uuid)
	return nil
}

func (rt *RequestTime) Create(requestTime RequestTime) (RequestTime, error) {
	result := db.Create(&requestTime)

	if result.Error != nil {
		return requestTime, result.Error
	}
	return requestTime, nil
}

func (rt *RequestTime) FindOne(id string) (RequestTime, error) {
	var requestTime RequestTime
	db.First(&requestTime, "id = ?", id)

	return requestTime, nil
}

func (rt *RequestTime) FindByApp(appId string) ([]RequestTime, error) {
	var requestTimes []RequestTime

	db.Find(&requestTimes, "\"appId\" = ?", appId)

	return requestTimes, nil
}

func (rt *RequestTime) Update() error {
	db.Save(&rt)

	return nil
}

func (r *RequestTime) Delete(id string) error {

	if err := db.Unscoped().Where("id = ?", id).Delete(&RequestTime{}).Error; err != nil {
		return err
	}

	return nil
}
