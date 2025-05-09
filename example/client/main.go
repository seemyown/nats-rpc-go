package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
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
	defer nc.Drain()

	req := HelloRequest{Name: "Gilfoyle"}
	data, _ := json.Marshal(req)

	msg, err := nc.Request("hello", data, 2*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	var resp HelloResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Message)
}
