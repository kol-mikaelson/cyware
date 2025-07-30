package api

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5" 
	"github.com/kol-mikaelson/cyware-go/internal/auth"
	"github.com/kol-mikaelson/cyware-go/internal/database"
	"github.com/google/uuid"
	"context"
)

type user struct{
	Username string `json:"username" binding:"required"`
	Email string `json:"email" binding:"required"`
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
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type QuestionData struct{
	Title string `json:"title" binding:"required"`
	Body string `json:"body" binding:"required"`
	Category string `json:"category"`
}

type QuestionResponse struct{
	UUID string
	Title string
	Body string
	PostedAt time.Time
	PostedBy string
}

type AnswerData struct{
	Body string `json:"body" binding:"required"`
}

type AnswerResponse struct{
	UUID string `json:"uuid"`
	QuestionID string `json:"question_id"`
	Body string `json:"body"`
	PostedAt time.Time `json:"posted_at"`
	PostedBy string `json:"posted_by"`
}

func Register(c *gin.Context){
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
    ID:           uuid.New().String(), // Generate a new unique ID
    Username:     data.Username,
    Email:        data.Email,
    PasswordHash: Hashed, // Use the hashed password
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

func Login(c *gin.Context){
	var req LoginData
	if err := c.BindJSON(&req); err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := req.Email
	esql := "SELECT id,password_hash FROM users WHERE email = $1"
	var userID string
	var Hashed string;
    err := database.Db.QueryRow(context.Background(), esql, email).Scan(&userID,&Hashed)
    if err != nil {
  		if err == pgx.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
        return
    }
    
    password := req.Password
    err = auth.VerifyPassword(context.Background(),password, Hashed); if err != nil{
    	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
        return
    }
    token,err := auth.GenerateJwt(userID)
    if err != nil {
    	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"token": token})

    
    
}

func CreateQuestion(c *gin.Context){
	userID := c.GetString("userID")
	var qd QuestionData
	if err := c.BindJSON(&qd); err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	quest := QuestionResponse{
		Title: qd.Title,
		Body: qd.Body,
		PostedBy: userID,
		PostedAt: time.Now(),
		UUID: uuid.NewString(),
		
	}
	sqlcomm := "INSERT INTO questions (id, user_id, title, body, created_at) VALUES ($1, $2, $3, $4, $5)"
	_, err := database.Db.Exec(context.Background(), sqlcomm, quest.UUID, quest.PostedBy, quest.Title, quest.Body, quest.PostedAt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create question"})
        return
    }
    c.JSON(http.StatusCreated, quest)
	
}




func CreateAnswer(c *gin.Context){
	userID := c.GetString("userID")
	Qid := c.Param("questionid")
	var ad AnswerData
	if err := c.BindJSON(&ad); err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	answer := AnswerResponse{
		QuestionID: Qid,
		Body: ad.Body,
		PostedBy: userID,
		PostedAt: time.Now(),
		UUID: uuid.NewString(),
	}
	sqlcomm := "INSERT INTO answers (id, user_id, body, created_at, question_id) VALUES ($1, $2, $3, $4, $5)"
	_, err := database.Db.Exec(context.Background(), sqlcomm, answer.UUID, answer.PostedBy, answer.Body, answer.PostedAt, answer.QuestionID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create answer"})
        return
    }
    c.JSON(http.StatusCreated, answer)
	
}
