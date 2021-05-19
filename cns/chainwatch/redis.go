package chainwatch

import (
	"context"
	"fmt"

	r "github.com/go-redis/redis/v8"
)

const setPrefix = "cns/waitingchains"

type Connection struct {
	conn *r.Client
}

func NewConnection(host string) (*Connection, error) {
	c := r.NewClient(&r.Options{
		Addr:     host,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := context.Background()
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cannot connect to redis, %w", err)
	}

	return &Connection{
		conn: c,
	}, nil
}

func (c *Connection) AddChain(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}

	return c.conn.SAdd(context.Background(), setPrefix, name).Err()
}

func (c *Connection) RemoveChain(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}

	return c.conn.SRem(context.Background(), setPrefix, name).Err()
}

func (c *Connection) HasChain(name string) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("empty name")
	}

	return c.conn.SIsMember(context.Background(), setPrefix, name).Result()
}

func (c *Connection) Chains() ([]string, error) {
	return c.conn.SMembers(context.Background(), setPrefix).Result()
}
