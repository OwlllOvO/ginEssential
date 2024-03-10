package middleware

import (
	"net/http"
	"owlllovo/ginessential/common"
	"owlllovo/ginessential/model"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get Authorization Header
		tokenString := ctx.GetHeader("Authorization")

		// Validate token formate
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer") {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Insufficient permissions"})
			ctx.Abort()
			return
		}

		tokenString = tokenString[7:]

		token, claims, err := common.ParseToken((tokenString))
		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Insufficient permissions"})
			ctx.Abort()
			return
		}

		// token validated, get userId in claim
		userId := claims.UserId
		DB := common.GetDB()
		var user model.User
		DB.First(&user, userId)

		// user don't exist
		if user.ID == 0 {
			ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Insufficient permissions"})
			ctx.Abort()
			return
		}

		// user exist, write user informations
		ctx.Set("user", user)
		ctx.Next()

	}
}
