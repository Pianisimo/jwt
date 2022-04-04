package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pianisimo/jwt/helpers"
	"net/http"
)

func Authenticate() gin.HandlerFunc {
	return func(context *gin.Context) {
		clientToken := context.Request.Header.Get("token")

		if clientToken == "" {
			context.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Not Authorization header provided")})
			context.Abort()
			return
		}

		claims, err := helpers.ValidateToken(clientToken)
		if err != "" {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err})
			context.Abort()
			return
		}

		context.Set("email", claims.Email)
		context.Set("first_name", claims.FirstName)
		context.Set("last_name", claims.LastName)
		context.Set("uid", claims.Uid)
		context.Set("user_type", claims.UserType)
		context.Next()
	}
}
