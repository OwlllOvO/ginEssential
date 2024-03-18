package middleware

import (
	"fmt"
	"owlllovo/ginessential/response"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				response.Fail(ctx, gin.H{"error": fmt.Sprint(err)}, "")
			}
		}()

		ctx.Next()
	}

}
