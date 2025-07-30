package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kol-mikaelson/cyware-go/internal/api"
	"github.com/kol-mikaelson/cyware-go/internal/auth"
	"github.com/kol-mikaelson/cyware-go/internal/database"
)

func main() {
	database.Dbconnect()
	defer database.Dbclose()
	createTable()


	router := gin.Default()
	v1 := router.Group("/api")
	{
		v1.POST("/users/register",api.Register)
		v1.POST("/users/login",api.Login)
		
	}
	authenticated := v1.Group("/")
	authenticated.Use(auth.Middleware())
	{
		authenticated.POST("/questions", api.CreateQuestion)

	}

	router.Run() // listen and serve on 0.0.0.0:8080
}

func createTable(){
	usersTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id UUID PRIMARY KEY,
        username VARCHAR(255) UNIQUE NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMPTZ NOT NULL
    );`

		questionsTableSQL := `
    CREATE TABLE IF NOT EXISTS questions (
        id UUID PRIMARY KEY,
        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        title VARCHAR(255) NOT NULL,
        body TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL
    );`

		answersTableSQL := `
    CREATE TABLE IF NOT EXISTS answers (
        id UUID PRIMARY KEY,
        question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        body TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL
    );`
    queries := []string{usersTableSQL, questionsTableSQL, answersTableSQL}
	for _, query := range queries {
		_, err := database.Db.Exec(context.Background(), query)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}

}
