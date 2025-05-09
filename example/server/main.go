package main

import (
	"github.com/nats-io/nats.go"
	"github.com/seemyown/nats-rpc-go/natsrpc"
	"log"
)

type HelloRequest struct {
	Name string `json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	app := natsrpc.New(nc)
	app.Use(natsrpc.Logger())

	app.Handle("hello", func(c *natsrpc.Ctx) error {
		var req HelloRequest
		if err := c.Bind(&req); err != nil {
			return natsrpc.ErrBadRequest
		}
		resp := HelloResponse{Message: "Hello, " + req.Name}
		return c.JSON(resp, nil)
	})

	if err := app.Listen(); err != nil {
		log.Fatal(err)
	}
	log.Println("server listening...")

	// Keep process alive
	select {}
}
