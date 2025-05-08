package natsrpc

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

// Context представляет контекст одного RPC-запроса.
// Встраивает context. Context для таймаутов и отмены.
type Context struct {
	context.Context             // базовый контекст
	Msg             *nats.Msg   // оригинальное NATS-сообщение
	router          *Router     // ссылка на роутер
	outHeader       nats.Header // заголовки для ответа
}

// Bind разбирает JSON из тела запроса в dest.
// Если тело пустое — ничего не делает.
func (c *Context) Bind(dest interface{}) error {
	if len(c.Msg.Data) == 0 {
		return nil
	}
	return json.Unmarshal(c.Msg.Data, dest)
}

// JSON сериализует response в JSON и отправляет ответ.
// Включает все заголовки, установленные через SetHeader.
func (c *Context) JSON(response interface{}) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	reply := c.Msg.Reply
	// Если нет reply-subject — выходим
	if reply == "" {
		return nil
	}
	// Если есть заголовки — шлём PublishMsg
	if len(c.outHeader) > 0 {
		msg := &nats.Msg{
			Subject: reply,
			Header:  c.outHeader,
			Data:    data,
		}
		return c.router.nc.PublishMsg(msg)
	}
	return c.router.nc.Publish(reply, data)
}

// Header возвращает значение заголовка из запроса.
func (c *Context) Header(key string) string {
	if c.Msg.Header == nil {
		return ""
	}
	return c.Msg.Header.Get(key)
}

// SetHeader устанавливает заголовок для ответа.
func (c *Context) SetHeader(key, value string) {
	c.outHeader.Set(key, value)
}
