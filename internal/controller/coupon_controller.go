package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/service"
)

type CouponController struct {
	service *service.CouponService
}

func NewCouponController(service *service.CouponService) *CouponController {
	return &CouponController{
		service: service,
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

type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateCoupon - POST /api/coupons
func (c *CouponController) CreateCoupon(ctx *gin.Context) {
	var req CreateCouponRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	coupon, err := c.service.CreateCoupon(ctx.Request.Context(), &service.CreateCouponRequest{
		Name:   req.Name,
		Amount: req.Amount,
	})
	if err != nil {
		if err == service.ErrCouponAlreadyExists {
			ctx.JSON(http.StatusConflict, ErrorResponse{Error: "coupon already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, coupon)
}

// ClaimCoupon - POST /api/coupons/claim
func (c *CouponController) CreateCouponClaim(ctx *gin.Context) {
	var req ClaimCouponRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := c.service.ClaimCoupon(ctx.Request.Context(), &service.ClaimCouponRequest{
		UserID:     req.UserID,
		CouponName: req.CouponName,
	})
	if err != nil {
		if err == service.ErrCouponNotFound {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "coupon not found"})
			return
		}
		if err == service.ErrAlreadyClaimed {
			ctx.JSON(http.StatusConflict, ErrorResponse{Error: "coupon already claimed by user"})
			return
		}
		if err == service.ErrNoStock {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "no stock available"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "coupon claimed successfully"})
}

// GetCoupon - GET /api/coupons?name={name} or /api/coupons/{name}
func (c *CouponController) GetCoupon(ctx *gin.Context) {
	// not sure which one is preferred based on the requirements, so i supported both
	name := ctx.Param("name")
	if name == "" {
		name = ctx.Query("name")
	}
	if name == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "coupon name is required"})
		return
	}

	details, err := c.service.GetCouponDetails(ctx.Request.Context(), name)
	if err != nil {
		if err == service.ErrCouponNotFound {
			ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "coupon not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, details)
}
