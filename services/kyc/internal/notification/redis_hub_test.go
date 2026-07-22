package notification_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/datakeys/kyc-service/internal/notification"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRedisHub_SubscribeAndPublish(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger, _ := zap.NewDevelopment()

	hub := notification.NewRedisHub(rdb, logger)
	hub.Start()
	defer hub.Stop()

	ch := hub.Subscribe("+221771234567")
	assert.Equal(t, 1, hub.SubscriberCount("+221771234567"))

	notif := notification.Notification{
		Event:          notification.EventKYCApproved,
		VerificationID: "v-redis-1",
		Phone:          "+221771234567",
		Status:         "approved",
		CountryCode:    "SN",
		Timestamp:      time.Now().UTC(),
	}

	hub.Publish("+221771234567", notif)

	select {
	case received := <-ch:
		assert.Equal(t, notif.Event, received.Event)
		assert.Equal(t, notif.VerificationID, received.VerificationID)
		assert.Equal(t, notif.Status, received.Status)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for notification via redis hub")
	}
}

func TestRedisHub_MultipleSubscribers(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger, _ := zap.NewDevelopment()

	hub := notification.NewRedisHub(rdb, logger)
	hub.Start()
	defer hub.Stop()

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
		case <-time.After(2 * time.Second):
			t.Fatalf("subscriber %d did not receive notification", i)
		}
	}
}

func TestRedisHub_DifferentPhoneIsolated(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger, _ := zap.NewDevelopment()

	hub := notification.NewRedisHub(rdb, logger)
	hub.Start()
	defer hub.Stop()

	ch := hub.Subscribe("+221771234567")

	hub.Publish("+33612345678", notification.Notification{
		Event:  notification.EventKYCApproved,
		Phone:  "+33612345678",
		Status: "approved",
	})

	select {
	case <-ch:
		t.Fatal("should not receive notification for different phone")
	case <-time.After(500 * time.Millisecond):
	}
}

func TestRedisHub_CrossInstance(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb1 := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rdb2 := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger, _ := zap.NewDevelopment()

	hub1 := notification.NewRedisHub(rdb1, logger)
	hub1.Start()
	defer hub1.Stop()

	hub2 := notification.NewRedisHub(rdb2, logger)
	hub2.Start()
	defer hub2.Stop()

	ch := hub2.Subscribe("+221771234567")

	hub1.Publish("+221771234567", notification.Notification{
		Event:  notification.EventKYCInitiated,
		Phone:  "+221771234567",
		Status: "pending",
	})

	select {
	case received := <-ch:
		assert.Equal(t, notification.EventKYCInitiated, received.Event)
		assert.Equal(t, "pending", received.Status)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout: hub2 did not receive message published by hub1")
	}
}

func TestRedisHub_Unsubscribe(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger, _ := zap.NewDevelopment()

	hub := notification.NewRedisHub(rdb, logger)
	hub.Start()
	defer hub.Stop()

	ch := hub.Subscribe("+221771234567")
	hub.Unsubscribe("+221771234567", ch)
	assert.Equal(t, 0, hub.SubscriberCount("+221771234567"))

	_, ok := <-ch
	assert.False(t, ok, "channel should be closed after unsubscribe")
}

func TestRedisHub_NoSubscriberDrops(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger, _ := zap.NewDevelopment()

	hub := notification.NewRedisHub(rdb, logger)
	hub.Start()
	defer hub.Stop()

	assert.NotPanics(t, func() {
		hub.Publish("+221771234567", notification.Notification{
			Event:  notification.EventKYCInitiated,
			Phone:  "+221771234567",
			Status: "pending",
		})
	})
}
