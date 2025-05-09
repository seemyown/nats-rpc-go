package natsrpc

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

// Ctx carries request and response information through the middleware chain.
type Ctx struct {
	context.Context           // Embedded standard context for cancellation / timeout
	Msg             *nats.Msg // Original NATS message
	conn            *nats.Conn
}

// Bind unmarshals the JSON body of the request into v.
func (c *Ctx) Bind(v any) error {
	if len(c.Msg.Data) == 0 {
		return ErrBadRequest
	}
	return json.Unmarshal(c.Msg.Data, v)
}

// JSON sends v serialized as JSON back to the requester.
func (c *Ctx) JSON(v any, hdr nats.Header) error {
	if c.Msg.Reply == "" {
		return ErrBadRequest
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	msg := &nats.Msg{
		Subject: c.Msg.Reply,
		Data:    data,
		Header:  hdr,
	}
	return c.conn.PublishMsg(msg)
}

// Header returns request headers.
func (c *Ctx) Header() nats.Header {
	return c.Msg.Header
}
