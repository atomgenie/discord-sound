package redis

import (
	"context"

	"github.com/mediocregopher/radix/v4"
)

// Client Redis client
var Client ClientType = ClientType{}

// ClientType Redis Client
type ClientType struct {
	Client radix.Client
	addr   string
}

// Init Redis client
func Init(addr string) error {
	client, err := (radix.PoolConfig{}).New(context.Background(), "tcp", addr)

	if err != nil {
		return err
	}

	Client.Client = client
	Client.addr = addr
	return nil
}
