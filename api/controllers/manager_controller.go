package controllers

import (
	"context"
	"library_management/api/database"
	"library_management/api/models"
	"library_management/api/utils/auth"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserController handles user-related requests.
var manager_collection *mongo.Collection
var systen_user_for_manager_collection *mongo.Collection

func InitManagerController() {
	if database.Client == nil {
		log.Fatal("MongoDB client is not initialized for user collection")
	}
	manager_collection = database.Client.Database("library_management").Collection("manager")
	systen_user_for_manager_collection = database.Client.Database("library_management").Collection("system_user")
}

// RegisterManager godoc
// @Summary Register a new manager
// @Description Register a new manager entry
// @Tags Managers
// @Accept json
// @Produce json
// @Param student body models.ManagerRegister true "Manager object to be registered"
// @Success 200 {object} models.Student "Manager registered successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /managers [post]
func RegisterManager(c *gin.Context) {
	var managerRegister models.ManagerRegister
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

	defer cancel()
	if err := c.ShouldBindJSON(&managerRegister); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Hashing the password
	hashedPassword, err := auth.HashPassword(managerRegister.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password."})
		return
	}

	// Create SystemUser entry
	systemUser := models.SystemUser{
		ID:          primitive.NewObjectID(),
		Email:       managerRegister.Email,
		Name:        managerRegister.FirstName + " " + managerRegister.LastName,
		Password:    hashedPassword, // Ensure to hash the password before saving
		IsSuperuser: false,
		Role:        models.ManagerRole,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = systen_user_for_manager_collection.InsertOne(ctx, systemUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	manager := models.Manager{
		ID:           primitive.NewObjectID(),
		SystemUserID: systemUser.ID,
		FirstName:    managerRegister.FirstName,
		LastName:     managerRegister.LastName,
		Email:        managerRegister.Email,
		ManagerID:    managerRegister.ManagerID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = manager_collection.InsertOne(ctx, manager)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, manager)
}

// GetManager godoc
// @Summary Get a list of managers
// @Description Get a list of all managers
// @Tags Managers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Manager "List of managers"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /managers [get]
func GetManagers(c *gin.Context) {
	var managers []models.Manager
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := manager_collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var manager models.Manager
		if err = cursor.Decode(&manager); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
			return
		}
		managers = append(managers, manager)
	}
	c.JSON(http.StatusOK, managers)
}

// GetManager godoc
// @Summary Get a manager by ID
// @Description Get details of a manager by ID
// @Tags Managers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Manager ID"
// @Success 200 {object} models.Manager "Manager details"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Manager not found"
// @Router /managers/{id} [get]
func GetManager(c *gin.Context) {
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)
	var manager models.Manager
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := manager_collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&manager)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, manager)
}

// UpdateManager godoc
// @Summary Update a manager by ID
// @Description Update details of a manager by ID
// @Tags Managers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Manager ID"
// @Param manager body models.ManagerUpdate true "Manager object to be registered"
// @Success 200 {object} models.Manager "Updated manager details"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "manager not found"
// @Router /managers/{id} [put]
func UpdateManager(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Manager ID is required"})
		return
	}

	var managerUpdate models.ManagerUpdate
	if err := c.ShouldBindJSON(&managerUpdate); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	managerID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid manager ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the existing manager
	var existingManager models.Manager
	err = manager_collection.FindOne(ctx, bson.M{"_id": managerID}).Decode(&existingManager)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Manager not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to find manager"})
		}
		return
	}

	// Update the student manager
	existingManager.FirstName = managerUpdate.FirstName
	existingManager.LastName = managerUpdate.LastName
	existingManager.Email = managerUpdate.Email
	existingManager.ManagerID = managerUpdate.ManagerID
	existingManager.UpdatedAt = time.Now()

	// Save the updated student
	_, err = manager_collection.ReplaceOne(ctx, bson.M{"_id": managerID}, existingManager)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update manager"})
		return
	}

	c.JSON(http.StatusOK, existingManager)
}

// PatchManager godoc
// @Summary Partially update a manager by ID
// @Description Partially update details of a manager by ID
// @Tags Managers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Manager ID"
// @Param manager body models.ManagerUpdate true "Manager object to be registered"
// @Success 200 {object} models.Manager "Updated manager details"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "manager not found"
// @Router /managers/{id} [patch]
func PatchManager(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Manager ID is required"})
		return
	}

	var managerUpdate models.ManagerUpdate
	if err := c.ShouldBindJSON(&managerUpdate); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	managerID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid manager ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the existing student
	var existingManager models.Manager
	err = manager_collection.FindOne(ctx, bson.M{"_id": managerID}).Decode(&existingManager)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Student not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to find student"})
		}
		return
	}

	// Create update document
	update := bson.M{"$set": bson.M{}}
	if managerUpdate.FirstName != "" {
		update["$set"].(bson.M)["first_name"] = managerUpdate.FirstName
	}
	if managerUpdate.LastName != "" {
		update["$set"].(bson.M)["last_name"] = managerUpdate.LastName
	}
	if managerUpdate.Email != "" {
		update["$set"].(bson.M)["email"] = managerUpdate.Email
	}
	if managerUpdate.ManagerID != "" {
		update["$set"].(bson.M)["manager_id"] = managerUpdate.ManagerID
	}
	update["$set"].(bson.M)["updated_at"] = time.Now()

	// Update the student
	_, err = manager_collection.UpdateOne(ctx, bson.M{"_id": managerID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update manager"})
		return
	}

	// Return the updated student
	err = student_collection.FindOne(ctx, bson.M{"_id": managerID}).Decode(&existingManager)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch updated manager"})
		return
	}

	c.JSON(http.StatusOK, existingManager)
}

// DeleteManager godoc
// @Summary Delete a manager by ID
// @Description Delete a manager by ID
// @Tags Managers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Manager ID"
// @Success 200 {string} string "Manager deleted successfully"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Manager not found"
// @Router /managers/{id} [delete]
func DeleteManager(c *gin.Context) {
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := manager_collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Manager deleted"})
}
