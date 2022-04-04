package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pianisimo/jwt/controllers"
	"github.com/pianisimo/jwt/middleware"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("users", controllers.GetUsers())
	incomingRoutes.GET("users/:user_id", controllers.GetUser())
}
