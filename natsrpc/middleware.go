package natsrpc

import (
	"log"
	"time"
)

// Logger logs execution time and errors for each RPC call.
func Logger() Middleware {
	return func(next Handler) Handler {
		return func(c *Ctx) error {
			start := time.Now()
			err := next(c)
			log.Printf("[natsrpc] subject=%s time=%s err=%v", c.Msg.Subject, time.Since(start), err)
			return err
		}
	}
}
