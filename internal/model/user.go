package model

type User struct {
	ID     uint `gorm:"primaryKey"`
	Name   string
	UserID string `json:"user_id" gorm:"type:text;uniqueIndex"`
}
