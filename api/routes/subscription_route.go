package routes

import (
	"library_management/api/controllers"

	"github.com/gin-gonic/gin"
)

func SubscriptionRouter(router *gin.Engine) {
	subscriptionGroup := router.Group("/subscriptions")
	{
		subscriptionGroup.POST("/", controllers.CreateSubscription)
		subscriptionGroup.GET("/", controllers.GetSubscriptions)
		subscriptionGroup.GET("/:id", controllers.GetSubscription)
		subscriptionGroup.PUT("/:id", controllers.UpdateSubscription)
		subscriptionGroup.DELETE("/:id", controllers.DeleteSubscription)
		subscriptionGroup.POST("/student/subscribe", controllers.StudentSubscription)
	}
}
