package notification_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/datakeys/kyc-service/internal/notification"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewSSEHub_SubscribeAndPublish(t *testing.T) {
	hub := notification.NewSSEHub()
	ch := hub.Subscribe("+221771234567")

	assert.Equal(t, 1, hub.SubscriberCount("+221771234567"))

	notif := notification.Notification{
		Event:          notification.EventKYCApproved,
		VerificationID: "v-12345",
		Phone:          "+221771234567",
		Status:         "approved",
		Score:          0.95,
		CountryCode:    "SN",
		Timestamp:      time.Now().UTC(),
	}

	hub.Publish("+221771234567", notif)

	select {
	case received := <-ch:
		assert.Equal(t, notif.Event, received.Event)
		assert.Equal(t, notif.VerificationID, received.VerificationID)
		assert.Equal(t, notif.Status, received.Status)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for notification")
	}
}

func TestSSEHub_Unsubscribe(t *testing.T) {
	hub := notification.NewSSEHub()
	ch := hub.Subscribe("+221771234567")

	hub.Unsubscribe("+221771234567", ch)

	assert.Equal(t, 0, hub.SubscriberCount("+221771234567"))

	_, ok := <-ch
	assert.False(t, ok, "channel should be closed after unsubscribe")
}

func TestSSEHub_NoSubscriberDrops(t *testing.T) {
	hub := notification.NewSSEHub()
	notif := notification.Notification{
		Event:  notification.EventKYCInitiated,
		Phone:  "+221771234567",
		Status: "pending",
	}

	assert.NotPanics(t, func() {
		hub.Publish("+221771234567", notif)
	})
}

func TestSSEHub_MultipleSubscribers(t *testing.T) {
	hub := notification.NewSSEHub()
	ch1 := hub.Subscribe("+221771234567")
	ch2 := hub.Subscribe("+221771234567")

	assert.Equal(t, 2, hub.SubscriberCount("+221771234567"))

	notif := notification.Notification{
		Event:  notification.EventKYCApproved,
		Phone:  "+221771234567",
		Status: "approved",
	}

	hub.Publish("+221771234567", notif)

	for i, ch := range []chan notification.Notification{ch1, ch2} {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d did not receive notification", i)
		}
	}
}

func TestSSEHub_DifferentPhoneIsolated(t *testing.T) {
	hub := notification.NewSSEHub()
	ch := hub.Subscribe("+221771234567")

	hub.Publish("+33612345678", notification.Notification{
		Event:  notification.EventKYCApproved,
		Phone:  "+33612345678",
		Status: "approved",
	})

	select {
	case <-ch:
		t.Fatal("should not receive notification for different phone")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestNewNotifier_Notify(t *testing.T) {
	hub := notification.NewSSEHub()
	ch := hub.Subscribe("+221771234567")

	notifier := notification.NewNotifier(hub)

	notif := notification.Notification{
		Event:          notification.EventKYCApproved,
		VerificationID: "v-test",
		Phone:          "+221771234567",
		Status:         "approved",
		CountryCode:    "SN",
		Timestamp:      time.Now().UTC(),
	}

	notifier.Notify(context.Background(), notif)

	select {
	case received := <-ch:
		assert.Equal(t, "v-test", received.VerificationID)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestConsoleSMS_Send(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sms := notification.NewConsoleSMS(logger)

	notif := notification.Notification{
		Event:          notification.EventKYCInitiated,
		VerificationID: "v-sms-test",
		Phone:          "+221771234567",
		CountryCode:    "SN",
		Status:         "pending",
		Timestamp:      time.Now().UTC(),
	}

	err := sms.Send(context.Background(), notif)
	assert.NoError(t, err)
	assert.True(t, sms.IsAvailable())
	assert.Equal(t, "console_sms", sms.Name())
}

func TestConsoleSMS_FormatApproved(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sms := notification.NewConsoleSMS(logger)

	notif := notification.Notification{
		Event:  notification.EventKYCApproved,
		Phone:  "+221771234567",
		Status: "approved",
		Score:  0.85,
	}

	err := sms.Send(context.Background(), notif)
	assert.NoError(t, err)
}

func TestConsoleSMS_FormatRejected(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sms := notification.NewConsoleSMS(logger)

	notif := notification.Notification{
		Event:  notification.EventKYCRejected,
		Phone:  "+221771234567",
		Status: "rejected",
	}

	err := sms.Send(context.Background(), notif)
	assert.NoError(t, err)
}

func TestNotifyAsync_NoBlock(t *testing.T) {
	hub := notification.NewSSEHub()
	notifier := notification.NewNotifier(hub)

	notif := notification.Notification{
		Event:  notification.EventKYCInitiated,
		Phone:  "+221771234567",
		Status: "pending",
	}

	assert.NotPanics(t, func() {
		notifier.NotifyAsync(notif)
	})

	time.Sleep(100 * time.Millisecond)
}

func TestNewNotifier_WithSMSChannel(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	sms := notification.NewConsoleSMS(logger)
	hub := notification.NewSSEHub()

	notifier := notification.NewNotifier(hub, sms)

	notif := notification.Notification{
		Event:          notification.EventKYCApproved,
		VerificationID: "v-channel-test",
		Phone:          "+221771234567",
		Status:         "approved",
		CountryCode:    "SN",
	}

	assert.NotPanics(t, func() {
		notifier.Notify(context.Background(), notif)
	})
}

func TestMarshalJSON(t *testing.T) {
	notif := notification.Notification{
		Event:          notification.EventKYCInitiated,
		VerificationID: "v-json-test",
		CountryCode:    "BF",
		Status:         "pending",
		Timestamp:      time.Now().UTC(),
		Metadata: map[string]interface{}{
			"provider": "smileid",
		},
	}

	data := notification.MarshalJSON(notif)
	assert.True(t, json.Valid(data))

	var decoded notification.Notification
	err := json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, notif.Event, decoded.Event)
	assert.Equal(t, notif.VerificationID, decoded.VerificationID)
}

func TestHub_SubscriberCount_Empty(t *testing.T) {
	hub := notification.NewSSEHub()
	assert.Equal(t, 0, hub.SubscriberCount("+221771234567"))
}

func TestNotifier_SSEHub(t *testing.T) {
	hub := notification.NewSSEHub()
	notifier := notification.NewNotifier(hub)
	assert.Same(t, hub, notifier.SSEHub())
}
