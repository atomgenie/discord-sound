package redis

import (
	"context"
	"time"

	"github.com/mediocregopher/radix/v4"
)

// SubscribePubSub Subscribe to topic with redis
func SubscribePubSub(topic ...string) (chan radix.PubSubMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := radix.Dial(ctx, "tcp", Client.addr)

	if err != nil {
		return nil, err
	}

	subRedis := (radix.PubSubConfig{}).New(conn)
	chanSub := make(chan radix.PubSubMessage)

	if err := subRedis.Subscribe(ctx, chanSub, topic...); err != nil {
		return nil, err
	}

	return chanSub, nil
}
