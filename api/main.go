package main

import (
	"library_management/api/controllers"
	"library_management/api/database"
	"library_management/api/middleware"
	"library_management/api/routes"
	"library_management/api/utils/auth"
	"log"
	"os"
	"strings"
	"time"

	_ "library_management/api/docs" // Import generated docs

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Library Management API
// @version 1.0
// @description This is a sample server for managing a library.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Initialize the MongoDB connection
	database.ConnectDB()

	// .env file setup
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Create a new Stripe client
	sc := &client.API{}
	sc.Init(stripe.Key, nil)

	// Initialize the controller
	controllers.InitBookController()
	controllers.InitUserController()
	controllers.InitStudentController()
	controllers.InitManagerController()
	controllers.InitSubscriptionController()
	auth.InitUserController()
	middleware.InitUserController()

	// Setup the router
	router := gin.Default()

	trustedProxies := os.Getenv("TRUSTED_PROXIES")
	proxyList := []string{}
	if trustedProxies != "" {
		proxyList = strings.Split(trustedProxies, ",")
	}
	err = router.SetTrustedProxies(proxyList)
	if err != nil {
		log.Fatal("Error setting trusted proxies: ", err)
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // Update this with your frontend's origin
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.POST("/webhook", controllers.StripeWebhookHandler)

	routes.UserRoutes(router)
	routes.StudentRouter(router)
	routes.ManagerRouter(router)

	// Authentication middleware
	router.Use(middleware.AuthMiddleware())

	// Setup routes
	routes.BookRouter(router)
	routes.SubscriptionRouter(router)

	// Swagger setup
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Run the router
	log.Fatal(router.Run(":8080"))
}
