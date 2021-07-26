package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	pending               = "pending"
	transit               = "transit"
	complete              = "complete"
	failed                = "failed"
	shadow                = "shadow"
	ibcReceiveFailed      = "IBC_receive_failed"
	tokensUnlockedTimeout = "Tokens_unlocked_timeout"
	tokensUnlockedAck     = "Tokens_unlocked_ack"
)

type Store struct {
	Client        *redis.Client
	ConnectionURL string
	Config        struct {
		ExpiryTime time.Duration
	}
}

type Ticket struct {
	Owner  string `json:"owner,omitempty"`
	Info   string `json:"info,omitempty"`
	Height int64  `json:"height,omitempty"`
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
func NewClient(connUrl string) (*Store, error) {

	var store Store

	store.Client = redis.NewClient(&redis.Options{
		Addr: connUrl,
		DB:   0,
	})

	store.ConnectionURL = connUrl

	store.Config.ExpiryTime = 300 * time.Second

	return &store, nil

}

func (s *Store) CreateTicket(chain, txHash, owner string) error {
	data := Ticket{
		Owner:  owner,
		Status: pending,
	}

	key := fmt.Sprintf("%s-%s", chain, txHash)
	if err := s.CreateShadowKey(key); err != nil {
		return err
	}

	if err := s.SetWithExpiry(key, data, 0); err != nil {
		return err
	}

	return s.Client.SAdd(context.Background(), owner, key).Err()
}

func (s *Store) SetComplete(key string, height int64) error {
	ticket, err := s.Get(key)
	if err != nil {
		return err
	}

	if err := s.SetWithExpiry(key, Ticket{Status: complete,
		Height: height}, 2); err != nil {
		return err
	}

	return s.SRemove(ticket.Owner, key)
}

func (s *Store) SetIBCReceiveFailed(key string, height int64) error {
	if err := s.CreateShadowKey(key); err != nil {
		return err
	}

	return s.SetWithExpiry(key, Ticket{Status: ibcReceiveFailed, Height: height}, 0)
}

func (s *Store) SetIBCReceiveSuccess(key, owner string, height int64) error {
	if err := s.SetWithExpiry(key, Ticket{
		Status: "IBC_receive_success",
		Height: height}, 2); err != nil {
		return err
	}

	if err := s.DeleteShadowKey(key); err != nil {
		return err
	}

	return s.SRemove(owner, key)
}

func (s *Store) SetUnlockTimeout(key, owner string, height int64) error {
	if err := s.SetWithExpiry(key, Ticket{Status: tokensUnlockedTimeout,
		Height: height}, 2); err != nil {
		return err
	}

	if err := s.DeleteShadowKey(key); err != nil {
		return err
	}

	return s.SRemove(owner, key)
}

func (s *Store) SetUnlockAck(key, owner string, height int64) error {
	if err := s.SetWithExpiry(key, Ticket{Status: tokensUnlockedAck,
		Height: height}, 2); err != nil {
		return err
	}

	if err := s.DeleteShadowKey(key); err != nil {
		return err
	}

	return s.SRemove(owner, key)
}

func (s *Store) SetFailedWithErr(key, error string, height int64) error {
	if !s.Exists(key) {
		return fmt.Errorf("key doesn't exists")
	}

	prev, err := s.Get(key)
	if err != nil {
		return err
	}

	data := Ticket{
		Height: height,
		Status: failed,
		Error:  error,
	}

	if err := s.SetWithExpiry(key, data, 2); err != nil {
		return err
	}

	return s.SRemove(prev.Owner, key)
}

func (s *Store) SetInTransit(key, destChain, sourceChannel, sendPacketSequence string, height int64) error {

	if !s.Exists(key) {
		return fmt.Errorf("key doesn't exists")
	}

	if err := s.CreateShadowKey(key); err != nil {
		return err
	}

	ticket, err := s.Get(key)
	if err != nil {
		return err
	}

	ticket.Status = pending
	if err := s.SetWithExpiry(key, ticket, 2); err != nil {
		return err
	}

	newKey := fmt.Sprintf("%s-%s-%s", destChain, sourceChannel, sendPacketSequence)

	return s.SetWithExpiry(newKey, Ticket{Info: key,
		Owner: ticket.Owner}, 2)
}

func (s *Store) SetIbcTimeoutUnlock(key string, height int64) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetUnlockTimeout(prev.Info, prev.Owner, height)
}

func (s *Store) SetIbcAckUnlock(key string, height int64) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetUnlockAck(prev.Info, prev.Owner, height)
}

func (s *Store) SetIbcReceived(key string, height int64) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetIBCReceiveSuccess(prev.Info, prev.Owner, height)
}

func (s *Store) SetIbcFailed(key string, height int64) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	return s.SetIBCReceiveFailed(prev.Info, height)
}
func (s *Store) CreateShadowKey(key string) error {
	shadowKey := shadow + key
	return s.SetWithExpiry(shadowKey, "", 1)
}

func (s *Store) Exists(key string) bool {
	exists, _ := s.Client.Exists(context.Background(), key).Result()

	return exists == 1
}

func (s *Store) SetWithExpiry(key string, value interface{}, mul int64) error {
	return s.Client.Set(context.Background(), key, value, time.Duration(mul)*(s.Config.ExpiryTime)).Err()
}

func (s *Store) Get(key string) (Ticket, error) {
	var res Ticket
	fmt.Printf("\nTHIS IS  KEY %s", key)
	if err := s.Client.Get(context.Background(), key).Scan(&res); err != nil {
		return Ticket{}, err
	}

	return res, nil
}

func (s *Store) GetUserTickets(user string) ([]string, error) {
	var keys []string
	if err := s.Client.SMembers(context.Background(), user).ScanSlice(&keys); err != nil {
		return []string{}, err
	}

	return keys, nil
}

func (s *Store) Delete(key string) error {
	return s.Client.Del(context.Background(), key).Err()
}

func (s *Store) DeleteShadowKey(key string) error {
	shadowKey := shadow + key
	return s.Delete(shadowKey)
}

func (s *Store) SRemove(user, key string) error {
	return s.Client.SRem(context.Background(), user, key).Err()
}
