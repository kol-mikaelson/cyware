package main

import (
	"context"
	"fmt"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/kol-mikaelson/cyware-go/internal/api"
	"github.com/kol-mikaelson/cyware-go/internal/database"
)

func main() {
	database.Dbconnect()
	defer database.Dbclose()
	createTableSQL := `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);`
	_, err := database.Db.Exec(context.Background(), createTableSQL)
	if err != nil {
	    log.Fatalf("Failed to create users table: %v", err)
	}
	fmt.Println("Users table is ready.")


	router := gin.Default()
	router.POST("/register",api.Register)
	router.Run() // listen and serve on 0.0.0.0:8080
}

