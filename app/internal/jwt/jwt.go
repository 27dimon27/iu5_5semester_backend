package pkg

import (
	"time"

	"software/internal/app/ds"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// Генерация JWT токена
func GenerateToken(user ds.Users) (string, error) {
	claims := &ds.Claims{
		UserID: uint(user.ID),
		Role:   user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "software-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(ds.JWTSecret))
}

// Валидация JWT токена
func ValidateToken(tokenString string) (*ds.Claims, error) {
	claims := &ds.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(ds.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// Хеширование пароля
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Проверка пароля
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
