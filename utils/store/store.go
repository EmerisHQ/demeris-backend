package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type Store struct {
	Client        *redis.Client
	ConnectionURL string
	Config        struct {
		ExpiryTime time.Duration
	}
}

// NewClient creates a new redis client
func NewClient(connUrl string) *Store {

	var store Store

	store.Client = redis.NewClient(&redis.Options{
		Addr: connUrl,
		DB:   0,
	})

	store.ConnectionURL = connUrl

	store.Config.ExpiryTime = time.Duration(300000000000)

	return &store

}

func (s *Store) CreateTicket(chain, txHash string) error {
	data := map[string]interface{}{
		"status": "pending",
	}

	b, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return s.Set(fmt.Sprintf("%s-%s", chain, txHash), string(b))
}

func (s *Store) SetComplete(key string) error {
	return s.Set(key, `{"status":"complete"}`)
}

func (s *Store) SetIBCReceiveFailed(key string) error {
	return s.Set(key, `{"status":"IBC_receive_failed"}`)
}

func (s *Store) SetIBCReceiveSuccess(key string) error {
	return s.Set(key, `{"status":"IBC_receive_success"}`)
}

func (s *Store) SetUnlockTimeout(key string) error {
	return s.Set(key, `{"status":"Tokens_unlocked_timeout"}`)
}

func (s *Store) SetUnlockAck(key string) error {
	return s.Set(key, `{"status":"Tokens_unlocked_ack"}`)
}

func (s *Store) SetFailedWithErr(key, error string) error {
	data := map[string]interface{}{
		"status": "failed",
		"err":    error,
	}

	b, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return s.Set(key, string(b))
}

func (s *Store) SetInTransit(key, destChain, sourceChannel, sendPacketSequence string) error {

	if !s.Exists(key) {
		return fmt.Errorf("key doesn't exists")
	}

	data := map[string]interface{}{
		"status": "transit",
	}

	b, err := json.Marshal(data)

	if err != nil {
		return err
	}

	if s.Set(key, string(b)) != nil {
		return err
	}

	newKey := fmt.Sprintf("%s-%s-%s", destChain, sourceChannel, sendPacketSequence)

	if s.Set(newKey, key) != nil {
		return err
	}

	return nil
}

func (s *Store) SetIbcTimeoutUnlock(key string) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetUnlockTimeout(prev)
}

func (s *Store) SetIbcAckUnlock(key string) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetUnlockAck(prev)
}

func (s *Store) SetIbcReceived(key string) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetComplete(prev)
}

func (s *Store) SetIbcFailed(key string) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetIBCReceiveFailed(prev)
}

func (s *Store) SetIbcSuccess(key string) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetIBCReceiveSuccess(prev)
}

func (s *Store) Exists(key string) bool {
	exists, _ := s.Client.Exists(ctx, key).Result()

	return exists == 1
}

func (s *Store) Set(key, value string) error {
	return s.Client.Set(ctx, key, value, s.Config.ExpiryTime).Err()
}

func (s *Store) Get(key string) (string, error) {
	return s.Client.Get(ctx, key).Result()
}
