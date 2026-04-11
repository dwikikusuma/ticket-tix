package events

import "context"

type Message struct {
	Key     []byte
	Value   []byte
	Headers map[string]string
	Topic   string
}

type MessageHandler func(ctx context.Context, msg Message) error

type Middleware func(MessageHandler) MessageHandler
