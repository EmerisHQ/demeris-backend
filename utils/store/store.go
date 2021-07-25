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
	Info     string   `json:"info,omitempty"`
	Height   int64    `json:"height,omitempty"`
	Status   string   `json:"status,omitempty"`
	TxHashes []string `json:"tx_hashes,omitempty"`
	Error    string   `json:"error,omitempty"`
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

func (s *Store) CreateTicket(chain, txHash string) error {
	data := Ticket{
		Status: pending,
	}

	key := fmt.Sprintf("%s-%s", chain, txHash)
	if err := s.CreateShadowKey(key); err != nil {
		return err
	}

	return s.SetWithExpiry(key, data, 0)
}

func (s *Store) SetComplete(key string, height int64) error {

	return s.SetWithExpiry(key, Ticket{Status: complete,
		Height: height}, 2)
}

func (s *Store) SetIBCReceiveFailed(key string, txHashes []string, height int64) error {
	if err := s.CreateShadowKey(key); err != nil {
		return err
	}

	return s.SetWithExpiry(key, Ticket{Status: ibcReceiveFailed,
		TxHashes: txHashes, Height: height}, 0)
}

func (s *Store) SetIBCReceiveSuccess(key string, txHashes []string, height int64) error {
	if err := s.SetWithExpiry(key, Ticket{
		Status:   "IBC_receive_success",
		TxHashes: txHashes,
		Height:   height}, 2); err != nil {
		return err
	}

	return s.DeleteShadowKey(key)
}

func (s *Store) SetUnlockTimeout(key string, txHashes []string, height int64) error {
	if err := s.SetWithExpiry(key, Ticket{Status: tokensUnlockedTimeout,
		Height:   height,
		TxHashes: txHashes}, 2); err != nil {
		return err
	}

	return s.DeleteShadowKey(key)
}

func (s *Store) SetUnlockAck(key string, txHashes []string, height int64) error {
	if err := s.SetWithExpiry(key, Ticket{Status: tokensUnlockedAck,
		Height:   height,
		TxHashes: txHashes}, 2); err != nil {
		return err
	}

	return s.DeleteShadowKey(key)
}

func (s *Store) SetFailedWithErr(key, error string, height int64) error {
	data := Ticket{
		Height: height,
		Status: failed,
		Error:  error,
	}

	return s.SetWithExpiry(key, data, 2)
}

func (s *Store) SetInTransit(key, destChain, sourceChannel, sendPacketSequence, txHash string, height int64) error {

	if !s.Exists(key) {
		return fmt.Errorf("key doesn't exists")
	}

	if err := s.CreateShadowKey(key); err != nil {
		return err
	}

	data := Ticket{
		Status: transit,
		Height: height,
	}

	if err := s.SetWithExpiry(key, data, 2); err != nil {
		return err
	}

	newKey := fmt.Sprintf("%s-%s-%s", destChain, sourceChannel, sendPacketSequence)

	if err := s.SetWithExpiry(newKey, Ticket{Info: key,
		TxHashes: []string{txHash}}, 2); err != nil {
		return err
	}

	return nil
}

func (s *Store) SetIbcTimeoutUnlock(key, txHash string, height int64) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	txHashes := append(prev.TxHashes, txHash)
	return s.SetUnlockTimeout(prev.Info, txHashes, height)
}

func (s *Store) SetIbcAckUnlock(key, txHash string, height int64) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	txHashes := append(prev.TxHashes, txHash)
	return s.SetUnlockAck(prev.Info, txHashes, height)
}

func (s *Store) SetIbcReceived(key, txHash string, height int64) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	txHashes := append(prev.TxHashes, txHash)
	fmt.Printf("this is prev %v", prev)
	return s.SetIBCReceiveSuccess(prev.Info, txHashes, height)
}

func (s *Store) SetIbcFailed(key, txHash string, height int64) error {

	prev, err := s.Get(key)

	if err != nil {
		return err
	}

	txHashes := append(prev.TxHashes, txHash)
	return s.SetIBCReceiveFailed(prev.Info, txHashes, height)
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
	if err := s.Client.Get(context.Background(), key).Scan(&res); err != nil {
		return Ticket{}, err
	}

	return res, nil
}

func (s *Store) Delete(key string) error {
	return s.Client.Del(context.Background(), key).Err()
}

func (s *Store) DeleteShadowKey(key string) error {
	shadowKey := shadow + key
	return s.Delete(shadowKey)
}
