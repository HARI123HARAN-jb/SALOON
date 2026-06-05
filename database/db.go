package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Booking represents a salon appointment booking.
type Booking struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name"`
	Email           string             `bson:"email" json:"email"`
	Phone           string             `bson:"phone" json:"phone"`
	Service         string             `bson:"service" json:"service"`
	Date            string             `bson:"date" json:"date"`
	Time            string             `bson:"time" json:"time"`
	AppointmentTime time.Time          `bson:"appointment_time" json:"appointment_time"`
	Status          string             `bson:"status" json:"status"` // pending, confirmed, completed, cancelled
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	Notes           string             `bson:"notes" json:"notes"`
}

// Review represents a client review.
type Review struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Rating    int                `bson:"rating" json:"rating"` // 1-5
	Text      string             `bson:"text" json:"text"`
	Status    string             `bson:"status" json:"status"` // pending, approved, rejected
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

var (
	Client             *mongo.Client
	BookingsCollection *mongo.Collection
	ReviewsCollection  *mongo.Collection
)

// InitDB initializes connection to MongoDB.
func InitDB() error {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = "saloon_db"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Connecting to MongoDB at: %s ...", uri)
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Ping database
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	Client = client
	db := client.Database(dbName)
	BookingsCollection = db.Collection("bookings")
	ReviewsCollection = db.Collection("reviews")

	log.Printf("Connected to MongoDB successfully! Database: %s", dbName)
	return nil
}

// InsertBooking inserts a new booking into MongoDB.
func InsertBooking(booking Booking) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	booking.ID = primitive.NewObjectID()
	booking.Status = "pending"
	booking.CreatedAt = time.Now()

	_, err := BookingsCollection.InsertOne(ctx, booking)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return booking.ID, nil
}

// GetBookings retrieves bookings, optionally filtering by status.
func GetBookings(statusFilter string) ([]Booking, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{}
	if statusFilter != "" {
		filter["status"] = statusFilter
	}

	// Sort by appointment time descending
	opts := options.Find().SetSort(bson.D{{Key: "appointment_time", Value: -1}})

	cursor, err := BookingsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	if bookings == nil {
		bookings = []Booking{}
	}
	return bookings, nil
}

// UpdateBookingStatus updates the status of a specific booking.
func UpdateBookingStatus(id string, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid booking ID: %v", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"status": status}}

	_, err = BookingsCollection.UpdateOne(ctx, filter, update)
	return err
}

// InsertReview inserts a new customer review into MongoDB.
func InsertReview(review Review) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	review.ID = primitive.NewObjectID()
	review.Status = "pending" // requires admin approval
	review.CreatedAt = time.Now()

	_, err := ReviewsCollection.InsertOne(ctx, review)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return review.ID, nil
}

// GetApprovedReviews retrieves reviews that have been approved.
func GetApprovedReviews() ([]Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"status": "approved"}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := ReviewsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}

	if reviews == nil {
		reviews = []Review{}
	}
	return reviews, nil
}

// GetAllReviews retrieves all reviews (for the admin dashboard).
func GetAllReviews() ([]Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := ReviewsCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}

	if reviews == nil {
		reviews = []Review{}
	}
	return reviews, nil
}

// UpdateReviewStatus approves or rejects a review.
func UpdateReviewStatus(id string, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid review ID: %v", err)
	}

	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"status": status}}

	_, err = ReviewsCollection.UpdateOne(ctx, filter, update)
	return err
}

// DeleteReview deletes a review.
func DeleteReview(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid review ID: %v", err)
	}

	filter := bson.M{"_id": objID}
	_, err = ReviewsCollection.DeleteOne(ctx, filter)
	return err
}
