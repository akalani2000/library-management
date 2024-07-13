package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"library_management/api/database"
	"library_management/api/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/price"
	"github.com/stripe/stripe-go/v72/product"
	"github.com/stripe/stripe-go/v72/webhook"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var subscription_collection *mongo.Collection
var student_subscription_collection *mongo.Collection
var system_user_collections *mongo.Collection

func InitSubscriptionController() {
	if database.Client == nil {
		log.Fatal("MongoDB client is not initialized for book collection")
	}
	subscription_collection = database.Client.Database("library_management").Collection("subscription")
	student_subscription_collection = database.Client.Database("library_management").Collection("student_subscription")
	system_user_collections = database.Client.Database("library_management").Collection("system_user")
}

func getStripeInterval(subscriptionType models.SubscriptionType) string {
	switch subscriptionType {
	case models.Monthly:
		return "month"
	case models.Quarterly:
		return "quarter"
	case models.Yearly:
		return "year"
	default:
		return ""
	}
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Create a new subscription entry
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subscription body models.SubscriptionCreateRequest true "Subscription object to be registered"
// @Success 200 {object} models.Subscription "Subscription registered successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /subscriptions [post]
func CreateSubscription(c *gin.Context) {
	var subscription_data models.SubscriptionCreateRequest
	if err := c.ShouldBindJSON(&subscription_data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a new product in Stripe
	productParams := &stripe.ProductParams{
		Name:        stripe.String(subscription_data.Title),
		Description: stripe.String(subscription_data.Description),
	}
	newProduct, err := product.New(productParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product in Stripe"})
		return
	}

	// Create a new price for the product in Stripe
	priceParams := &stripe.PriceParams{
		Product:    stripe.String(newProduct.ID),
		UnitAmount: stripe.Int64(int64(subscription_data.Price * 100)), // Stripe uses cents
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
	}

	interval := getStripeInterval(subscription_data.Type)
	if interval != "" {
		priceParams.Recurring = &stripe.PriceRecurringParams{
			Interval: stripe.String(interval),
		}
	}

	newPrice, err := price.New(priceParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create price in Stripe"})
		return
	}

	// Create a subscription in the local database
	subscription := models.Subscription{
		ID:          primitive.NewObjectID(),
		Title:       subscription_data.Title,
		Description: subscription_data.Description,
		Type:        subscription_data.Type,
		Price:       subscription_data.Price,
		ProductID:   newProduct.ID,
		PriceID:     newPrice.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = subscription_collection.InsertOne(ctx, subscription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription in the database"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// GetSubscriptions godoc
// @Summary Get a list of subscriptions
// @Description Get a list of all subscriptions
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Subscription "List of subscriptions"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /subscriptions [get]
func GetSubscriptions(c *gin.Context) {
	var subscriptions []models.Subscription
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := subscription_collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var subscription models.Subscription
		if err = cursor.Decode(&subscription); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
			return
		}
		subscriptions = append(subscriptions, subscription)
	}
	c.JSON(http.StatusOK, subscriptions)
}

// GetSubscription godoc
// @Summary Get a subscription by ID
// @Description Get details of a subscription by ID
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.Subscription "Subscription details"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Subscription not found"
// @Router /subscriptions/{id} [get]
func GetSubscription(c *gin.Context) {
	subscriptionID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(subscriptionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var subscription models.Subscription
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = subscription_collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&subscription)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// UpdateSubscription godoc
// @Summary Update a subscription by ID
// @Description Update details of a subscription by ID
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscription ID"
// @Param subscription body models.SubscriptionUpdateRequest true "Subscription object to be registered"
// @Success 200 {object} models.Subscription "Updated subscription details"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "subscription not found"
// @Router /subscriptions/{id} [put]
func UpdateSubscription(c *gin.Context) {
	subscriptionID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(subscriptionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	var subscription_data models.SubscriptionUpdateRequest
	if err := c.ShouldBindJSON(&subscription_data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{}
	if subscription_data.Title != "" {
		update["title"] = subscription_data.Title
	}
	if subscription_data.Description != "" {
		update["description"] = subscription_data.Description
	}
	if subscription_data.Type != 0 {
		update["type"] = subscription_data.Type
	}
	if subscription_data.Price != 0 {
		// Get the existing subscription from the database
		var existingSubscription models.Subscription
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := subscription_collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&existingSubscription)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}

		// Create a new price in Stripe
		priceParams := &stripe.PriceParams{
			Product:    stripe.String(existingSubscription.ProductID),
			UnitAmount: stripe.Int64(int64(subscription_data.Price * 100)), // Stripe uses cents
			Currency:   stripe.String(string(stripe.CurrencyUSD)),
		}

		interval := getStripeInterval(subscription_data.Type)
		if interval != "" {
			priceParams.Recurring = &stripe.PriceRecurringParams{
				Interval: stripe.String(interval),
			}
		}

		newPrice, err := price.New(priceParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create price in Stripe"})
			return
		}

		update["price"] = subscription_data.Price
		update["price_id"] = newPrice.ID
	}

	update["updated_at"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = subscription_collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription in the database"})
		return
	}

	var subscription models.Subscription
	err = subscription_collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&subscription)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription godoc
// @Summary Delete a subscription by ID
// @Description Delete a subscription by ID
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Subscription ID"
// @Success 200 {string} string "Subscription deleted successfully"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 404 {object} models.ErrorResponse "Subscription not found"
// @Router /subscriptions/{id} [delete]
func DeleteSubscription(c *gin.Context) {
	subscriptionID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(subscriptionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the subscription to get the Stripe product ID
	var subscription models.Subscription
	err = subscription_collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&subscription)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	// Delete the product in Stripe
	_, err = product.Del(subscription.ProductID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product in Stripe"})
		return
	}

	_, err = subscription_collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription in the database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription deleted successfully"})
}

// StudentSubscription godoc
// @Summary Student subscribe to a subscription
// @Description Student subscribe to a subscription
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subscription body models.StudentSubscriptionRequest true "Student is Subscribeing"
// @Success 200 {object} models.StudentSubscription "Subscription registered successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Router /subscriptions/student/subscribe [post]
func StudentSubscription(c *gin.Context) {
	// Retrieve user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	systemUser, ok := user.(models.SystemUser)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if the user role is student
	if systemUser.Role != models.StudentRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to perform this action"})
		return
	}

	// Parse the request body to get the subscription ID
	var student_subscription_request models.StudentSubscriptionRequest
	if err := c.ShouldBindJSON(&student_subscription_request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	subscriptionID := student_subscription_request.SubscriptionID

	// Retrieve the subscription details from the database
	var subscription models.Subscription
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	err := subscription_collection.FindOne(ctx, bson.M{"_id": subscriptionID}).Decode(&subscription)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	var stripeCustomer *stripe.Customer

	if systemUser.StripeCustomerID != "" {
		stripeCustomer, err = customer.Get(systemUser.StripeCustomerID, nil)
		if err != nil {
			stripeCustomer, err = createStripeCustomer(systemUser)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Stripe customer"})
				return
			}
			err = updateStripeCustomerID(ctx, systemUser, stripeCustomer.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Stripe customer ID in the system user table"})
				return
			}
		}
	} else {
		stripeCustomer, err = createStripeCustomer(systemUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Stripe customer"})
			return
		}
		err = updateStripeCustomerID(ctx, systemUser, stripeCustomer.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Stripe customer ID in the system user table"})
			return
		}
	}

	// Create a new StudentSubscription entry
	studentSubscription := models.StudentSubscription{
		ID:             primitive.NewObjectID(),
		SystemUserID:   systemUser.ID,
		SubscriptionID: subscription.ID,
		PriceID:        subscription.PriceID,
		PaymentStatus:  models.Open,
		Status:         models.Pending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	insert_results, err := student_subscription_collection.InsertOne(ctx, studentSubscription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription in the database"})
		return
	}

	insertedID := fmt.Sprintf("%v", insert_results.InsertedID)

	// Create a Stripe Checkout Session
	checkoutParams := &stripe.CheckoutSessionParams{
		Customer:           stripe.String(stripeCustomer.ID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(subscription.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String("https://example.com"),
		CancelURL:  stripe.String("https://example.com"),
		Params: stripe.Params{
			Metadata: map[string]string{
				"student_subscription_id": insertedID,
			},
		},
	}

	session, err := session.New(checkoutParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Stripe checkout session"})
		return
	}

	filter := bson.M{"_id": insert_results.InsertedID}
	update := bson.M{"customer_id": session.Customer.ID, "payment_link": session.URL}
	_, err = student_subscription_collection.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update the fields."})
		return
	}

	var updated_student_subscription models.StudentSubscription

	err = student_subscription_collection.FindOne(ctx, filter).Decode(&updated_student_subscription)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get the data."})
		return
	}

	// Return the Stripe Checkout Session URL
	c.JSON(http.StatusOK, updated_student_subscription)
}

func updateStripeCustomerID(ctx context.Context, systemUser models.SystemUser, stripeCustomerID string) error {
	_, err := system_user_collections.UpdateOne(ctx, bson.M{"_id": systemUser.ID}, bson.M{"$set": bson.M{"stripe_customer_id": stripeCustomerID}})
	return err
}

func createStripeCustomer(systemUser models.SystemUser) (*stripe.Customer, error) {
	customerParams := &stripe.CustomerParams{
		Email: stripe.String(systemUser.Email),
		Name:  stripe.String(systemUser.Name),
	}
	return customer.New(customerParams)
}

func StripeWebhookHandler(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Unable to read request body"})
		return
	}

	// Verify webhook signature
	// endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	endpointSecret := "whsec_e131d4cb82e0c1ff6f3e1a7ac6e6765f2d0b05b8bd07882bac8878e8d8d19d36"
	sigHeader := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		log.Printf("Error verifying webhook signature: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
		return
	}

	// Handle the event
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}
		if session.PaymentStatus == "paid" && session.Status == "complete" {

			subscriptionID := session.Metadata["student_subscription_id"]
			if subscriptionID == "" {
				log.Println("student_subscription_id is empty")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student_subscription_id"})
				return
			}

			fmt.Println("subscriptionID:", subscriptionID)

			// Convert subscriptionID to ObjectID
			objID, err := primitive.ObjectIDFromHex(subscriptionID)
			if err != nil {
				log.Printf("Invalid ObjectID format: %v\n", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student_subscription_id format"})
				return
			}

			filter := bson.M{"_id": objID}
			update := bson.M{"payment_status": models.Paid, "status": models.Subscribed, "stripe_sub_id": session.Subscription.ID}
			var student_subscription_webhook models.StudentSubscription
			student_subscription_collection.FindOne(ctx, filter).Decode(&student_subscription_webhook)
			update_success, err := student_subscription_collection.UpdateOne(ctx, filter, bson.M{"$set": update})
			fmt.Println("update_success", update_success)
			fmt.Println("session.Metadata['student_subscription_id']", session.Metadata["student_subscription_id"])
			fmt.Println("err", err)
			fmt.Println("=================================================")
			fmt.Println("student_subscription_webhook")
			fmt.Println(student_subscription_webhook)
			if err != nil {
				log.Printf("Error while updaing the data: %v\n", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Error while updating the data", "errormessage": err})
				return
			}
		}

		// Fulfill the purchase...
		handleCheckoutSessionCompleted(&session)
	default:
		log.Printf("Unhandled event type: %s\n", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func handleCheckoutSessionCompleted(session *stripe.CheckoutSession) {
	// Implement your logic to handle the checkout session completion event
	log.Printf("Checkout session completed: %s\n", session.ID)
	// e.g., update your database, send a confirmation email, etc.
}
