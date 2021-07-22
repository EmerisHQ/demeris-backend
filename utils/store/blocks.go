package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var ErrBlockNotFound = fmt.Errorf("block not found")

const (
	blockFmt       = "block/%d"
	defaultTimeout = 100 * 10 * time.Second // we keep the last 100 blocks, assuming block time of 10 seconds
)

type Blocks struct {
	storeInstance *Store
}

func NewBlocks(s *Store) *Blocks {
	return &Blocks{storeInstance: s}
}

func blockKey(height int64) string {
	return fmt.Sprintf(blockFmt, height)
}

func (b *Blocks) Block(height int64) ([]byte, error) {
	res, err := b.storeInstance.Client.Get(context.Background(), blockKey(height)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrBlockNotFound
		}

		return nil, fmt.Errorf("redis error, %w", err)
	}

	return []byte(res), nil
}

func (b *Blocks) Add(data []byte, height int64) error {
	return b.storeInstance.Client.Set(context.Background(), blockKey(height), string(data), defaultTimeout).Err()
}
