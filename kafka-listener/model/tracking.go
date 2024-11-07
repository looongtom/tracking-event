package model

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

type EventRecord struct {
	ClientID          string `json:"client_id"`
	StoreID           string `json:"store_id"`
	EventType         string `json:"event_type"`
	StatusDestination string `json:"status_destination"`
	EventID           string `json:"event_id"`
	Timestamp         int64  `json:"timestamp"`
	BucketDate        string `json:"bucket_date"`
}

type EventRecordV3 struct {
	ClientID    string         `json:"client_id"`
	StoreID     string         `json:"store_id"`
	BucketDate  string         `json:"bucket_date"`
	ListSuccess []EventDetails `json:"list_success"`
	ListFailure []EventDetails `json:"list_failure"`
}
type EventRecordRequestV3 struct {
	ClientID    string       `json:"client_id"`
	StoreID     string       `json:"store_id"`
	BucketDate  string       `json:"bucket_date"`
	Status      string       `json:"status"`
	EventDetail EventDetails `json:"event"`
}

type EventDetails struct {
	EventID   string `json:"event_id"`
	Timestamp int64  `json:"timestamp"`
	EventType string `json:"event_type"`
}
