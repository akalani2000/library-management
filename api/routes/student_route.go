package routes

import (
	"library_management/api/controllers"
	"library_management/api/middleware"

	"github.com/gin-gonic/gin"
)

func StudentRouter(router *gin.Engine) {
	studentGroup := router.Group("/students")
	{
		studentGroup.POST("/", controllers.RegisterStudent)
		studentGroup.GET("/", middleware.AuthMiddleware(), controllers.GetStudents)
		studentGroup.GET("/:id", middleware.AuthMiddleware(), controllers.GetStudent)
		studentGroup.PUT("/:id", middleware.AuthMiddleware(), controllers.UpdateStudent)
		studentGroup.PATCH("/:id", middleware.AuthMiddleware(), controllers.PatchStudent)
		studentGroup.DELETE("/:id", middleware.AuthMiddleware(), controllers.DeleteStudent)
	}
}
