package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type TrackingRecord struct {
	ID         string  `json:"id"`
	StoreId    string  `json:"store_id"`
	UserId     string  `json:"user_id"`
	BucketDate int64   `json:"bucket_date"`
	EventType  string  `json:"event_type"`
	Count      int     `json:"count"`
	ListEvent  []Event `json:"list_events"`
}

type Event struct {
	ID        string `json:"event_id"`
	TimeStamp int64  `json:"timestamp"`
	Status    string `json:"status"`
}
type TrackingEvent struct {
	StoreId    string `json:"store_id"`
	UserId     string `json:"client_id"`
	BucketDate int64  `json:"bucket_date"`
	EventType  string `json:"event_type"`
	Count      int    `json:"count"`
	Event      Event  `json:"event"`
}

type ListEvent struct {
	EventID           string    `json:"event_id"`
	Timestamp         time.Time `json:"timestamp"`
	StatusDestination string    `json:"status_destination"`
}

type MockData struct {
	StoreID    string      `json:"store_id"`
	ClientID   string      `json:"client_id"`
	BucketDate time.Time   `json:"bucket_date"`
	EventType  string      `json:"event_type"`
	Count      int         `json:"count"`
	ListEvent  []ListEvent `json:"list_event"`
}

// GenerateMockData generates mock data for stores, clients, and event types with list of events
func GenerateMockData(nStores, mEventTypes, nClients int) []MockData {

	storePrefix := "store"
	clientPrefix := "client"
	eventTypes := make([]string, mEventTypes)
	for i := 0; i < mEventTypes; i++ {
		eventTypes[i] = fmt.Sprintf("event_type%d", i+1)
	}

	var mockDataList []MockData

	for i := 0; i < nStores; i++ {
		for j := 0; j < nClients; j++ {
			storeID := fmt.Sprintf("%s%d", storePrefix, i+1)
			clientID := fmt.Sprintf("%s%d", clientPrefix, j+1)
			bucketDate := time.Now().AddDate(0, 0, -rand.Intn(30)) // Random date within the last 30 days
			eventType := eventTypes[rand.Intn(mEventTypes)]

			// Generate between 1 and 1000 events for list_event
			numEvents := rand.Intn(1000) + 1
			listEvents := make([]ListEvent, numEvents)
			for k := 0; k < numEvents; k++ {
				eventID := fmt.Sprintf("evt%d", k+1)
				timestamp := bucketDate.Add(time.Duration(rand.Intn(24)) * time.Hour).Add(time.Duration(rand.Intn(60)) * time.Minute)
				status := []string{"success", "failed"}[rand.Intn(2)]

				listEvents[k] = ListEvent{
					EventID:           eventID,
					Timestamp:         timestamp,
					StatusDestination: status,
				}
			}

			mockData := MockData{
				StoreID:    storeID,
				ClientID:   clientID,
				BucketDate: bucketDate,
				EventType:  eventType,
				Count:      numEvents,
				ListEvent:  listEvents,
			}

			mockDataList = append(mockDataList, mockData)
		}
	}

	return mockDataList
}

func generateMockDataV2(nStores, mEventTypes, mEvents, nClients int, bucketDate time.Time) {
	storePrefix := "store"
	clientPrefix := "client"
	eventTypes := make([]string, mEventTypes)
	for i := 0; i < mEventTypes; i++ {
		eventTypes[i] = fmt.Sprintf("event_type%d", i+1)
	}

	var wg sync.WaitGroup

	numEvents := rand.Intn(mEvents) + 1
	fmt.Println("Count total events: ", numEvents)
	indexStore := rand.Intn(nStores) + 1
	storeID := fmt.Sprintf("%s%d", storePrefix, indexStore)
	indexClient := rand.Intn(nClients) + 1
	clientID := fmt.Sprintf("%s%d", clientPrefix, indexClient)
	eventType := eventTypes[rand.Intn(mEventTypes)]
	fmt.Println("Store ID: ", storeID)
	fmt.Println("Client ID: ", clientID)
	fmt.Println("Bucket Date: ", bucketDate)
	fmt.Println("Event Type: ", eventType)

	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			eventID := fmt.Sprintf("evt%d", i+1)
			timestamp := bucketDate.Add(time.Duration(rand.Intn(24)) * time.Hour).Add(time.Duration(rand.Intn(60)) * time.Minute)
			status := []string{"success", "failed"}[rand.Intn(2)]
			event := Event{
				ID:        eventID,
				TimeStamp: timestamp.Unix(),
				Status:    status,
			}
			trackingEvent := TrackingEvent{
				StoreId:    storeID,
				UserId:     clientID,
				BucketDate: bucketDate.UnixNano(),
				EventType:  eventType,
				Count:      1,
				Event:      event,
			}
			fmt.Println(trackingEvent)

		}(i)
	}
	wg.Wait()

	fmt.Println("===========================Message produced successfully!=============================")
}

func main() {
	nStores := 10
	mEventTypes := 10
	mEvents := 100
	nClients := 10

	generateMockDataV2(nStores, mEventTypes, mEvents, nClients, time.Now())

}
