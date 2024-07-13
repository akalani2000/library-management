package routes

import (
	"library_management/api/controllers"

	"github.com/gin-gonic/gin"
)

// AuthRoutes sets up the authentication routes
func UserRoutes(router *gin.Engine) {
	userGroup := router.Group("/user")
	{
		userGroup.POST("/login", controllers.Login)
		userGroup.POST("/logout", controllers.Logout)
		userGroup.POST("/register", controllers.Register)
	}
}
