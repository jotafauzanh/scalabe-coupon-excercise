package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/controller"
	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/repository"
	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/service"
	"github.com/jotafauzanh/scalabe-coupon-excercise/pkg/db"
	"github.com/jotafauzanh/scalabe-coupon-excercise/pkg/redis"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Initialize dependencies
	userRepo := repository.NewUserRepository(db.DB)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)
	couponRepo := repository.NewCouponRepository(db.DB, redis.Client)
	couponService := service.NewCouponService(couponRepo)
	couponController := controller.NewCouponController(couponService)
	devController := controller.NewDevController()

	// Routes
	v1 := r.Group("/api")
	{
		// Users
		v1.POST("/users", userController.CreateUser)
		v1.GET("/users", userController.GetUsers)
		v1.GET("/users/:id", userController.GetUser)

		// Coupons
		v1.POST("/coupons", couponController.CreateCoupon)
		v1.POST("/coupons/claim", couponController.CreateCouponClaim)
		v1.GET("/coupons/:id", couponController.GetCoupon)

		// DEV
		v1.GET("/health", devController.HealthCheck)
	}

	return r
}
