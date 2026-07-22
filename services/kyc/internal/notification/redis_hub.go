package notification

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const redisChannelPrefix = "kyc:sse:"

type RedisHub struct {
	local  *SSEHub
	rdb    *redis.Client
	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewRedisHub(rdb *redis.Client, logger *zap.Logger) *RedisHub {
	ctx, cancel := context.WithCancel(context.Background())
	return &RedisHub{
		local:  NewSSEHub(),
		rdb:    rdb,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (h *RedisHub) Subscribe(phone string) chan Notification {
	return h.local.Subscribe(phone)
}

func (h *RedisHub) Unsubscribe(phone string, ch chan Notification) {
	h.local.Unsubscribe(phone, ch)
}

func (h *RedisHub) Publish(phone string, notif Notification) {
	channel := redisChannelPrefix + phone
	data, err := json.Marshal(notif)
	if err != nil {
		h.logger.Error("redis hub marshal", zap.Error(err))
		return
	}
	if err := h.rdb.Publish(h.ctx, channel, data).Err(); err != nil {
		h.logger.Error("redis hub publish", zap.String("channel", channel), zap.Error(err))
	}
}

func (h *RedisHub) SubscriberCount(phone string) int {
	return h.local.SubscriberCount(phone)
}

func (h *RedisHub) Start() {
	pubsub := h.rdb.PSubscribe(h.ctx, redisChannelPrefix+"*")
	_, err := pubsub.Receive(h.ctx)
	if err != nil {
		h.logger.Error("redis hub subscribe failed", zap.Error(err))
		pubsub.Close()
		return
	}
	h.wg.Add(1)
	go h.consumeLoop(pubsub)
}

func (h *RedisHub) Stop() {
	h.cancel()
	h.wg.Wait()
}

func (h *RedisHub) consumeLoop(pubsub *redis.PubSub) {
	defer h.wg.Done()
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-h.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			h.handleMessage(msg)
		}
	}
}

func (h *RedisHub) handleMessage(msg *redis.Message) {
	phone := strings.TrimPrefix(msg.Channel, redisChannelPrefix)
	if phone == "" {
		return
	}

	var notif Notification
	if err := json.Unmarshal([]byte(msg.Payload), &notif); err != nil {
		h.logger.Error("redis hub unmarshal", zap.Error(err))
		return
	}

	h.local.Publish(phone, notif)
}

func phoneFromChannel(channel string) string {
	return strings.TrimPrefix(channel, redisChannelPrefix)
}
