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

func saveEventInDb(updatedEvent *model.Event, tracking model.TrackingEvent) error {
	if updatedEvent == nil {
		return fmt.Errorf("event is nil")
	}
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

func main() {
	kafkaBroker = os.Getenv("KAFKA_BROKER")
	topic = os.Getenv("KAFKA_TOPIC")
	groupID = os.Getenv("KAFKA_GROUP_ID")
	fmt.Println("Kafka Broker: ", kafkaBroker)
	fmt.Println("Kafka Topic: ", topic)
	fmt.Println("Kafka Group ID: ", groupID)

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

				err = saveEventInDb(updatedEvent, tracking)
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
