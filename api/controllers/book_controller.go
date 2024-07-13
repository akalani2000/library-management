package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"library_management/api/database"
	"library_management/api/models"
	"library_management/api/utils"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// BookController handles book-related requests.
var book_collection *mongo.Collection

func InitBookController() {
	if database.Client == nil {
		log.Fatal("MongoDB client is not initialized for book collection")
	}
	book_collection = database.Client.Database("library_management").Collection("books")
}

// CreateBook godoc
// @Summary Create a new book
// @Description Create a new book entry
// @Tags Books
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param Title formData string true "Title"
// @Param Author formData string true "Author"
// @Param Publisher formData string true "Publisher"
// @Param PublishDate formData string true "PublishDate"
// @Param ISBN formData string true "ISBN"
// @Param CoverImage formData file true "CoverImage"
// @Param BookPDF formData file true "BookPDF"
// @Param Tags formData []string true "Tags"
// @Success 200 {object} models.Book "Book created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /books [post]
func CreateBook(c *gin.Context) {
	var book models.Book
	var bookrequest models.BookRequest
	if err := c.ShouldBind(&bookrequest); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Handle cover image file upload
	coverFile, err := c.FormFile("CoverImage")
	if err == nil {
		// Validate cover image file extension
		if !utils.AllowedFileExtension(coverFile.Filename, "image") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid cover image format. Only jpg, jpeg, and png are allowed."})
			return
		}

		// Save the cover image file to a directory (e.g., "./uploads")
		coverFilename := filepath.Base(coverFile.Filename)
		coverFilePath := "./uploads/CoverImage/" + coverFilename
		if err := c.SaveUploadedFile(coverFile, coverFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to save cover image"})
			return
		}

		book.CoverImage = coverFilePath
	}

	// Handle book PDF file upload
	pdfFile, err := c.FormFile("BookPDF")
	if err == nil {
		// Validate book PDF file extension
		if !utils.AllowedFileExtension(pdfFile.Filename, "pdf") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid book PDF format. Only pdf and doc are allowed."})
			return
		}

		// Save the book PDF file to a directory (e.g., "./uploads")
		pdfFilename := filepath.Base(pdfFile.Filename)
		pdfFilePath := "./uploads/BookPDF/" + pdfFilename
		if err := c.SaveUploadedFile(pdfFile, pdfFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to save book PDF"})
			return
		}

		book.BookPDF = pdfFilePath
	}

	// Convert BookRequest to Book using mapstructure
	if err := mapstructure.Decode(bookrequest, &book); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	book.ID = primitive.NewObjectID()
	book.CreatedAt = time.Now()
	book.UpdatedAt = time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = book_collection.InsertOne(ctx, book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, book)
}

// GetBooks godoc
// @Summary Get a list of books
// @Description Get a list of all books
// @Tags Books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Book "List of books"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /books [get]
func GetBooks(c *gin.Context) {
	var books []models.Book
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := book_collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var book models.Book
		if err = cursor.Decode(&book); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
			return
		}
		books = append(books, book)
	}
	c.JSON(http.StatusOK, books)
}

// GetBook godoc
// @Summary Get a book by ID
// @Description Get details of a book by ID
// @Tags Books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} models.Book "Book details"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Book not found"
// @Router /books/{id} [get]
func GetBook(c *gin.Context) {
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)
	var book models.Book
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := book_collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, book)
}

// UpdateBook godoc
// @Summary Update a book by ID
// @Description Update details of a book by ID
// @Tags Books
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Param Title formData string true "Title"
// @Param Author formData string true "Author"
// @Param Publisher formData string true "Publisher"
// @Param PublishDate formData string true "PublishDate"
// @Param ISBN formData string true "ISBN"
// @Param CoverImage formData file true "CoverImage"
// @Param BookPDF formData file true "BookPDF"
// @Param Tags formData []string true "Tags"
// @Success 200 {object} models.Book "Updated book details"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Book not found"
// @Router /books/{id} [put]
func UpdateBook(c *gin.Context) {
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)
	var book models.Book
	var bookrequest models.BookRequest
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	if err := c.ShouldBind(&bookrequest); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Fetch the record
	err := book_collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&book)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Book not found"})
		return
	}

	// Handle cover image file upload
	coverFile, err := c.FormFile("CoverImage")
	if err == nil {
		// Validate cover image file extension
		if !utils.AllowedFileExtension(coverFile.Filename, "image") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid cover image format. Only jpg, jpeg, and png are allowed."})
			return
		}

		// Save the cover image file to a directory (e.g., "./uploads")
		coverFilename := filepath.Base(coverFile.Filename)
		coverFilePath := "./uploads/CoverImage/" + coverFilename
		if err := c.SaveUploadedFile(coverFile, coverFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to save cover image"})
			return
		}

		// Delete the existing cover image if it exists
		if book.CoverImage != "" {
			if err := os.Remove(book.CoverImage); err != nil {
				log.Printf("Failed to delete old cover image: %v", err)
			}
		}

		book.CoverImage = coverFilePath
	}

	// Handle book PDF file upload
	pdfFile, err := c.FormFile("BookPDF")
	if err == nil {
		// Validate book PDF file extension
		if !utils.AllowedFileExtension(pdfFile.Filename, "pdf") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid book PDF format. Only pdf and doc are allowed."})
			return
		}

		// Save the book PDF file to a directory (e.g., "./uploads")
		pdfFilename := filepath.Base(pdfFile.Filename)
		pdfFilePath := "./uploads/BookPDF/" + pdfFilename
		if err := c.SaveUploadedFile(pdfFile, pdfFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to save book PDF"})
			return
		}

		// Delete the existing book PDF if it exists
		if book.BookPDF != "" {
			if err := os.Remove(book.BookPDF); err != nil {
				log.Printf("Failed to delete old book PDF: %v", err)
			}
		}

		book.BookPDF = pdfFilePath
	}

	// Convert BookRequest to Book using mapstructure
	if err := mapstructure.Decode(bookrequest, &book); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	book.UpdatedAt = time.Now()
	defer cancel()
	_, err = book_collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": book})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	defer cancel()
	filter := bson.M{"_id": objID}
	err = book_collection.FindOne(ctx, filter).Decode(&book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve book"})
		return
	}

	c.JSON(http.StatusOK, book)
}

// PatchBook godoc
// @Summary Partially update a book by ID
// @Description Partially update details of a book by ID
// @Tags Books
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Param Title formData string false "Title"
// @Param Author formData string false "Author"
// @Param Publisher formData string false "Publisher"
// @Param PublishDate formData string false "PublishDate"
// @Param ISBN formData string false "ISBN"
// @Param CoverImage formData file false "CoverImage"
// @Param BookPDF formData file false "BookPDF"
// @Param Tags formData []string false "Tags"
// @Success 200 {object} models.Book "Updated book details"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Book not found"
// @Router /books/{id} [patch]
func PatchBook(c *gin.Context) {
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)
	var book models.Book
	var bookRequest models.BookRequest

	if err := c.ShouldBind(&bookRequest); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Retrieve the existing book details
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	err := book_collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&book)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Book not found"})
		return
	}

	updateFields := bson.M{}

	if bookRequest.Title != "" {
		updateFields["title"] = bookRequest.Title
	}

	if bookRequest.Author != "" {
		updateFields["author"] = bookRequest.Author
	}

	if bookRequest.Publisher != "" {
		updateFields["publisher"] = bookRequest.Publisher
	}

	if bookRequest.PublishDate != "" {
		updateFields["publish_date"] = bookRequest.PublishDate
	}

	if bookRequest.ISBN != "" {
		updateFields["isbn"] = bookRequest.ISBN
	}

	if len(bookRequest.Tags) > 0 {
		updateFields["tags"] = bookRequest.Tags
	}

	// Handle file upload for CoverImage
	coverFile, err := c.FormFile("CoverImage")
	if err == nil {
		if !utils.AllowedFileExtension(coverFile.Filename, "image") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid cover image format. Only jpg, jpeg, and png are allowed."})
			return
		}

		coverFilename := filepath.Base(coverFile.Filename)
		coverFilePath := "./uploads/" + coverFilename
		if err := c.SaveUploadedFile(coverFile, coverFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to save cover image"})
			return
		}

		// Delete the existing cover image if it exists
		if book.CoverImage != "" {
			if err := os.Remove(book.CoverImage); err != nil {
				log.Printf("Failed to delete old cover image: %v", err)
			}
		}

		updateFields["cover_image"] = coverFilePath
	}

	// Handle file upload for BookPDF
	pdfFile, err := c.FormFile("BookPDF")
	if err == nil {
		if !utils.AllowedFileExtension(pdfFile.Filename, "pdf") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid book PDF format. Only pdf and doc are allowed."})
			return
		}

		pdfFilename := filepath.Base(pdfFile.Filename)
		pdfFilePath := "./uploads/" + pdfFilename
		if err := c.SaveUploadedFile(pdfFile, pdfFilePath); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to save book PDF"})
			return
		}

		// Delete the existing book PDF if it exists
		if book.BookPDF != "" {
			if err := os.Remove(book.BookPDF); err != nil {
				log.Printf("Failed to delete old book PDF: %v", err)
			}
		}

		updateFields["book_pdf"] = pdfFilePath
	}

	updateFields["updated_at"] = time.Now()

	_, err = book_collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateFields})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	err = book_collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve book"})
		return
	}

	c.JSON(http.StatusOK, book)
}

// DeleteBook godoc
// @Summary Delete a book by ID
// @Description Delete a book by ID
// @Tags Books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {string} string "Book deleted successfully"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Book not found"
// @Router /books/{id} [delete]
func DeleteBook(c *gin.Context) {
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := book_collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Book deleted"})
}
