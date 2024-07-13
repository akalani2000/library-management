package middleware

import (
	"context"
	"library_management/api/database"
	"library_management/api/models"
	"library_management/api/utils/auth"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var user_collection *mongo.Collection

func InitUserController() {
	if database.Client == nil {
		log.Fatal("MongoDB client is not initialized for user collection")
	}
	user_collection = database.Client.Database("library_management").Collection("system_user")
}

// AuthMiddleware checks if the token is valid and exists in the database
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the request path requires authentication
		if requiresAuth(c.Request.RequestURI) {
			tokenString := c.GetHeader("Authorization")
			if tokenString == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication is required"})
				c.Abort()
				return
			}

			// Remove "Bearer " prefix from the token string if present
			if len(tokenString) > 7 && strings.ToLower(tokenString[:7]) == "bearer " {
				tokenString = tokenString[7:]
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
				c.Abort()
				return
			}

			// Verify JWT token
			isValid, err := auth.VerifyJWTToken(tokenString)
			if err != nil || !isValid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}

			// Check if the token exists in the database
			var user models.SystemUser
			err = user_collection.FindOne(context.TODO(), bson.M{"token": tokenString}).Decode(&user)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}

			// Continue if token is valid and exists in the database
			c.Set("user", user)
			c.Next()
		}
	}
}

// requiresAuth checks if the request path requires authentication
func requiresAuth(path string) bool {
	// Define paths that require authentication
	authPaths := []string{
		"/swagger",
		"/favicon",
		"/user/login",
		"/user/register",
	}

	// Check if the request path matches any of the paths that require authentication
	for _, p := range authPaths {
		if strings.HasPrefix(path, p) {
			return false
		}
	}
	return true
}
