package controllers

import (
	"bufio"
	"fmt"
	"time"

	"warkop/internal/middleware"
	"warkop/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

type EventController struct {
	hub *services.EventHub
}

func NewEventController(hub *services.EventHub) *EventController {
	return &EventController{hub: hub}
}

func (h *EventController) Stream(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("X-Accel-Buffering", "no")
	ch := h.hub.Subscribe(middleware.CurrentUser(c).KlienID)
	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		defer h.hub.Unsubscribe(ch)
		fmt.Fprint(w, ": ok\n\n")
		if w.Flush() != nil {
			return
		}
		tick := time.NewTicker(25 * time.Second)
		defer tick.Stop()
		for {
			select {
			case ev := <-ch:
				fmt.Fprintf(w, "event: %s\ndata: 1\n\n", ev)
			case <-tick.C:
				fmt.Fprint(w, ": ping\n\n") // flush gagal di sini berarti klien sudah putus
			}
			if w.Flush() != nil {
				return
			}
		}
	}))
	return nil
}
