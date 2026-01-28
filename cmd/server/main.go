package main

import (
	"log"

	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/api"
	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/model"
	"github.com/jotafauzanh/scalabe-coupon-excercise/pkg/db"
)

func main() {
	// Connect to database
	db.ConnectDatabase()

	// Migrate the schema
	db.DB.AutoMigrate(&model.User{})

	// Setup router
	r := api.SetupRouter()

	// Run server
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
