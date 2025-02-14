package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB() {
	// Railway injects environment variables, no need to load .env
	dsn := os.Getenv("DATABASE_URL") // Railway uses DATABASE_URL format

	var err error
	DB, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	fmt.Println("Connected to the database!")
}

func CloseDB() {
    if DB != nil {
        DB.Close()
        fmt.Println("Database connection closed.")
    }
}
