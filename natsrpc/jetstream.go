package natsrpc

import (
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
)

// SubscribeJSQueue создаёт подписку JetStream с ручным подтверждением.
// js         — JetStreamContext,
// subject    — subject для подписки,
// queueGroup — имя группы (пустая строка = без queue),
// durable    — имя durable-консьюмера (пусто = без durable).
func SubscribeJSQueue(js nats.JetStreamContext, subject, queueGroup, durable string, cb nats.MsgHandler) (*nats.Subscription, error) {
	opts := []nats.SubOpt{nats.ManualAck()}
	if durable != "" {
		opts = append(opts, nats.Durable(durable))
	}
	if queueGroup != "" {
		return js.QueueSubscribe(subject, queueGroup, cb, opts...)
	}
	return js.Subscribe(subject, cb, opts...)
}

// EnsureStream делает стрим с именем streamName и subjectPattern, если он ещё не существует.
// Это удобно вызывать при инициализации сервиса.
func EnsureStream(js nats.JetStreamContext, streamName, subjectPattern string) error {
	_, err := js.StreamInfo(streamName)
	if err == nil {
		return nil // уже есть
	}
	if !errors.Is(err, nats.ErrStreamNotFound) {
		return fmt.Errorf("natsrpc: checking stream info: %w", err)
	}
	// создаём новый стрим
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{subjectPattern},
	})
	if err != nil {
		return fmt.Errorf("natsrpc: creating stream: %w", err)
	}
	return nil
}
