package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// AddRequest/Response должны совпадать с сервером
type AddRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}
type AddResponse struct {
	Result int `json:"result"`
}

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	req := AddRequest{A: 3, B: 4}
	data, _ := json.Marshal(req)
	msg, err := nc.Request("calc.add", data, 1*time.Second)
	if err != nil {
		log.Fatal("request error:", err)
	}
	var resp AddResponse
	json.Unmarshal(msg.Data, &resp)
	fmt.Printf("%d + %d = %d\n", req.A, req.B, resp.Result)
}
