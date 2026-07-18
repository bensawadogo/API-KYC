package notification

import (
	"context"
	"encoding/json"
	"time"
)

type Event string

const (
	EventKYCInitiated    Event = "kyc.initiated"
	EventKYCApproved     Event = "kyc.approved"
	EventKYCRejected     Event = "kyc.rejected"
	EventKYCPendingReview Event = "kyc.pending_review"
	EventKYCError        Event = "kyc.error"
)

type Notification struct {
	Event          Event                  `json:"event"`
	VerificationID string                 `json:"verification_id"`
	Phone          string                 `json:"-"`
	CountryCode    string                 `json:"country_code"`
	Status         string                 `json:"status"`
	Score          float64                `json:"score,omitempty"`
	Message        string                 `json:"message,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type Channel interface {
	Name() string
	Send(ctx context.Context, notif Notification) error
}

type Notifier struct {
	channels []Channel
	hub      *SSEHub
}

func NewNotifier(hub *SSEHub, channels ...Channel) *Notifier {
	return &Notifier{hub: hub, channels: channels}
}

func (n *Notifier) Notify(ctx context.Context, notif Notification) {
	if n.hub != nil {
		n.hub.Publish(notif.Phone, notif)
	}
	for _, ch := range n.channels {
		if err := ch.Send(ctx, notif); err != nil {
			continue
		}
	}
}

func (n *Notifier) NotifyAsync(notif Notification) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		n.Notify(ctx, notif)
	}()
}

func (n *Notifier) SSEHub() *SSEHub { return n.hub }

func MarshalJSON(notif Notification) []byte {
	data, err := json.Marshal(notif)
	if err != nil {
		return []byte(`{"error":"marshal failed"}`)
	}
	return data
}
