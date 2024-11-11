package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"kafka-listener/model"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

var (
	groupID     string
	topic       string
	kafkaBroker string
)

func updateEvent(event model.Event) (*model.Event, error) {
	url := fmt.Sprintf("%s:%s/update-event", os.Getenv("SERVER_HOST_UPDATE_EVENT"), os.Getenv("SERVER_PORT_UPDATE_EVENT"))
	method := "POST"

	timestamp := time.Unix(event.TimeStamp, 0).Unix()

	payload := strings.NewReader(fmt.Sprintf(`{"event_id": "%s", "timestamp": %d, "status": "%s"}`, event.ID, timestamp, event.Status))

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err

	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err

	}

	var response model.Event
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &response, nil
}

func saveEventInDb(client *mongo.Client, updatedEvent *model.Event, tracking model.TrackingEvent) error {
	if updatedEvent == nil {
		return fmt.Errorf("event is nil")
	}

	db := client.Database(os.Getenv("MONGO_DB"))
	collection := db.Collection(os.Getenv("MONGO_COLLECTION"))
	//insert event in db
	filter := bson.M{
		"store_id":    tracking.StoreId,
		"client_id":   tracking.UserId,
		"bucket_date": tracking.BucketDate,
		"event_type":  tracking.EventType,
	}
	update := bson.M{
		"$setOnInsert": bson.M{
			"store_id":    tracking.StoreId,
			"client_id":   tracking.UserId,
			"bucket_date": tracking.BucketDate,
			"event_type":  tracking.EventType,
		},
		"$inc": bson.M{
			"count": tracking.Count, // Increment the count field by 1
		},
		"$push": bson.M{
			"list_event": bson.M{
				"event_id":           updatedEvent.ID,
				"timestamp":          updatedEvent.TimeStamp,
				"status_destination": updatedEvent.Status,
			},
		},
	}
	opts := options.Update().SetUpsert(true)
	resp, err := collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		log.Printf("Error upserting document: %v", err)
		return err
	}
	fmt.Println(fmt.Sprintf("Response: %v", resp))
	return nil
}

func createTopicIfNotExists(admin *kafka.AdminClient, topic string) {
	metadata, err := admin.GetMetadata(&topic, false, 5000)
	if err != nil {
		log.Fatalf("Failed to get metadata: %s", err)
	}

	if _, exists := metadata.Topics[topic]; exists {
		fmt.Printf("Topic %s already exists\n", topic)
		return
	}

	topics := []kafka.TopicSpecification{{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := admin.CreateTopics(ctx, topics)
	if err != nil {
		log.Fatalf("Failed to create topic: %s", err)
	}

	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError {
			log.Fatalf("Failed to create topic %s: %v\n", result.Topic, result.Error)
		}
		fmt.Printf("Topic %s created successfully\n", result.Topic)
	}
}

func checkAndCreateCollection(client *mongo.Client, dbName, collectionName string) (bool, error) {
	// List collections in the database
	db := client.Database(dbName)
	collections, err := db.ListCollectionNames(context.TODO(), map[string]interface{}{})
	if err != nil {
		return false, fmt.Errorf("failed to list collections: %w", err)
	}

	// Check if the collection already exists
	for _, col := range collections {
		if col == collectionName {
			return true, nil
		}
	}

	// Create the collection if it doesn't exist
	err = db.CreateCollection(context.TODO(), collectionName)
	if err != nil {
		return false, fmt.Errorf("failed to create collection: %w", err)
	}

	return false, nil
}

func main() {
	kafkaBroker = os.Getenv("KAFKA_BROKER")
	topic = os.Getenv("KAFKA_TOPIC")
	groupID = os.Getenv("KAFKA_GROUP_ID")
	fmt.Println("Kafka Broker: ", kafkaBroker)
	fmt.Println("Kafka Topic: ", topic)
	fmt.Println("Kafka Group ID: ", groupID)

	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s", os.Getenv("MONGO_URI")))
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}(client, context.Background())

	dbName := os.Getenv("MONGO_DB")
	collectionName := os.Getenv("MONGO_COLLECTION")

	collectionExists, err := checkAndCreateCollection(client, dbName, collectionName)
	if err != nil {
		log.Fatalf("Failed to check or create collection: %s", err)
	}

	if collectionExists {
		fmt.Printf("Collection %s already exists in database %s\n", collectionName, dbName)
	} else {
		fmt.Printf("Collection %s created in database %s\n", collectionName, dbName)
	}

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": kafkaBroker})
	if err != nil {
		log.Fatalf("Failed to create Admin client: %s", err)
	}
	defer adminClient.Close()

	// Check if topic exists and create if not
	createTopicIfNotExists(adminClient, topic)

	kafkaBrokerServer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaBroker})
	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		return
	}
	defer kafkaBrokerServer.Close()

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaBroker,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		fmt.Printf("Failed to create consumer: %s\n", err)
		return
	}
	defer func(c *kafka.Consumer) {
		err := c.Close()
		if err != nil {
			fmt.Printf("Failed to close consumer: %s\n", err)
		}
	}(c)

	// Subscribe to the Kafka topic
	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		fmt.Printf("Failed to subscribe to topic: %s\n", err)
		return
	}

	// Setup a channel to handle OS signals for graceful shutdown
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	// Start consuming messages
	fmt.Printf("Consuming messages from topic: %s\n", topic)
	run := true
	for run == true {
		select {
		case sig := <-sigchan:
			fmt.Printf("Received signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				// Process the consumed message
				var tracking model.TrackingEvent
				err := json.Unmarshal(e.Value, &tracking)
				if err != nil {
					fmt.Printf("Failed to deserialize message: %s\n", err)
					continue
				}
				fmt.Printf("Received booking: %+v\n", tracking)

				updatedEvent, err := updateEvent(tracking.Event)
				if err != nil {
					fmt.Println(err)
				}

				err = saveEventInDb(client, updatedEvent, tracking)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(time.Now())
				fmt.Println("===========================Message consumed successfully!=============================")
			case kafka.Error:
				// Handle Kafka errors
				fmt.Printf("Error: %v\n", e)
			}
		}
	}
}
