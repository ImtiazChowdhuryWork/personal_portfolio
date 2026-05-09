// ============================================================
// FILE: internal/handlers/sse_handler.go
// WHAT IT IS:     Server-Sent Events stream for the dashboard
// ENDPOINT:       GET /api/v1/events?token=<JWT>   [protected via query]
// WHY IT EXISTS:  Browser EventSource cannot send custom headers, so the
//                 JWT is supplied as a query param instead of an
//                 Authorization header. The endpoint validates the token
//                 itself rather than going through AuthMiddleware.
// LAST UPDATED:   2026-05-10 — initial creation
// ============================================================

package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"imtiaz-portfolio/config"
	"imtiaz-portfolio/internal/services"
	"imtiaz-portfolio/internal/utils"

	"github.com/gin-gonic/gin"
)

// SSEHandler holds the broker the dashboard subscribes to and the JWT
// secret needed to validate the query-string token.
type SSEHandler struct {
	broker *services.EventBroker
	cfg    *config.Config
}

func NewSSEHandler(b *services.EventBroker, cfg *config.Config) *SSEHandler {
	return &SSEHandler{broker: b, cfg: cfg}
}

// Stream holds an SSE connection open and pushes events as they arrive.
// Heartbeats every 25s keep proxies and Cloudflare-style intermediaries
// from killing the idle connection. The handler exits when the client
// disconnects (browser tab close, navigation away, network drop).
func (h *SSEHandler) Stream(c *gin.Context) {
	// Validate the JWT from the query string (EventSource can't set headers).
	token := c.Query("token")
	if token == "" {
		utils.Unauthorized(c, "token query parameter is required")
		return
	}
	if _, err := utils.ValidateToken(token, h.cfg.JWTSecret); err != nil {
		utils.Unauthorized(c, "invalid or expired token")
		return
	}

	// SSE headers — text/event-stream + disable buffering at the proxy layer
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-transform")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // hint for nginx

	flusher, ok := c.Writer.(interface{ Flush() })
	if !ok {
		// Server doesn't support flushing — SSE won't work
		c.String(500, "streaming not supported")
		return
	}

	events, unsubscribe := h.broker.Subscribe()
	defer unsubscribe()

	// Send a hello event so the dashboard knows the stream is live.
	fmt.Fprintf(c.Writer, "event: ready\ndata: {\"ok\":true}\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	clientGone := c.Request.Context().Done()

	for {
		select {
		case <-clientGone:
			return
		case <-heartbeat.C:
			// Comment line — clients ignore but it keeps the TCP connection
			// from going idle long enough to be cut by intermediaries.
			fmt.Fprint(c.Writer, ": heartbeat\n\n")
			flusher.Flush()
		case evt, alive := <-events:
			if !alive {
				return
			}
			payload, err := json.Marshal(evt.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", evt.Type, payload)
			flusher.Flush()
		}
	}
}
