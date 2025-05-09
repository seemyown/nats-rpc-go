# natsrpc

Tiny RPC framework on top of [NATS](https://nats.io) with an API inspired by [Fiber](https://github.com/gofiber/fiber).

## Install

```bash
go get github.com/seemyown/nats-rpc-go/natsrpc
```

## Quick Start

### Server

```go
package main

import (
    "log"
    "github.com/nats-io/nats.go"
    "github.com/seemyown/nats-rpc-go/natsrpc"
)

type HelloRequest struct {
    Name string `json:"name"`
}

type HelloResponse struct {
    Message string `json:"message"`
}

func main() {
    nc, _ := nats.Connect(nats.DefaultURL)
    app := natsrpc.New(nc)
    app.Use(natsrpc.Logger())

    app.Handle("hello", func(c *natsrpc.Ctx) error {
        var req HelloRequest
        if err := c.Bind(&req); err != nil {
            return natsrpc.ErrBadRequest
        }
        return c.JSON(HelloResponse{Message: "Hello " + req.Name}, nil)
    })

    _ = app.Listen()
    select {}
}
```

### Client

```go
package main

import (
    "encoding/json"
    "fmt"
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
    nc, _ := nats.Connect(nats.DefaultURL)
    defer nc.Drain()

    req := HelloRequest{Name: "Gilfoyle"}
    data, _ := json.Marshal(req)

    msg, _ := nc.Request("hello", data, 2*time.Second)

    var resp HelloResponse
    _ = json.Unmarshal(msg.Data, &resp)
    fmt.Println(resp.Message)
}
```

## Middleware

You can register global or local middleware:

```go
app.Use(natsrpc.Logger())            // global
app.Handle("foo", handler, auth())   // local
```

A middleware has the signature:

```go
type Middleware func(natsrpc.Handler) natsrpc.Handler
```

Wrap the next handler to add behaviour.

## Error Handling

Return predefined errors (`ErrBadRequest`, `ErrInternal`, …) or create your own `natsrpc.RPCError{Code, Msg}`.

## License

MIT
