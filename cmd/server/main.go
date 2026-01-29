package main

import (
	"log"

	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/api"
	"github.com/jotafauzanh/scalabe-coupon-excercise/pkg/db"
	"github.com/jotafauzanh/scalabe-coupon-excercise/pkg/redis"
)

func main() {
	log.Println("Server is starting...")
	// Connect to database
	db.ConnectDatabase()

	// Connect to Redis
	redis.ConnectRedis()

	// Setup router
	r := api.SetupRouter()

	// Run server
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
