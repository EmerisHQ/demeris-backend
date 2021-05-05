package models

import (
	"encoding/json"
	"errors"
)

// Chain represents CNS chain metadata row on the database.
type Chain struct {
	ID                uint64      `db:"id" json:"-"`
	ChainName         string      `db:"chain_name" binding:"required" json:"chain_name"`                 // the unique name of the chain
	Logo              string      `db:"logo" binding:"required" json:"logo"`                             // logo of the chain
	DisplayName       string      `db:"display_name" binding:"required" json:"display_name"`             // user-friendly chain name
	CounterpartyNames DbStringMap `db:"counterparty_names" binding:"required" json:"counterparty_names"` // a mapping of client_id to chain names used to identify which chain a given client_id corresponds to
	PrimaryChannel    DbStringMap `db:"primary_channel" binding:"required" json:"primary_channel"`       // a mapping of chain name to primary channel
	NativeDenoms      DenomList   `db:"native_denoms" binding:"required" json:"native_denoms"`           // a list of denoms native to the chain
	FeeTokens         DenomList   `db:"fee_tokens" binding:"required" json:"fee_tokens"`                 // a list of denoms accepted as fee on the chain, fee tokens must be verified
	FeeAddress        string      `db:"fee_address" binding:"required" json:"fee_address"`               // the address on which we accept fee payments
	PriceModifier     float64     `db:"price_modifier" binding:"required" json:"price_modifier"`         // modifier (between 0 and 1) applied when estimating the price of a token hopping through the chain
	BaseIBCFee        float64     `db:"base_ibc_fee" binding:"required" json:"base_ibc_fee"`             // average cost (in dollar) to submit an IBC transaction to the chain
	BaseFee           float64     `db:"base_fee" binding:"required" json:"base_fee"`                     // average cost (in dollar) to submit a transaction to the chain
	GenesisHash       string      `db:"genesis_hash" binding:"required" json:"genesis_hash"`             // hash of the chain's genesis file
	NodeInfo          NodeInfo    `db:"node_info" binding:"required" json:"node_info"`                   // info required to query full-node (e.g. to submit tx)
}

// VerifiedFeeTokens returns a DenomList of fee tokens that are verified.
func (c Chain) VerifiedFeeTokens() DenomList {
	var ret DenomList
	for _, ft := range c.FeeTokens {
		if !ft.Verified {
			continue
		}

		ret = append(ret, ft)
	}

	return ret
}

// VerifiedNativeDenoms returns a DenomList of native denoms that are verified.
func (c Chain) VerifiedNativeDenoms() DenomList {
	var ret DenomList
	for _, ft := range c.NativeDenoms {
		if !ft.Verified {
			continue
		}

		ret = append(ret, ft)
	}

	return ret
}

// NodeInfo holds information useful to connect to a full node and broadcast transactions.
type NodeInfo struct {
	Endpoint     string       `binding:"required" json:"endpoint"`
	ChainID      string       `binding:"required" json:"chain_id"`
	Bech32Config Bech32Config `binding:"required" json:"bech32_config"`
}

// Scan is the sql.Scanner implementation for DbStringMap.
func (a *NodeInfo) Scan(value interface{}) error {
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
	Logo      string `db:"logo" json:"logo,omitempty"`
	Precision int64  `db:"precision" binding:"required" json:"precision"`
	Name      string `db:"name" binding:"required" json:"name"`
	Verified  bool   `db:"verified" binding:"required" json:"verified"`
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
