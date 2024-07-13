package routes

import (
	"library_management/api/controllers"
	"library_management/api/middleware"

	"github.com/gin-gonic/gin"
)

func ManagerRouter(router *gin.Engine) {
	managerGroup := router.Group("/managers")
	{
		managerGroup.POST("/", controllers.RegisterManager)
		managerGroup.GET("/", middleware.AuthMiddleware(), controllers.GetManagers)
		managerGroup.GET("/:id", middleware.AuthMiddleware(), controllers.GetManager)
		managerGroup.PUT("/:id", middleware.AuthMiddleware(), controllers.UpdateManager)
		managerGroup.PATCH("/:id", middleware.AuthMiddleware(), controllers.PatchManager)
		managerGroup.DELETE("/:id", middleware.AuthMiddleware(), controllers.DeleteManager)
	}
}
