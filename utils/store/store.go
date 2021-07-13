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
	timeout = 500 * time.Millisecond
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
	store.Client.Do(store.Client.Context(),"CONFIG", "SET", "notify-keyspace-events", "KEA")
	store.ConnectionURL = connUrl

	store.Config.ExpiryTime = time.Duration(300000000000)

	return &store

}

func (s *Store) CreateTicket(chain, txHash string) error {
	data := Ticket{
		Status: "pending",
	}

	return s.Set(fmt.Sprintf("%s-%s", chain, txHash), data, s.Config.ExpiryTime)
}

func (s *Store) SetComplete(key string) error {
	shadow := "shadow" + key
	if err := s.Set(shadow, Ticket{}, timeout); err != nil {
		return err
	}

	return s.Set(key, Ticket{Status: "complete"}, s.Config.ExpiryTime)
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

	//if !s.Exists(key) {
	//	return fmt.Errorf("key doesn't exists")
	//}

	shadow := "shadow" + key
	if err := s.Set(shadow, Ticket{}, timeout); err != nil {
		return err
	}

	data :=Ticket{
		Status: "transit",
	}

	if err := s.Set(key, data, s.Config.ExpiryTime); err != nil {
		return err
	}

	newKey := fmt.Sprintf("%s-%s-%s", destChain, sourceChannel, sendPacketSequence)

	if err := s.Set(newKey, Ticket{Info: key}, s.Config.ExpiryTime); err != nil {
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

	return s.SetIBCReceiveSuccess(prev.Info)
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

func (s *Store) Set(key string, value Ticket, expiry time.Duration,info ...string,) error {
	return s.Client.Set(ctx, key, value, expiry).Err()
}

func (s *Store) Get(key string) (Ticket, error) {
	var res Ticket
	 if err := s.Client.Get(ctx, key).Scan(&res); err != nil{
	 	return Ticket{}, err
	 }
	return res, nil
}
