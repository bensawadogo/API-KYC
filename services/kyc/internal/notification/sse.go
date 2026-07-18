package notification

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type SSEHandler struct {
	hub    *SSEHub
	logger *zap.Logger
}

func NewSSEHandler(hub *SSEHub, logger *zap.Logger) *SSEHandler {
	return &SSEHandler{hub: hub, logger: logger}
}

func (h *SSEHandler) Stream(c fiber.Ctx) error {
	phone := c.Query("phone")
	if phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "paramètre 'phone' requis (format E.164, ex: +221771234567)",
		})
	}

	ch := h.hub.Subscribe(phone)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	c.RequestCtx().Hijack(func(conn net.Conn) {
		defer conn.Close()
		defer h.hub.Unsubscribe(phone, ch)

		initData, _ := json.Marshal(map[string]string{
			"status": "connected",
			"phone":  phone,
		})
		fmt.Fprintf(conn, "event: connected\ndata: %s\n\n", initData)

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case notif, ok := <-ch:
				if !ok {
					return
				}
				data, err := json.Marshal(notif)
				if err != nil {
					continue
				}
				_, err = fmt.Fprintf(conn, "event: %s\ndata: %s\n\n", notif.Event, data)
				if err != nil {
					return
				}
			case <-ticker.C:
				_, err := fmt.Fprintf(conn, ": keepalive\n\n")
				if err != nil {
					return
				}
			}
		}
	})

	return nil
}
