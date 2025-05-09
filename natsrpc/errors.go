package natsrpc

import "fmt"

// RPCError represents an error with a numeric code and message, similar to HTTP status codes.
type RPCError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Msg)
}

// Predefined common errors
var (
	ErrBadRequest = &RPCError{Code: 400, Msg: "bad request"}
	ErrNotFound   = &RPCError{Code: 404, Msg: "not found"}
	ErrTimeout    = &RPCError{Code: 408, Msg: "timeout"}
	ErrInternal   = &RPCError{Code: 500, Msg: "internal server error"}
)
