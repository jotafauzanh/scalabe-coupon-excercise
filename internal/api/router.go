package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/controller"
	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/repository"
	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/service"
	"github.com/jotafauzanh/scalabe-coupon-excercise/pkg/db"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Initialize dependencies
	userRepo := repository.NewUserRepository(db.DB)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	// Routes
	v1 := r.Group("/api/v1")
	{
		v1.POST("/users", userController.CreateUser)
		v1.GET("/users", userController.GetUsers)
		v1.GET("/users/:id", userController.GetUser)
	}

	return r
}
