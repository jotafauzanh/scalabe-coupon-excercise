package model

type CouponClaims struct {
	ID       uint
	CouponID uint   `json:"coupon_id" gorm:"not null;index:idx_coupon_user,unique"`
	Coupon   Coupon `gorm:"belongsTo;foreignKey:CouponID;references:ID"`

	// Use User.UserID, instead of User.ID
	UserID string `json:"user_id" gorm:"type:text;not null;index:idx_coupon_user,unique"`
	User   User   `gorm:"belongsTo;foreignKey:UserID;references:UserID"`
}
