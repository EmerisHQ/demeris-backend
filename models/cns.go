package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// Chain represents CNS chain metadata row on the database.
type Chain struct {
	ID                          uint64         `db:"id" json:"-"`
	Enabled                     bool           `db:"enabled" json:"enabled"`                                                                  // boolean that marks whether the given chain is enabled or not (when enabled, API endpoints will return data)
	ChainName                   string         `db:"chain_name" binding:"required" json:"chain_name"`                                         // the unique name of the chain
	Logo                        string         `db:"logo" binding:"required" json:"logo"`                                                     // logo of the chain
	DisplayName                 string         `db:"display_name" binding:"required" json:"display_name"`                                     // user-friendly chain name
	CounterpartyNames           DbStringMap    `db:"counterparty_names" binding:"required" json:"counterparty_names"`                         // a mapping of client_id to chain names used to identify which chain a given client_id corresponds to
	PrimaryChannel              DbStringMap    `db:"primary_channel" binding:"required" json:"primary_channel"`                               // a mapping of chain name to primary channel
	Denoms                      DenomList      `db:"denoms" binding:"dive" json:"denoms"`                                                     // a list of denoms native to the chain
	DemerisAddresses            pq.StringArray `db:"demeris_addresses" binding:"required" json:"demeris_addresses"`                           // the addresses on which we accept fee payments
	BaseTxFee                   TxFee          `db:"base_tx_fee" binding:"required,dive" json:"base_tx_fee"`                                  // average cost (in dollar) to submit a transaction to the chain
	GenesisHash                 string         `db:"genesis_hash" binding:"required" json:"genesis_hash"`                                     // hash of the chain's genesis file
	NodeInfo                    NodeInfo       `db:"node_info" binding:"required,dive" json:"node_info"`                                      // info required to query full-node (e.g. to submit tx)
	ValidBlockThresh            Threshold      `db:"valid_block_thresh" binding:"required" json:"valid_block_thresh"`                         // valid block time expressed in time.Duration format
	MinimumThreshRelayerBalance int64          `db:"minimum_thresh_relayer_balance" binding:"required" json:"minimum_thresh_relayer_balance"` // minimum relayer balance threshold that a relayer account must contains
}

// VerifiedTokens returns a DenomList of native denoms that are verified.
func (c Chain) VerifiedTokens() DenomList {
	var ret DenomList
	for _, ft := range c.Denoms {
		if !ft.Verified {
			continue
		}

		ret = append(ret, ft)
	}

	return ret
}

// FeeTokens returns a DenomList of denoms that are usable as fee.
func (c Chain) FeeTokens() DenomList {
	var ret DenomList
	for _, ft := range c.Denoms {
		if !ft.FeeToken {
			continue
		}

		ret = append(ret, ft)
	}

	return ret
}

// Threshold is a database-friendly time.Duration.
type Threshold time.Duration

func (t *Threshold) UnmarshalJSON(bytes []byte) error {
	str := ""

	if err := json.Unmarshal(bytes, &str); err != nil {
		return err
	}

	d, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	*t = Threshold(d)

	return nil
}

func (t Threshold) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Duration().String())
}

// Duration returns t as time.Duration.
func (t Threshold) Duration() time.Duration {
	return time.Duration(t)
}

// Scan is the sql.Scanner implementation for Threshold.
func (t *Threshold) Scan(value interface{}) error {
	vs, ok := value.(string)
	if !ok {
		return fmt.Errorf("threshold value is of type %T, not string", value)
	}

	vsd, err := time.ParseDuration(vs)
	if err != nil {
		return fmt.Errorf("cannot parse value as duration, %w", err)
	}

	*t = Threshold(vsd)

	return nil
}

// Value is the driver.Value implementation for Threshold.
func (t Threshold) Value() (driver.Value, error) {
	td := time.Duration(t)
	return driver.Value(td.String()), nil
}

// NodeInfo holds information useful to connect to a full node and broadcast transactions.
type NodeInfo struct {
	Endpoint     string       `binding:"required" json:"endpoint"`
	ChainID      string       `binding:"required" json:"chain_id"`
	Bech32Config Bech32Config `binding:"required,dive" json:"bech32_config"`
}

// Scan is the sql.Scanner implementation for DbStringMap.
func (a *NodeInfo) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// TxFee holds levels of payment for fees.
type TxFee struct {
	Low     uint64 `binding:"required" json:"low"`
	Average uint64 `binding:"required" json:"average"`
	High    uint64 `binding:"required" json:"high"`
}

// Scan is the sql.Scanner implementation for DbStringMap.
func (a *TxFee) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// Bech32Config represents the chain's bech32 configuration
type Bech32Config struct {
	MainPrefix      string `json:"main_prefix" binding:"required"`
	PrefixAccount   string `json:"prefix_account" binding:"required"`
	PrefixValidator string `json:"prefix_validator" binding:"required"`
	PrefixConsensus string `json:"prefix_consensus" binding:"required"`
	PrefixPublic    string `json:"prefix_public" binding:"required"`
	PrefixOperator  string `json:"prefix_operator" binding:"required"`
}

// MarshalJSON implements the json.Marshaler interface.
// Returns the json representation of Bech32Config with prefixes methods as fields.
func (b Bech32Config) MarshalJSON() ([]byte, error) {
	var ret bech32ConfigMarshaled

	ret = bech32ConfigMarshaled{
		MainPrefix:      b.MainPrefix,
		PrefixAccount:   b.PrefixAccount,
		PrefixValidator: b.PrefixValidator,
		PrefixConsensus: b.PrefixConsensus,
		PrefixPublic:    b.PrefixPublic,
		PrefixOperator:  b.PrefixOperator,
		AccAddr:         b.Bech32PrefixAccAddr(),
		ValAddr:         b.Bech32PrefixAccAddr(),
		ValPub:          b.Bech32PrefixValPub(),
		ConsAddr:        b.Bech32PrefixConsAddr(),
		ConsPub:         b.Bech32PrefixConsPub(),
	}

	return json.Marshal(ret)
}

// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
func (b Bech32Config) Bech32PrefixAccAddr() string {
	return b.MainPrefix
}

// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
func (b Bech32Config) Bech32PrefixAccPub() string { return b.MainPrefix + b.PrefixPublic }

// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
func (b Bech32Config) Bech32PrefixValAddr() string {
	return b.MainPrefix + b.PrefixValidator + b.PrefixOperator
}

// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
func (b Bech32Config) Bech32PrefixValPub() string {
	return b.MainPrefix + b.PrefixValidator + b.PrefixOperator + b.PrefixPublic
}

// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
func (b Bech32Config) Bech32PrefixConsAddr() string {
	return b.MainPrefix + b.PrefixValidator + b.PrefixConsensus
}

// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
func (b Bech32Config) Bech32PrefixConsPub() string {
	return b.MainPrefix + b.PrefixValidator + b.PrefixConsensus + b.PrefixPublic
}

type bech32ConfigMarshaled struct {
	MainPrefix      string `json:"main_prefix" binding:"required"`
	PrefixAccount   string `json:"prefix_account" binding:"required"`
	PrefixValidator string `json:"prefix_validator" binding:"required"`
	PrefixConsensus string `json:"prefix_consensus" binding:"required"`
	PrefixPublic    string `json:"prefix_public" binding:"required"`
	PrefixOperator  string `json:"prefix_operator" binding:"required"`
	AccAddr         string `json:"acc_addr,omitempty" db:"-"`
	AccPub          string `json:"acc_pub,omitempty" db:"-"`
	ValAddr         string `json:"val_addr,omitempty" db:"-"`
	ValPub          string `json:"val_pub,omitempty" db:"-"`
	ConsAddr        string `json:"cons_addr,omitempty" db:"-"`
	ConsPub         string `json:"cons_pub,omitempty" db:"-"`
}

// Denom holds a token denomination and its verification status.
type Denom struct {
	Name        string `db:"name" binding:"required" json:"name,omitempty"`
	DisplayName string `db:"display_name" json:"display_name"`
	Logo        string `db:"logo" json:"logo,omitempty"`
	Precision   int64  `db:"precision" json:"precision,omitempty"`
	Verified    bool   `db:"verified" json:"verified,omitempty"`
	Stakable    bool   `db:"stakable" json:"stakable,omitempty"`
	Ticker      string `db:"ticker" json:"ticker,omitempty"`
	FeeToken    bool   `db:"fee_token" json:"fee_token,omitempty"`
}

// DenomList represents a slice of Denom.
type DenomList []Denom

// Scan is the sql.Scanner implementation for DenomList.
func (a *DenomList) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// DbStringMap represent a JSON database-enabled string map.
type DbStringMap map[string]string

// Scan is the sql.Scanner implementation for DbStringMap.
func (a *DbStringMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

// ChannelQuery represents a query to get a specified channel or counterparty data.
type ChannelQuery struct {
	ChainName    string `db:"chain_name" json:"chain_name"`
	Counterparty string `db:"key" json:"counterparty"`
	ChannelName  string `db:"value" json:"channel_name"`
}
