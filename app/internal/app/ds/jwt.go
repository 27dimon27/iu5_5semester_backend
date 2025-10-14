package ds

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	JWTSecret       = "your-secret-key"
	TokenExpireTime = 24 * time.Hour
)

// type User struct {
// 	ID       uint   `json:"id"`
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// }

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}
