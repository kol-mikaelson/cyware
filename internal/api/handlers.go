package api

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
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
        // If the insert fails (e.g., duplicate email), send an error response.
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
	
}