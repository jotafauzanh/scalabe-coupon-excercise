package model

import "gorm.io/gorm"

type CouponClaims struct {
	gorm.Model
	CouponID uint   `json:"coupon_id" gorm:"index:idx_coupon_claims_coupon_id,unique"`
	UserID   uint   `json:"user_id" gorm:"index:idx_coupon_claims_user_id,unique"`
	User     User   `gorm:"foreignKey:UserID"`
	Coupon   Coupon `gorm:"foreignKey:CouponID"`
}
