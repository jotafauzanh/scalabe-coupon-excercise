package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/model"
)

var (
	ErrCouponNotFound      = errors.New("coupon not found")
	ErrCouponAlreadyExists = errors.New("coupon already exists")
	ErrNoStock             = errors.New("no stock available")
	ErrAlreadyClaimed      = errors.New("coupon already claimed by user")
)

type CouponRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewCouponRepository(db *gorm.DB, redisClient *redis.Client) *CouponRepository {
	return &CouponRepository{
		db:    db,
		redis: redisClient,
	}
}

func (r *CouponRepository) CreateCoupon(ctx context.Context, name string, amount int) (*model.Coupon, error) {
	// Check if coupon already exists
	var existingCoupon model.Coupon
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&existingCoupon).Error
	if err == nil {
		return nil, ErrCouponAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new coupon
	coupon := &model.Coupon{
		Name:            name,
		Amount:          amount,
		RemainingAmount: amount,
	}

	if err := r.db.WithContext(ctx).Create(coupon).Error; err != nil {
		return nil, err
	}

	return coupon, nil
}

func (r *CouponRepository) GetCouponByName(ctx context.Context, name string) (*model.Coupon, error) {
	var coupon model.Coupon
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}

	return &coupon, nil
}

func (r *CouponRepository) ClaimCoupon(ctx context.Context, userID string, couponName string) error {
	// Use Redis distributed lock for this coupon claim operation
	lockKey := fmt.Sprintf("coupon_claim:%s:%s", couponName, userID)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	// Try to acquire lock with SET NX EX
	acquired, err := r.redis.SetNX(ctx, lockKey, lockValue, 30*time.Second).Result()
	if err != nil {
		return err
	}
	if !acquired {
		return errors.New("operation already in progress")
	}

	// Ensure lock is released
	defer func() {
		// Use Lua script to safely delete lock only if we still own it
		script := `
			if redis.call("GET", KEYS[1]) == ARGV[1] then
				return redis.call("DEL", KEYS[1])
			else
				return 0
			end
		`
		r.redis.Eval(ctx, script, []string{lockKey}, lockValue)
	}()

	// Start database transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get coupon with pessimistic lock
		var coupon model.Coupon
		if err := tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").
			Where("name = ?", couponName).First(&coupon).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrCouponNotFound
			}
			return err
		}

		// Check if there's stock available
		if coupon.RemainingAmount <= 0 {
			return ErrNoStock
		}

		// Get user by user_id
		var user model.User
		if err := tx.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("user not found: %s", userID)
			}
			return err
		}

		// Check if user already claimed this coupon
		var existingClaim model.CouponClaims
		err := tx.WithContext(ctx).Where("coupon_id = ? AND user_id = ?", coupon.ID, user.ID).
			First(&existingClaim).Error
		if err == nil {
			return ErrAlreadyClaimed
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Create the claim
		claim := &model.CouponClaims{
			CouponID: coupon.ID,
			UserID:   user.ID,
		}
		if err := tx.WithContext(ctx).Create(claim).Error; err != nil {
			return err
		}

		// Update remaining amount
		if err := tx.WithContext(ctx).Model(&coupon).
			Update("remaining_amount", coupon.RemainingAmount-1).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *CouponRepository) GetCouponDetails(ctx context.Context, name string) (*model.Coupon, []string, error) {
	var coupon model.Coupon
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&coupon).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrCouponNotFound
		}
		return nil, nil, err
	}

	// Get all users who claimed this coupon
	var claims []model.CouponClaims
	err = r.db.WithContext(ctx).Preload("User").Where("coupon_id = ?", coupon.ID).Find(&claims).Error
	if err != nil {
		return nil, nil, err
	}

	claimedBy := make([]string, len(claims))
	for i, claim := range claims {
		claimedBy[i] = claim.User.UserID
	}

	return &coupon, claimedBy, nil
}
