package natsrpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/nats-io/nats.go"
)

// publisher — минимальный интерфейс для отправки сообщений.
type publisher interface {
	Publish(subject string, data []byte) error
	PublishMsg(msg *nats.Msg) error
}

// Router управляет RPC-подписками и middleware.
type Router struct {
	nc             publisher
	js             nats.JetStreamContext
	durablePrefix  string
	defaultCtx     context.Context
	requestTimeout time.Duration
	middlewares    []MiddlewareFunc
	subs           []*nats.Subscription
}

// Option — функциональный параметр для NewRouter.
type Option func(*Router) error

// NewRouter создаёт Router на базе *nats.Conn и опций.
func NewRouter(nc *nats.Conn, opts ...Option) (*Router, error) {
	if nc == nil {
		return nil, errors.New("nats connection cannot be nil")
	}
	r := &Router{
		nc:             nc,
		defaultCtx:     context.Background(),
		requestTimeout: 0,
		middlewares:    []MiddlewareFunc{},
		subs:           []*nats.Subscription{},
	}
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("applying option: %w", err)
		}
	}
	return r, nil
}

// WithContext задаёт базовый context.Context для всех запросов.
func WithContext(ctx context.Context) Option {
	return func(r *Router) error {
		if ctx == nil {
			return errors.New("context cannot be nil")
		}
		r.defaultCtx = ctx
		return nil
	}
}

// WithRequestTimeout задаёт дедлайн для каждого запроса.
func WithRequestTimeout(d time.Duration) Option {
	return func(r *Router) error {
		if d < 0 {
			return errors.New("timeout must be non-negative")
		}
		r.requestTimeout = d
		return nil
	}
}

// WithJetStream включает JetStream (durablePrefix необязателен).
func WithJetStream(js nats.JetStreamContext, durablePrefix string) Option {
	return func(r *Router) error {
		if js == nil {
			return errors.New("jetstream context cannot be nil")
		}
		r.js = js
		r.durablePrefix = durablePrefix
		return nil
	}
}

// Use добавляет глобальные middleware (выполняются первыми).
func (r *Router) Use(mws ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, mws...)
}

// Handle регистрирует обработчик на subject с опциональными route-middleware.
func (r *Router) Handle(subject string, handler HandlerFunc, routeMws ...MiddlewareFunc) error {
	if subject == "" {
		return errors.New("subject cannot be empty")
	}
	// Собираем цепочку middleware
	all := append([]MiddlewareFunc{}, r.middlewares...)
	all = append(all, routeMws...)
	final := handler
	for i := len(all) - 1; i >= 0; i-- {
		final = all[i](final)
	}
	// Callback для каждой заявки
	callback := func(msg *nats.Msg) {
		// Panic recovery
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[natsrpc] panic on %s: %v\n%s", msg.Subject, rec, debug.Stack())
			}
		}()
		// Сформировать контекст с таймаутом
		ctx, cancel := func() (context.Context, context.CancelFunc) {
			if r.requestTimeout > 0 {
				return context.WithTimeout(r.defaultCtx, r.requestTimeout)
			}
			return context.WithCancel(r.defaultCtx)
		}()
		defer cancel()
		// Создаём наш Context и запускаем цепочку
		c := &Context{
			Context:   ctx,
			Msg:       msg,
			router:    r,
			outHeader: make(nats.Header),
		}
		final(c)
		// Аcknowledge для JetStream
		if r.js != nil {
			if err := msg.Ack(); err != nil {
				log.Printf("[natsrpc] ack error on %s: %v", msg.Subject, err)
			}
		}
	}
	// Подписываемся либо через JetStream, либо обычный Subscribe
	var sub *nats.Subscription
	var err error
	if r.js != nil {
		dur := r.durablePrefix
		if dur != "" {
			dur = fmt.Sprintf("%s_%s", dur, subject)
		}
		opts := []nats.SubOpt{nats.ManualAck()}
		if dur != "" {
			opts = append(opts, nats.Durable(dur))
		}
		sub, err = r.js.Subscribe(subject, callback, opts...)
	} else {
		conn := r.nc.(*nats.Conn)
		sub, err = conn.Subscribe(subject, callback)
	}
	if err != nil {
		return fmt.Errorf("subscribe %s: %w", subject, err)
	}
	r.subs = append(r.subs, sub)
	log.Printf("[natsrpc] subscribed to %s", subject)
	return nil
}

// Close отписывает все подписки (не закрывая сам соединение).
func (r *Router) Close() error {
	var firstErr error
	for _, s := range r.subs {
		if s == nil {
			continue
		}
		if err := s.Unsubscribe(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	r.subs = nil
	return firstErr
}
