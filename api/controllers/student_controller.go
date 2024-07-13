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
var student_collection *mongo.Collection
var systen_user_for_student_collection *mongo.Collection

func InitStudentController() {
	if database.Client == nil {
		log.Fatal("MongoDB client is not initialized for user collection")
	}
	student_collection = database.Client.Database("library_management").Collection("student")
	systen_user_for_student_collection = database.Client.Database("library_management").Collection("system_user")
}

// RegisterStudent godoc
// @Summary Register a new student
// @Description Register a new student entry
// @Tags Students
// @Accept json
// @Produce json
// @Param student body models.StudentRegister true "Student object to be registered"
// @Success 200 {object} models.Student "Student registered successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /students [post]
func RegisterStudent(c *gin.Context) {
	var studentRegister models.StudentRegister
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

	defer cancel()
	if err := c.ShouldBindJSON(&studentRegister); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Hashing the password
	hashedPassword, err := auth.HashPassword(studentRegister.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password."})
		return
	}

	// Create SystemUser entry
	systemUser := models.SystemUser{
		ID:          primitive.NewObjectID(),
		Email:       studentRegister.Email,
		Name:        studentRegister.FirstName + " " + studentRegister.LastName,
		Password:    hashedPassword, // Ensure to hash the password before saving
		IsSuperuser: false,
		Role:        models.StudentRole,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = systen_user_for_student_collection.InsertOne(ctx, systemUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	student := models.Student{
		ID:           primitive.NewObjectID(),
		SystemUserID: systemUser.ID,
		FirstName:    studentRegister.FirstName,
		LastName:     studentRegister.LastName,
		Email:        studentRegister.Email,
		StudentID:    studentRegister.StudentID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = student_collection.InsertOne(ctx, student)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, student)
}

// GetStudents godoc
// @Summary Get a list of students
// @Description Get a list of all students
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Student "List of students"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /students [get]
func GetStudents(c *gin.Context) {
	var students []models.Student
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := student_collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var student models.Student
		if err = cursor.Decode(&student); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
			return
		}
		students = append(students, student)
	}
	c.JSON(http.StatusOK, students)
}

// GetStudent godoc
// @Summary Get a student by ID
// @Description Get details of a student by ID
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Success 200 {object} models.Student "Student details"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Student not found"
// @Router /students/{id} [get]
func GetStudent(c *gin.Context) {
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)
	var student models.Student
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := student_collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&student)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, student)
}

// UpdateStudent godoc
// @Summary Update a student by ID
// @Description Update details of a student by ID
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Param student body models.StudentUpdate true "Student object to be registered"
// @Success 200 {object} models.Student "Updated student details"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "student not found"
// @Router /students/{id} [put]
func UpdateStudent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Student ID is required"})
		return
	}

	var studentUpdate models.StudentUpdate
	if err := c.ShouldBindJSON(&studentUpdate); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	studentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid student ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the existing student
	var existingStudent models.Student
	err = student_collection.FindOne(ctx, bson.M{"_id": studentID}).Decode(&existingStudent)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Student not found"})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to find student"})
		}
		return
	}

	// Update the student fields
	existingStudent.FirstName = studentUpdate.FirstName
	existingStudent.LastName = studentUpdate.LastName
	existingStudent.Email = studentUpdate.Email
	existingStudent.StudentID = studentUpdate.StudentID
	existingStudent.UpdatedAt = time.Now()

	// Save the updated student
	_, err = student_collection.ReplaceOne(ctx, bson.M{"_id": studentID}, existingStudent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update student"})
		return
	}

	c.JSON(http.StatusOK, existingStudent)
}

// PatchStudent godoc
// @Summary Partially update a student by ID
// @Description Partially update details of a student by ID
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Param student body models.StudentUpdate true "Student object to be registered"
// @Success 200 {object} models.Student "Updated student details"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "student not found"
// @Router /students/{id} [patch]
func PatchStudent(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Student ID is required"})
		return
	}

	var studentUpdate models.StudentUpdate
	if err := c.ShouldBindJSON(&studentUpdate); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	studentID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid student ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the existing student
	var existingStudent models.Student
	err = student_collection.FindOne(ctx, bson.M{"_id": studentID}).Decode(&existingStudent)
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
	if studentUpdate.FirstName != "" {
		update["$set"].(bson.M)["first_name"] = studentUpdate.FirstName
	}
	if studentUpdate.LastName != "" {
		update["$set"].(bson.M)["last_name"] = studentUpdate.LastName
	}
	if studentUpdate.Email != "" {
		update["$set"].(bson.M)["email"] = studentUpdate.Email
	}
	if studentUpdate.StudentID != "" {
		update["$set"].(bson.M)["student_id"] = studentUpdate.StudentID
	}
	update["$set"].(bson.M)["updated_at"] = time.Now()

	// Update the student
	_, err = student_collection.UpdateOne(ctx, bson.M{"_id": studentID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update student"})
		return
	}

	// Return the updated student
	err = student_collection.FindOne(ctx, bson.M{"_id": studentID}).Decode(&existingStudent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to fetch updated student"})
		return
	}

	c.JSON(http.StatusOK, existingStudent)
}

// DeleteStudent godoc
// @Summary Delete a student by ID
// @Description Delete a student by ID
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Success 200 {string} string "Student deleted successfully"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Student not found"
// @Router /students/{id} [delete]
func DeleteStudent(c *gin.Context) {
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := student_collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Student deleted"})
}
