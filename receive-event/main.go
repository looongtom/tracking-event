package main

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"math/rand"
	"net/http"
	"os"
	"receive-event/model"
	"strconv"
	"sync"
	"time"
)

var (
	kafkaBroker string
	topic       string
)

func generateMockData(nStores, mEventTypes, mEvents, nClients int, bucketDate time.Time, p *kafka.Producer) error {

	storePrefix := "store"
	clientPrefix := "client"
	eventTypes := make([]string, mEventTypes)
	for i := 0; i < mEventTypes; i++ {
		eventTypes[i] = fmt.Sprintf("event_type%d", i+1)
	}

	var wg sync.WaitGroup

	errChan := make(chan error, mEvents)

	for i := 0; i < mEvents; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			indexStore := rand.Intn(nStores) + 1
			storeID := fmt.Sprintf("%s%d", storePrefix, indexStore)
			indexClient := rand.Intn(nClients) + 1
			clientID := fmt.Sprintf("%s%d", clientPrefix, indexClient)
			eventType := eventTypes[rand.Intn(mEventTypes)]
			eventID := fmt.Sprintf("evt%d", i+1)
			timestamp := bucketDate.Add(time.Duration(rand.Intn(24)) * time.Hour).Add(time.Duration(rand.Intn(60)) * time.Minute)
			status := []string{"success", "failed"}[rand.Intn(2)]

			event := model.Event{
				ID:        eventID,
				TimeStamp: timestamp.Unix(),
				Status:    status,
			}
			trackingEvent := model.TrackingEvent{
				StoreId:    storeID,
				UserId:     clientID,
				BucketDate: bucketDate.UnixNano(),
				EventType:  eventType,
				Count:      1,
				Event:      event,
			}
			serializedBookingRequest, err := json.Marshal(trackingEvent)
			if err != nil {
				//http.Error(w, fmt.Sprintf("Failed to serialize booking request: %s", err), http.StatusInternalServerError)
				errChan <- fmt.Errorf("Failed to serialize booking request: %s", err)
				return
			}

			// Produce the message to the Kafka topic
			err = produceMessage(p, topic, serializedBookingRequest)
			if err != nil {
				//http.Error(w, fmt.Sprintf("Failed to produce message: %s", err), http.StatusInternalServerError)
				errChan <- fmt.Errorf("Failed to produce message: %s", err)
				return
			}

			errChan <- nil
		}(i)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	fmt.Println("===========================Message produced successfully!=============================")
	return nil
}

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
		var wg sync.WaitGroup

		// Generate multiple events within the list
		maxCount := os.Getenv("MAX_AMOUNT_EVENT")
		maxValue, ok := strconv.Atoi(maxCount)
		if ok != nil {
			fmt.Println("Error: ", ok)
			fmt.Println("set max value into 10")
			maxValue = 10
		}
		errChan := make(chan error, maxValue)
		fmt.Println("Count: ", maxValue)
		for i := 0; i < maxValue; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				tracking := model.TrackingEvent{
					StoreId:    storeID,
					UserId:     clientID,
					BucketDate: bucketDate.UnixNano(),
					EventType:  eventType,
					Count:      1,
					Event: model.Event{
						ID:        fmt.Sprintf("evt%d", i+1),
						TimeStamp: time.Now().UnixNano(),
						Status:    randomStatus(),
					},
				}
				fmt.Println(tracking)

				serializedBookingRequest, err := json.Marshal(tracking)
				if err != nil {
					//http.Error(w, fmt.Sprintf("Failed to serialize booking request: %s", err), http.StatusInternalServerError)
					errChan <- fmt.Errorf("Failed to serialize booking request: %s", err)
					return
				}

				// Produce the message to the Kafka topic
				err = produceMessage(p, topic, serializedBookingRequest)
				if err != nil {
					//http.Error(w, fmt.Sprintf("Failed to produce message: %s", err), http.StatusInternalServerError)
					errChan <- fmt.Errorf("Failed to produce message: %s", err)
					return
				}
				errChan <- nil
			}(i)
		}
		wg.Wait()
		close(errChan)

		for err := range errChan {
			if err != nil {
				fmt.Println(err)
				http.Error(w, fmt.Sprintf("Error: %s", err), http.StatusInternalServerError)
			}
		}

		fmt.Println("===========================Message produced successfully!=============================")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Main function executed successfully"))
}

func handleMainV2(w http.ResponseWriter, r *http.Request) {
	// log current time
	fmt.Println(time.Now())

	// Create a new Kafka producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaBroker})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create producer: %s", err), http.StatusInternalServerError)
		return
	}
	defer p.Close()

	totalStore := os.Getenv("TOTAL_STORE")
	totalStoreValue, _ := strconv.Atoi(totalStore)
	totalClient := os.Getenv("TOTAL_CLIENT")
	totalClientValue, _ := strconv.Atoi(totalClient)
	totalEventType := os.Getenv("TOTAL_EVENT_TYPE")
	totalEventTypeValue, _ := strconv.Atoi(totalEventType)
	maxEvent, _ := strconv.Atoi(os.Getenv("MAX_AMOUNT_EVENT"))

	err = generateMockData(totalStoreValue, totalEventTypeValue, maxEvent, totalClientValue, time.Now(), p)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate mock data: %s", err), http.StatusInternalServerError)
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
	http.HandleFunc("/receive-event-v2", handleMainV2)
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
