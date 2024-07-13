package controllers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"library_management/api/database"
	"library_management/api/models"
	"library_management/api/services"
	"library_management/api/utils/auth"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserController handles user-related requests.
var system_user_collection *mongo.Collection

func InitUserController() {
	if database.Client == nil {
		log.Fatal("MongoDB client is not initialized for user collection")
	}
	system_user_collection = database.Client.Database("library_management").Collection("system_user")
}

// Register godoc
// @Summary Register a new user
// @Description Registers a new user with hashed password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body models.SystemUserRegister true "User object to be created"
// @Success 200 {object} models.SystemUserRegisterResponse "User registered successfully"
// @Failure 400 {string} string "Invalid request"
// @Router /user/register [post]
func Register(c *gin.Context) {
	var user models.SystemUser
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

	defer cancel()
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request."})
		return
	}

	// check email in database
	err := system_user_collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&user)
	defer cancel()
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists."})
		return
	}

	// Hashing the password
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password."})
		return
	}

	// Addeding the user data in the database
	defer cancel()
	user.Password = hashedPassword
	user.IsSuperuser = true
	user.Role = models.SuperUser
	insert_result, err := system_user_collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Get the user after insertion
	filter := bson.M{"_id": insert_result.InsertedID}
	defer cancel()
	err = system_user_collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	// Send welcome email
	subject := "Welcome to Our Library"
	body := "<h1>Welcome to Our Library</h1><p>Thank you for registering, " + user.Name + "!</p>"
	if err := services.SendEmail(user.Email, subject, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	systemUserResponse := models.SystemUserRegisterResponse{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		IsSuperuser: user.IsSuperuser,
		Role:        user.Role,
	}

	c.JSON(http.StatusOK, systemUserResponse)
}

// Login godoc
// @Summary Log in
// @Description Logs in the user and returns a JWT token
// @Tags Authentication
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param Email formData string true "Email"
// @Param Password formData string true "Password"
// @Success 200 {object} models.SystemUserLoginResponse "Login successfully"
// @Failure 400 {string} string "Invalid username or password"
// @Router /user/login [post]
func Login(c *gin.Context) {
	var user models.SystemUser
	var userlogin models.SystemUserLogin
	if err := c.ShouldBind(&userlogin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Validate user credentials
	user, err := auth.AuthenticateUser(userlogin.Email, userlogin.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}

	isValid, err := auth.VerifyJWTToken(user.Token)

	// Generate JWT token
	if user.Token == "" || user.Token != "" && err != nil && !isValid {
		user, err = auth.GenerateJWTToken(user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
	}

	systemUserResponse := models.SystemUserLoginResponse{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		IsSuperuser: user.IsSuperuser,
		Role:        user.Role,
		Token:       user.Token,
	}

	c.JSON(http.StatusOK, systemUserResponse)
}

// Logout godoc
// @Summary Log out
// @Description Logs out the user and invalidates the JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse "Logout successful"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /user/logout [post]
func Logout(c *gin.Context) {
	var user models.SystemUser
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
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

	filter := bson.M{"token": tokenString}

	err := system_user_collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	// Remove the token from the database
	_, err = system_user_collection.UpdateOne(context.TODO(), filter, bson.M{"$unset": bson.M{"token": ""}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}
