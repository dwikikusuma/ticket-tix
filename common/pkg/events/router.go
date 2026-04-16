package events

import (
	"context"
	"fmt"
	"sync"
)

type Router struct {
	mu      sync.RWMutex
	handler map[string]MessageHandler
}

func NewRouter() *Router {
	return &Router{
		handler: make(map[string]MessageHandler),
	}
}

func (r *Router) Handle(topic string, handler MessageHandler, mw ...Middleware) {
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i](handler)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.handler[topic] = handler
}

func (r *Router) Route(ctx context.Context, msg Message) error {
	r.mu.RLock()
	handler, ok := r.handler[msg.Topic]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("no handler for topic '%s'", msg.Topic)
	}
	return handler(ctx, msg)
}

func (r *Router) GetTopics() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	topics := make([]string, 0, len(r.handler))

	for topic, _ := range r.handler {
		topics = append(topics, topic)
	}
	return topics
}
