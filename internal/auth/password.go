package auth
import (
    "context"
    "golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    return string(hashedPassword), nil
}

func VerifyPassword(ctx context.Context, password string, hashedPassword string) error {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    if err != nil {
        return err
    }
    return nil
}
