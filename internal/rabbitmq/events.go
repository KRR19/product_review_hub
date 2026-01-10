package rabbitmq

import "time"

// EventType represents the type of review event.
type EventType string

const (
	EventReviewCreated EventType = "review.created"
	EventReviewUpdated EventType = "review.updated"
	EventReviewDeleted EventType = "review.deleted"
)

// ReviewEventData contains the data for a review event.
type ReviewEventData struct {
	ReviewID  string `json:"review_id"`
	ProductID string `json:"product_id"`
	Rating    int    `json:"rating,omitempty"`
}

// ReviewEvent represents an event that is published when a review is created, updated, or deleted.
type ReviewEvent struct {
	EventType EventType       `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      ReviewEventData `json:"data"`
}

// NewReviewEvent creates a new ReviewEvent with the current timestamp.
func NewReviewEvent(eventType EventType, reviewID, productID string, rating int) ReviewEvent {
	return ReviewEvent{
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Data: ReviewEventData{
			ReviewID:  reviewID,
			ProductID: productID,
			Rating:    rating,
		},
	}
}
