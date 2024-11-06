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
