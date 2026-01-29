package model

import "gorm.io/gorm"

type Coupon struct {
	gorm.Model
	Name            string `json:"coupon_name"`
	Amount          int    `json:"amount"`
	RemainingAmount int    `json:"remaining_amount"`
}
