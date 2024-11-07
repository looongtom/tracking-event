package main

import (
	"fmt"
	"math/rand"
	"receive-event/model"
	"sync"
	"time"
)

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

	for i := 0; i < numEvents; i++ {
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

func generateMockData(nStores, mEventTypes, mEvents, nClients int, bucketDate time.Time) error {

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

			detailEvent := model.EventDetails{
				EventID:   eventID,
				Timestamp: timestamp.Unix(),
				EventType: eventType,
			}

			sentEvent := model.EventRecordRequestV3{
				ClientID:    clientID,
				StoreID:     storeID,
				BucketDate:  bucketDate.Format("02-01-2006"),
				EventDetail: detailEvent,
			}
			fmt.Println(sentEvent)
			//serializedBookingRequest, err := json.Marshal(sentEvent)
			//if err != nil {
			//	//http.Error(w, fmt.Sprintf("Failed to serialize booking request: %s", err), http.StatusInternalServerError)
			//	errChan <- fmt.Errorf("Failed to serialize booking request: %s", err)
			//	return
			//}
			//
			//// Produce the message to the Kafka topic
			//err = produceMessage(p, topic, serializedBookingRequest)
			//if err != nil {
			//	//http.Error(w, fmt.Sprintf("Failed to produce message: %s", err), http.StatusInternalServerError)
			//	errChan <- fmt.Errorf("Failed to produce message: %s", err)
			//	return
			//}

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

func main() {
	nStores := 10
	mEventTypes := 5
	mEvents := 100
	nClients := 10

	err := generateMockData(nStores, mEventTypes, mEvents, nClients, time.Now())
	if err != nil {
		return
	}
	fmt.Println("Done")

}
