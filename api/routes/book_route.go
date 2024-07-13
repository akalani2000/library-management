package routes

import (
	"library_management/api/controllers"

	"github.com/gin-gonic/gin"
)

func BookRouter(router *gin.Engine) {
	bookGroup := router.Group("/books")
	{
		bookGroup.POST("/", controllers.CreateBook)
		bookGroup.GET("/", controllers.GetBooks)
		bookGroup.GET("/:id", controllers.GetBook)
		bookGroup.PUT("/:id", controllers.UpdateBook)
		bookGroup.PATCH("/:id", controllers.PatchBook)
		bookGroup.DELETE("/:id", controllers.DeleteBook)
	}
}
