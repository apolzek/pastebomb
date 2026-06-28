package middleware

import (
	"strings"

	utils "gin-goinc-api/security"

	"github.com/gin-gonic/gin"
)

const Bearer = "Bearer"

func AuthMiddleware(ctx *gin.Context) {
	bearerToken := ctx.GetHeader("Authorization")
	if !strings.Contains(bearerToken, Bearer) {
		ctx.AbortWithStatusJSON(401, gin.H{"message": "token is required"})
		return
	}

	token := strings.TrimSpace(strings.Replace(bearerToken, "Bearer ", "", -1))
	if token == "" {
		ctx.AbortWithStatusJSON(401, gin.H{"message": "token is required"})
		return
	}

	claimsData, err := utils.DecodeToken(token)
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"message": "token is invalid"})
		return
	}

	var userID int64
	switch v := claimsData["id"].(type) {
	case float64:
		userID = int64(v)
	case int:
		userID = int64(v)
	case int64:
		userID = v
	default:
		ctx.AbortWithStatusJSON(401, gin.H{"message": "token is invalid"})
		return
	}

	ctx.Set("claimsData", claimsData)
	ctx.Set("user_id", claimsData["id"])
	ctx.Set("user_int", userID)
	ctx.Set("user_name", claimsData["name"])
	ctx.Set("user_email", claimsData["email"])

	ctx.Next()
}

// UserIDFromContext returns the int64 user id stored by AuthMiddleware.
// Returns 0 if the middleware did not run or the value is missing — callers
// should treat 0 as unauthenticated.
func UserIDFromContext(ctx *gin.Context) int64 {
	v, ok := ctx.Get("user_int")
	if !ok {
		return 0
	}
	if id, ok := v.(int64); ok {
		return id
	}
	return 0
}

func TokenMiddleware(ctx *gin.Context) {
	token := ctx.GetHeader("X-Token")
	if token == "" {
		ctx.AbortWithStatusJSON(401, gin.H{"message": "token is required"})
		return
	}
	if token != "123" {
		ctx.AbortWithStatusJSON(401, gin.H{"message": "token is invalid"})
		return
	}
	ctx.Next()
}
