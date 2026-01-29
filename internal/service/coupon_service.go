package service

import (
	"context"
	"errors"

	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/model"
	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/repository"
)

var (
	ErrCouponNotFound      = errors.New("coupon not found")
	ErrCouponAlreadyExists = errors.New("coupon already exists")
	ErrNoStock             = errors.New("no stock available")
	ErrAlreadyClaimed      = errors.New("coupon already claimed by user")
)

type CouponService struct {
	repo *repository.CouponRepository
}

func NewCouponService(repo *repository.CouponRepository) *CouponService {
	return &CouponService{
		repo: repo,
	}
}

type CreateCouponRequest struct {
	Name   string `json:"name" binding:"required"`
	Amount int    `json:"amount" binding:"required,min=1"`
}

type ClaimCouponRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	CouponName string `json:"coupon_name" binding:"required"`
}

type CouponDetailsResponse struct {
	Name            string   `json:"name"`
	Amount          int      `json:"amount"`
	RemainingAmount int      `json:"remaining_amount"`
	ClaimedBy       []string `json:"claimed_by"`
}

func (s *CouponService) CreateCoupon(ctx context.Context, req *CreateCouponRequest) (*model.Coupon, error) {
	// Check if coupon already exists
	_, err := s.repo.GetCouponByName(ctx, req.Name)
	if err == nil {
		return nil, ErrCouponAlreadyExists
	}
	if !errors.Is(err, repository.ErrCouponNotFound) {
		return nil, err
	}

	// Create new coupon
	return s.repo.CreateCoupon(ctx, req.Name, req.Amount)
}

func (s *CouponService) ClaimCoupon(ctx context.Context, req *ClaimCouponRequest) error {
	return s.repo.ClaimCoupon(ctx, req.UserID, req.CouponName)
}

func (s *CouponService) GetCouponDetails(ctx context.Context, name string) (*CouponDetailsResponse, error) {
	coupon, claimedBy, err := s.repo.GetCouponDetails(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrCouponNotFound) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}

	return &CouponDetailsResponse{
		Name:            coupon.Name,
		Amount:          coupon.Amount,
		RemainingAmount: coupon.RemainingAmount,
		ClaimedBy:       claimedBy,
	}, nil
}
