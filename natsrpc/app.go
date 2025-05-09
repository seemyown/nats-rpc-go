package natsrpc

import (
	"context"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Handler processes incoming RPC requests.
type Handler func(*Ctx) error

// Middleware transforms or wraps a Handler.
type Middleware func(Handler) Handler

// App is the central RPC application.
type App struct {
	nc          *nats.Conn
	handlers    map[string]Handler
	middlewares []Middleware
}

// New creates a new App bound to a NATS connection.
func New(nc *nats.Conn) *App {
	return &App{
		nc:       nc,
		handlers: make(map[string]Handler),
	}
}

// Use registers global middleware.
func (a *App) Use(m Middleware) {
	a.middlewares = append(a.middlewares, m)
}

// Handle registers a handler for a specific subject.
func (a *App) Handle(subject string, h Handler, m ...Middleware) {
	// Apply local middleware first, then global, to mimic Fiber order (appUse -> route)
	chain := h
	// Local middleware (closest to handler)
	for i := len(m) - 1; i >= 0; i-- {
		chain = m[i](chain)
	}
	// Global middleware
	for i := len(a.middlewares) - 1; i >= 0; i-- {
		chain = a.middlewares[i](chain)
	}
	a.handlers[subject] = chain
}

// Listen subscribes to all registered subjects and starts processing.
func (a *App) Listen() error {
	for subj, h := range a.handlers {
		_, err := a.nc.QueueSubscribe(subj, "natsrpc", func(msg *nats.Msg) {
			// Derive context with cancellation tied to client timeout
			ctx, cancel := context.WithCancel(context.Background())
			// Ensure cancel on function exit
			defer cancel()

			// NATS doesn't propagate cancellation from client; we rely on server timeout if set
			go func() {
				// Auto-cancel after reasonable time to avoid goroutine leaks
				<-time.After(30 * time.Second)
				cancel()
			}()

			c := &Ctx{
				Context: ctx,
				Msg:     msg,
				conn:    a.nc,
			}
			err := h(c)
			if err != nil {
				// Log error and optionally send error back
				if msg.Reply != "" {
					_ = c.JSON(err, nil)
				}
				log.Printf("handler error on %s: %v", subj, err)
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}
