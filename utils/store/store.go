package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

const (
	Pending               = "pending"
	Transit               = "transit"
	Complete              = "complete"
	Failed                = "failed"
	Shadow                = "shadow"
	IBCReceiveFailed      = "IBC_receive_failed"
	TokensUnlockedTimeout = "Tokens_unlocked_timeout"
	TokensUnlockedAck     = "Tokens_unlocked_ack"
)

type Store struct {
	Client        *redis.Client
	ConnectionURL string
	Config        struct {
		ExpiryTime time.Duration
	}
}

type Ticket struct {
	Info   string `json:"info,omitempty"`
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

func (t *Ticket) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, t)
}
func (t Ticket) MarshalBinary() (data []byte, err error) {
	return json.Marshal(t)
}

// NewClient creates a new redis client
func NewClient(connUrl string) *Store {

	var store Store

	store.Client = redis.NewClient(&redis.Options{
		Addr: connUrl,
		DB:   0,
	})
	store.Client.Do(store.Client.Context(), "CONFIG", "SET", "notify-keyspace-events", "KEA")
	store.ConnectionURL = connUrl

	store.Config.ExpiryTime = 300 * time.Second

	return &store

}

func (s *Store) CreateTicket(chain, txHash string) error {
	data := Ticket{
		Status: Pending,
	}

	key := fmt.Sprintf("%s-%s", chain, txHash)
	if err := s.CreateShadowKey(key); err != nil {
		return err
	}

	return s.SetWithExpiry(key, data, 0)
}

func (s *Store) SetComplete(key string) error {

	return s.SetWithExpiry(key, Ticket{Status: Complete}, 2)
}

func (s *Store) SetIBCReceiveFailed(key string) error {
	if err := s.CreateShadowKey(key); err != nil {
		return err
	}

	return s.SetWithExpiry(key, Ticket{Status: IBCReceiveFailed}, 0)
}

func (s *Store) SetIBCReceiveSuccess(key string) error {
	return s.SetWithExpiry(key, Ticket{
		Status: "IBC_receive_success"}, 2)
}

func (s *Store) SetUnlockTimeout(key string) error {
	return s.SetWithExpiry(key, Ticket{Status: TokensUnlockedTimeout}, 2)
}

func (s *Store) SetUnlockAck(key string) error {
	return s.SetWithExpiry(key, Ticket{Status: TokensUnlockedAck}, 2)
}

func (s *Store) SetFailedWithErr(key, error string) error {
	data := Ticket{
		Status: Failed,
		Error:  error,
	}

	return s.SetWithExpiry(key, data, 2)
}

func (s *Store) SetInTransit(key, destChain, sourceChannel, sendPacketSequence string) error {

	if !s.Exists(key) {
		return fmt.Errorf("key doesn't exists")
	}

	if err := s.CreateShadowKey(key); err != nil {
		return err
	}

	data := Ticket{
		Status: Transit,
	}

	if err := s.SetWithExpiry(key, data, 0); err != nil {
		return err
	}

	newKey := fmt.Sprintf("%s-%s-%s", destChain, sourceChannel, sendPacketSequence)

	if err := s.SetWithExpiry(newKey, Ticket{Info: key}, 2); err != nil {
		return err
	}

	return nil
}

func (s *Store) SetIbcTimeoutUnlock(key string) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetUnlockTimeout(prev.Info)
}

func (s *Store) SetIbcAckUnlock(key string) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetUnlockAck(prev.Info)
}

func (s *Store) SetIbcReceived(key string) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetIBCReceiveSuccess(prev.Info)
}

func (s *Store) SetIbcFailed(key string) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetIBCReceiveFailed(prev.Info)
}

func (s *Store) CreateShadowKey(key string) error {
	shadow := Shadow + key
	return s.SetWithExpiry(shadow, "", 1)
}

func (s *Store) Exists(key string) bool {
	exists, _ := s.Client.Exists(ctx, key).Result()

	return exists == 1
}

func (s *Store) SetWithExpiry(key string, value interface{}, mul int64) error {

	return s.Client.Set(ctx, key, value, time.Duration(mul)*(s.Config.ExpiryTime)).Err()
}

func (s *Store) Get(key string) (Ticket, error) {
	var res Ticket
	if err := s.Client.Get(ctx, key).Scan(&res); err != nil {
		return Ticket{}, err
	}
	return res, nil
}
