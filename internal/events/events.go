package events

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const ConfigChannel = "gateway:config_events"

const (
	EventRouteCreated = "route:created"
	EventRouteDeleted = "route:deleted"
	EventServiceCreated = "service:created"
	EventServiceDeleted = "service:deleted"
	EventAPIKeyCreated = "apikey:created"
	EventAPIKeyRevoked = "apikey:revoked"
)

type ConfigEvent struct {
	Type string `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

type Broadcaster struct {
	rdb *redis.Client
}

func NewBroadcaster(rdb *redis.Client) *Broadcaster {
	return &Broadcaster{rdb: rdb}
}

func (b *Broadcaster) Publish(ctx context.Context, eventType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	event := ConfigEvent {
		Type: eventType,
		Payload: data,
		Timestamp: time.Now().UTC(),
	}

	msg, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return b.rdb.Publish(ctx, ConfigChannel, msg).Err()
}

func StartSubscriber(ctx context.Context, rdb *redis.Client, handleFunc func(event ConfigEvent)) {
	pubsub := rdb.Subscribe(ctx, ConfigChannel)

	go func() {
		defer pubsub.Close()
		log.Printf("Subscribed to Redis Pub/Sub channel: %s", ConfigChannel)

		ch := pubsub.Channel()
		for msg := range ch {
			var event ConfigEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("[PubSub Error] failed to parse event: %v", err)
				continue
			}

			handleFunc(event)
		}
	}()
}