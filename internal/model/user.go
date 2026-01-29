package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name   string `json:"name"`
	UserID string `json:"user_id" gorm:"unique"`
}
