package auth

import (
	"context"
	"errors"
	"library_management/api/database"
	"library_management/api/models"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	secretKey       = []byte(os.Getenv("SECRET_KEY"))
	user_collection *mongo.Collection
)

func InitUserController() {
	if database.Client == nil {
		log.Fatal("MongoDB client is not initialized for user collection")
	}
	user_collection = database.Client.Database("library_management").Collection("system_user")
}

// Claims defines the JWT claims structure
type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

// GenerateJWTToken generates a new JWT token and stores it in the database
func GenerateJWTToken(email string) (models.SystemUser, error) {
	var user models.SystemUser

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Create the JWT claims
	claims := &Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			Issuer:    "library-management-system",
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and return it as string
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return user, err
	}

	// Find the user by email and update the token
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"token": tokenString}}

	err = user_collection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
	if err != nil {
		return user, err
	}

	// Fetch the updated user
	err = user_collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return user, err
	}

	return user, nil
}

// VerifyJWTToken verifies the JWT token
func VerifyJWTToken(tokenString string) (bool, error) {
	claims, err := ParseJWTToken(tokenString)
	if err != nil {
		return false, err
	}

	if claims != nil {
		return true, nil
	}
	return false, errors.New("invalid token")
}

// AuthenticateUser authenticates the user by email and password
func AuthenticateUser(email, password string) (models.SystemUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

	var user models.SystemUser
	defer cancel()
	err := user_collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return user, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return user, err
	}

	return user, nil
}

// ParseJWTToken parses the JWT token and returns the claims
func ParseJWTToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
