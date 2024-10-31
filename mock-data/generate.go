package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"
	"tracking_event/model"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Document represents the optimized structure

func main() {
	// Replace with your MongoDB connection details
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Set up the database and collection
	//db := client.Database("test")
	//collection := db.Collection("tracking_event")

	// Fixed values
	storeID := "store1"
	clientID := "client1"
	eventType := "add_to_cart"

	// Group events by bucket_date
	bucketDates := generateRandomBucketDates(100)

	// Generate and insert documents
	for _, bucketDate := range bucketDates {

		// Generate multiple events within the list
		var events []model.Event
		for i := 0; i < rand.Intn(5)+1; i++ {
			events = append(events, model.Event{
				ID:        fmt.Sprintf("evt%d", i),
				TimeStamp: bucketDate,
				Status:    randomStatus(),
			})
		}

		doc := model.TrackingRecord{
			StoreId:    storeID,
			UserId:     clientID,
			BucketDate: bucketDate,
			EventType:  eventType,
			Count:      len(bucketDates),
			ListEvent:  events,
		}

		fmt.Println(doc)

		//_, err := collection.InsertOne(context.Background(), doc)
		//if err != nil {
		//	log.Printf("Error inserting document: %v", err)
		//}
	}

	fmt.Println("Inserted optimized documents successfully!")
}

// Helper function to generate random bucket dates
func generateRandomBucketDates(numDates int) []int64 {
	var dates []int64
	baseDate := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < numDates; i++ {
		dates = append(dates, baseDate.Add(time.Duration(i)*24*time.Hour).Unix())
	}
	return dates
}

// Helper function to generate random status
func randomStatus() string {
	statuses := []string{"success", "failed"}
	return statuses[rand.Intn(len(statuses))]
}
