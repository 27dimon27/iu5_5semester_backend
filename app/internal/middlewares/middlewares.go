package middlewares

import (
	"errors"
	"net/http"

	"software/internal/app/redis"
	"software/internal/app/role"
	pkg "software/internal/jwt"

	red "github.com/go-redis/redis/v8"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(redisClient *redis.Client, roles ...role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userRole := ""
		userID := 0

		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" {
			userRole = "guest"
		}

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		err := redisClient.CheckJWTInBlacklist(ctx.Request.Context(), tokenString)
		if err == nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "token blacklisted"})
			return
		}
		if !errors.Is(err, red.Nil) {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		claims, err := pkg.ValidateToken(tokenString)
		if userRole != "guest" {
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "Invalid token",
				})
				return
			} else {
				userID = int(claims.UserID)
				userRole = claims.Role
			}
		}

		ctx.Set("role", userRole)
		ctx.Set("userID", userID)

		for _, curRole := range roles {
			if role.RoleToString(curRole) == userRole {
				ctx.Next()
				return
			}
		}

		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "unauthorized",
			"role":  userRole,
		})
	}
}
