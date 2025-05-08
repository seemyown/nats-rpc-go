package main

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/seemyown/nats-rpc-go/natsrpc"
)

// AddRequest — пример структуры запроса
type AddRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

// AddResponse — пример структуры ответа
type AddResponse struct {
	Result int `json:"result"`
}

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Новый роутер с таймаутом 1s
	router, err := natsrpc.NewRouter(nc, natsrpc.WithRequestTimeout(time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer router.Close()

	// Глобальный логгер
	router.Use(func(next natsrpc.HandlerFunc) natsrpc.HandlerFunc {
		return func(c *natsrpc.Context) {
			log.Println("got request on", c.Msg.Subject)
			next(c)
		}
	})

	router.Handle("calc.add", func(c *natsrpc.Context) {
		var req AddRequest
		if err := c.Bind(&req); err != nil {
			log.Println("bind error:", err)
			return
		}
		res := AddResponse{Result: req.A + req.B}
		c.JSON(res)
	})

	log.Println("listening on calc.add")
	select {}
}
