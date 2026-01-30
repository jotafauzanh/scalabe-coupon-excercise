package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	var err error

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	// Setting timezone kept returning error, so i will omit it for now
	// TODO: Figure out db timezone
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, user, password, dbname, port)

	for i := 1; i <= 10; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			log.Println("Database connected")
			break
		}

		log.Printf("Database not ready (attempt %d/10): %v", i, err)
		time.Sleep(time.Duration(i) * time.Second)
	}

	if err != nil {
		log.Fatal("Database never ready:", err)
	}

	log.Println("Nuking tables")
	// Nuke the tables each run, for clean slate, lol
	err = DB.Migrator().DropTable(&model.User{}, &model.Coupon{}, &model.CouponClaims{})
	if err != nil {
		log.Fatal("Failed to nuke tables!", err)
	}

	log.Println("Migrating database")
	// Auto-migrate the schema
	err = DB.AutoMigrate(&model.User{}, &model.Coupon{}, &model.CouponClaims{})
	if err != nil {
		log.Fatal("Failed to migrate database!", err)
	}

	// For some reason, User.user_id kept being created with type int8. So, this is a very bad fix for that
	// Fixed with model edits
	// DB.Exec(`
	// 	ALTER TABLE users
	// 	ALTER COLUMN user_id TYPE TEXT
	// `)

	log.Println("Database connection established")
}
