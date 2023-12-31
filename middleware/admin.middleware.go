package middleware

import (
	"github.com/gin-gonic/gin"
)

func AdminMiddleware(ctx *gin.Context) {
	adminHeader := ctx.GetHeader("owner")

	if adminHeader != "apolzek" {
		ctx.AbortWithStatusJSON(401, gin.H{
			"message": "oops, you don't have access here",
		})
	}

	ctx.Next()
}
