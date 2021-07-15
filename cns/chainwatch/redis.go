package chainwatch

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	r "github.com/go-redis/redis/v8"
)

const (
	setPrefix            = "cns/waitingchains"
	chainStatusPrefixFmt = "cns/ChainStatus/%s"
)

//go:generate stringer -type=ChainStatus
type ChainStatus uint

func (cs *ChainStatus) UnmarshalBinary(data []byte) error {
	rr, err := binary.ReadUvarint(bytes.NewReader(data))
	if err != nil {
		*cs = undefined
		return err
	}

	*cs = ChainStatus(rr)
	return nil
}

func (cs ChainStatus) MarshalBinary() (data []byte, err error) {
	dest := make([]byte, 8) // size of a uint
	binary.PutUvarint(dest, uint64(cs))

	return dest, nil
}

const (
	starting ChainStatus = iota
	running
	relayerConnecting
	done
	undefined
)

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

func chainStatusKey(chainName string) string {
	return fmt.Sprintf(chainStatusPrefixFmt, chainName)
}

func (c *Connection) SetChainStatus(chainName string, status ChainStatus) error {
	return c.conn.Set(context.Background(), chainStatusKey(chainName), status, 0).Err()
}

func (c *Connection) ChainStatus(chainName string) (ChainStatus, bool, error) {
	var cs ChainStatus
	getRes := c.conn.Get(context.Background(), chainStatusKey(chainName))
	err := getRes.Err()
	if err != nil {
		if errors.Is(err, r.Nil) {
			return undefined, false, nil
		}

		return undefined, false, err
	}

	err = getRes.Scan(&cs)

	return cs, true, err
}

func (c *Connection) DeleteChainStatus(chainName string) error {
	return c.conn.Del(context.Background(), chainStatusKey(chainName)).Err()
}

func (c *Connection) HasChainStatus(chainName string) (bool, error) {
	res, err := c.conn.Exists(context.Background(), chainStatusKey(chainName)).Result()
	if err != nil {
		return false, err
	}

	return res == 1, nil
}
