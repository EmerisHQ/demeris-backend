package chainwatch

import (
	"context"
	"encoding/json"
	"fmt"

	r "github.com/go-redis/redis/v8"
)

const setPrefix = "cns/waitingchains"

type Connection struct {
	conn *r.Client
}

type Chain struct {
	Name          string
	AddressPrefix string
	HasFaucet     bool
	HDPath        string
}

func (cc Chain) Validate() error {
	if cc.Name == "" {
		return fmt.Errorf("empty name")
	}

	if cc.AddressPrefix == "" {
		return fmt.Errorf("empty address prefix")
	}

	if cc.HDPath == "" {
		return fmt.Errorf("empty HD path")
	}

	return nil
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

func (c *Connection) AddChain(cc Chain) error {
	if err := cc.Validate(); err != nil {
		return err
	}

	data, err := json.Marshal(cc)
	if err != nil {
		return err
	}

	return c.conn.SAdd(context.Background(), setPrefix, data).Err()
}

func (c *Connection) RemoveChain(cc Chain) error {
	if err := cc.Validate(); err != nil {
		return err
	}

	data, err := json.Marshal(cc)
	if err != nil {
		return err
	}

	return c.conn.SRem(context.Background(), setPrefix, data).Err()
}

func (c *Connection) HasChain(name string) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("empty name")
	}

	return c.conn.SIsMember(context.Background(), setPrefix, name).Result()
}

func (c *Connection) Chains() ([]Chain, error) {
	var ret []Chain

	result, err := c.conn.SMembers(context.Background(), setPrefix).Result()
	if err != nil {
		return nil, err
	}

	for _, rr := range result {
		d := Chain{}

		if err := json.Unmarshal([]byte(rr), &d); err != nil {
			return nil, err
		}

		ret = append(ret, d)
	}

	return ret, nil
}
