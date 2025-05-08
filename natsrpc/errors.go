package natsrpc

import "errors"

// Ошибки, которые могут пригодиться для логики и валидации опций.
var (
	ErrRouterStarted            = errors.New("natsrpc: router already started, cannot add new routes")
	ErrHandlerAlreadyRegistered = errors.New("natsrpc: handler already registered for this subject")
	ErrNoReplySubject           = errors.New("natsrpc: no reply subject specified")
	ErrInvalidOption            = errors.New("natsrpc: invalid router option")
	ErrNilContext               = errors.New("natsrpc: context must not be nil")
)
