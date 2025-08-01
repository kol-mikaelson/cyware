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
		v1.POST("/users/register", api.Register)
		v1.POST("/users/login", api.Login)
		v1.GET("/questions/:questionid/summarize", api.SummarizeQuestion)
		v1.GET("/questions/:questionid", api.GetQuestion)

	}
	authenticated := v1.Group("/")
	authenticated.Use(auth.Middleware())
	{
		authenticated.POST("/questions", api.CreateQuestion)
		authenticated.POST("/questions/:questionid/answer", api.CreateAnswer)

		authenticated.POST("/questions/:questionid/vote", api.VoteQuestion)
		authenticated.POST("/answers/:answerid/vote", api.VoteAnswer)
	}

	router.Run() // listen and serve on 0.0.0.0:8080
}

func createTable() {
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

	votesTableSQL := `
    CREATE TABLE IF NOT EXISTS votes (
        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        post_id UUID NOT NULL,
        post_type VARCHAR(10) NOT NULL, -- 'question' or 'answer'
        vote_type SMALLINT NOT NULL,    -- 1 for upvote, -1 for downvote
        PRIMARY KEY (user_id, post_id) -- Composite key ensures one vote per user per post
    );`
	queries := []string{usersTableSQL, questionsTableSQL, answersTableSQL, votesTableSQL}
	for _, query := range queries {
		_, err := database.Db.Exec(context.Background(), query)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}

}
