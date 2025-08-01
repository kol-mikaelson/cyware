package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kol-mikaelson/cyware-go/internal/auth"
	"github.com/kol-mikaelson/cyware-go/internal/database"
)

type user struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type DBUser struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
type LoginData struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type QuestionData struct {
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body" binding:"required"`
	Category string `json:"category"`
}

type QuestionResponse struct {
	UUID      string    `json:"id"`
	PostedBy  string    `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	PostedAt  time.Time `json:"created_at"`
	VoteScore int       `json:"vote_score"`
}

type AnswerData struct {
	Body string `json:"body" binding:"required"`
}

type AnswerResponse struct {
	UUID       string    `json:"uuid"`
	QuestionID string    `json:"question_id"`
	Body       string    `json:"body"`
	PostedAt   time.Time `json:"posted_at"`
	PostedBy   string    `json:"posted_by"`
	VoteScore  int
}

type VoteData struct {
	VoteType int `json:"vote_type" binding:"required,oneof=-1 1"`
}
type QuestionWithAnswers struct {
	Question QuestionResponse `json:"question"`
	Answers  []AnswerResponse `json:"answers"`
}

func Register(c *gin.Context) {
	var data user
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var Hashed, err = auth.HashPassword(data.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newUser := DBUser{
		ID:           uuid.New().String(),
		Username:     data.Username,
		Email:        data.Email,
		PasswordHash: Hashed,
		CreatedAt:    time.Now(),
	}
	sql := `INSERT INTO users (id, username, email, password_hash, created_at)
         VALUES ($1, $2, $3, $4, $5)`
	_, err = database.Db.Exec(context.Background(), sql, newUser.ID, newUser.Username, newUser.Email, newUser.PasswordHash, newUser.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":         newUser.ID,
		"username":   newUser.Username,
		"email":      newUser.Email,
		"created_at": newUser.CreatedAt,
	})

}

func Login(c *gin.Context) {
	var req LoginData
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := req.Email
	esql := "SELECT id,password_hash FROM users WHERE email = $1"
	var userID string
	var Hashed string
	err := database.Db.QueryRow(context.Background(), esql, email).Scan(&userID, &Hashed)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	password := req.Password
	err = auth.VerifyPassword(context.Background(), password, Hashed)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	token, err := auth.GenerateJwt(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})

}

func CreateQuestion(c *gin.Context) {
	userID := c.GetString("userID")
	var qd QuestionData
	if err := c.BindJSON(&qd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	quest := QuestionResponse{
		Title:    qd.Title,
		Body:     qd.Body,
		PostedBy: userID,
		PostedAt: time.Now(),
		UUID:     uuid.NewString(),
	}
	sqlcomm := "INSERT INTO questions (id, user_id, title, body, created_at) VALUES ($1, $2, $3, $4, $5)"
	_, err := database.Db.Exec(context.Background(), sqlcomm, quest.UUID, quest.PostedBy, quest.Title, quest.Body, quest.PostedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question"})
		return
	}
	c.JSON(http.StatusCreated, quest)

}

func CreateAnswer(c *gin.Context) {
	userID := c.GetString("userID")
	Qid := c.Param("questionid")
	var ad AnswerData
	if err := c.BindJSON(&ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	answer := AnswerResponse{
		QuestionID: Qid,
		Body:       ad.Body,
		PostedBy:   userID,
		PostedAt:   time.Now(),
		UUID:       uuid.NewString(),
	}
	sqlcomm := "INSERT INTO answers (id, user_id, body, created_at, question_id) VALUES ($1, $2, $3, $4, $5)"
	_, err := database.Db.Exec(context.Background(), sqlcomm, answer.UUID, answer.PostedBy, answer.Body, answer.PostedAt, answer.QuestionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create answer"})
		return
	}
	c.JSON(http.StatusCreated, answer)

}

func GetQuestion(c *gin.Context) {
	questionID := c.Param("questionid")
	var question QuestionResponse

	qSQL := `
        SELECT q.id, q.user_id, q.title, q.body, q.created_at, COALESCE(SUM(v.vote_type), 0) as score
        FROM questions q
        LEFT JOIN votes v ON q.id = v.post_id AND v.post_type = 'question'
        WHERE q.id = $1
        GROUP BY q.id;
    `
	err := database.Db.QueryRow(context.Background(), qSQL, questionID).Scan(&question.UUID, &question.PostedBy, &question.Title, &question.Body, &question.PostedAt, &question.VoteScore)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch question"})
		return
	}

	var answers []AnswerResponse
	aSQL := `
        SELECT a.id, a.question_id, a.user_id, a.body, a.created_at, COALESCE(SUM(v.vote_type), 0) as score
        FROM answers a
        LEFT JOIN votes v ON a.id = v.post_id AND v.post_type = 'answer'
        WHERE a.question_id = $1
        GROUP BY a.id
        ORDER BY score DESC;
    `
	rows, err := database.Db.Query(context.Background(), aSQL, questionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch answers"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var answer AnswerResponse
		if err := rows.Scan(&answer.UUID, &answer.QuestionID, &answer.PostedBy, &answer.Body, &answer.PostedAt, &answer.VoteScore); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan answer"})
			return
		}
		answers = append(answers, answer)
	}

	response := QuestionWithAnswers{
		Question: question,
		Answers:  answers,
	}
	c.JSON(http.StatusOK, response)
}

func voteHandler(c *gin.Context, postType string, postID string) {
	userID := c.GetString("userID")

	var vd VoteData
	if err := c.BindJSON(&vd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	sql := `
        INSERT INTO votes (user_id, post_id, post_type, vote_type)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, post_id) DO UPDATE SET vote_type = $4;
    `

	_, err := database.Db.Exec(context.Background(), sql, userID, postID, postType, vd.VoteType)
	if err != nil {
		fmt.Printf("Database error while processing vote: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process vote"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "vote recorded"})
}

func VoteQuestion(c *gin.Context) {
	questionID := c.Param("questionid")
	voteHandler(c, "question", questionID)
}

func VoteAnswer(c *gin.Context) {
	answerID := c.Param("answerid")
	voteHandler(c, "answer", answerID)
}

func SummarizeQuestion(c *gin.Context) {
	questionID := c.Param("questionid")
	var question QuestionResponse
	qSQL := `SELECT id, user_id, title, body, created_at FROM questions WHERE id = $1`
	err := database.Db.QueryRow(context.Background(), qSQL, questionID).Scan(&question.UUID, &question.PostedBy, &question.Title, &question.Body, &question.PostedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	var answers []AnswerResponse
	aSQL := `SELECT id, question_id, user_id, body, created_at FROM answers WHERE question_id = $1`
	rows, _ := database.Db.Query(context.Background(), aSQL, questionID)
	defer rows.Close()
	for rows.Next() {
		var answer AnswerResponse
		rows.Scan(&answer.UUID, &answer.QuestionID, &answer.PostedBy, &answer.Body, &answer.PostedAt)
		answers = append(answers, answer)
	}

	var promptBuilder strings.Builder
	promptBuilder.WriteString("Summarize the following question and its answers in a single paragraph.\n\n")
	promptBuilder.WriteString(fmt.Sprintf("Question Title: %s\n", question.Title))
	promptBuilder.WriteString(fmt.Sprintf("Question Body: %s\n\n", question.Body))
	promptBuilder.WriteString("---\nAnswers:\n")
	for i, answer := range answers {
		promptBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, answer.Body))
	}
	prompt := promptBuilder.String()

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenRouter API key is not set"})
		return
	}

	apiURL := "https://openrouter.ai/api/v1/chat/completions"

	requestBody, _ := json.Marshal(map[string]interface{}{
		"model": "mistralai/mistral-7b-instruct:free",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call OpenRouter API"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response from OpenRouter API"})
		return
	}
	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid choice format from OpenRouter API"})
		return
	}
	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid message format from OpenRouter API"})
		return
	}
	summary, ok := message["content"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not extract summary content"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}
