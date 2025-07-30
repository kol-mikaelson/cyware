package auth

import (
	"os"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

var Secret = []byte(os.Getenv("JWT_SECRET"))
func GenerateJwt(userID string)(string,error){
	claims:= jwt.MapClaims{
		"user_id":userID,
		"exp":time.Now().Add(time.Hour * 24).Unix(),
		"iat":time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(Secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
	
}