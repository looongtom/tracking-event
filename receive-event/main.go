package main

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var (
	kafkaBroker string
	topic       string
)

func handleMain(w http.ResponseWriter, r *http.Request) {
	// log current time
	fmt.Println(time.Now())

	// Create a new Kafka producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaBroker})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create producer: %s", err), http.StatusInternalServerError)
		return
	}
	defer p.Close()

	storeID := os.Getenv("STORE_ID")
	clientID := os.Getenv("CLIENT_ID")
	eventType := os.Getenv("EVENT_TYPE")

	// Group events by bucket_date
	bucketDates := generateRandomBucketDates(5)

	// Generate and insert documents
	for _, bucketDate := range bucketDates {

		// Generate multiple events within the list
		var events []Event
		count := rand.Intn(100) + 1
		for i := 0; i < count; i++ {
			events = append(events, Event{
				ID:        fmt.Sprintf("evt%d", i+1),
				TimeStamp: time.Now().Unix(),
				Status:    randomStatus(),
			})
		}

		tracking := TrackingRecord{
			StoreId:    storeID,
			UserId:     clientID,
			BucketDate: bucketDate.Unix(),
			EventType:  eventType,
			Count:      count,
			ListEvent:  events,
		}
		fmt.Println(tracking)

		serializedBookingRequest, err := json.Marshal(tracking)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to serialize booking request: %s", err), http.StatusInternalServerError)
			return
		}

		// Produce the message to the Kafka topic
		err = produceMessage(p, topic, serializedBookingRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to produce message: %s", err), http.StatusInternalServerError)
			return
		}
		fmt.Println("Message produced successfully!")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Main function executed successfully"))
}

func main() {
	kafkaBroker = os.Getenv("KAFKA_BROKER")
	topic = os.Getenv("KAFKA_TOPIC")
	fmt.Println("Kafka Broker: ", kafkaBroker)
	fmt.Println("Kafka Topic: ", topic)

	http.HandleFunc("/receive-event", handleMain)
	fmt.Println(fmt.Sprintf("Server is listening on port %v...", os.Getenv("SERVER_PORT_RECEIVE_EVENT")))
	server := &http.Server{
		Addr:              fmt.Sprintf(":%v", os.Getenv("SERVER_PORT_RECEIVE_EVENT")),
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func produceMessage(p *kafka.Producer, topic string, message []byte) error {
	// Create a new Kafka message to be produced
	kafkaMessage := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}
	// Produce the Kafka message
	deliveryChan := make(chan kafka.Event)
	err := p.Produce(kafkaMessage, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}
	// Wait for delivery report or error
	e := <-deliveryChan
	m := e.(*kafka.Message)
	// Check for delivery errors
	if m.TopicPartition.Error != nil {
		return fmt.Errorf("delivery failed: %s", m.TopicPartition.Error)
	}
	// Close the delivery channel
	close(deliveryChan)
	return nil
}

// Helper function to generate random bucket dates
func generateRandomBucketDates(numDates int) []time.Time {
	var dates []time.Time
	baseDate, _ := time.Parse(time.RFC3339, "2024-10-01T00:00:00Z")
	for i := 0; i < numDates; i++ {
		dates = append(dates, baseDate.Add(time.Duration(i)*24*time.Hour))
	}
	return dates
}

// Helper function to generate random status
func randomStatus() string {
	statuses := []string{"success", "failed"}
	return statuses[rand.Intn(len(statuses))]
}
