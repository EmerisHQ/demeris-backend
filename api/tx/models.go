package tx

import (
	"github.com/allinbits/demeris-backend/models"
)

type TxRequest struct {
	TxBytes []byte `json:"tx_bytes"`
}

// type Amount interface {
// 	Foo() error
// }
// type AmountSingle struct {
// 	Denom  string `json:"denom"`
// 	Amount string `json:"amount"`
// }

// type AmountList []AmountSingle

// func (a *AmountList) Foo() error {
// 	return nil
// }

// func (a *AmountSingle) Foo() error {
// 	return nil
// }

// type TxData struct {
// 	TxBody struct {
// 		Messages []struct {
// 			Type        string `json:"@type"`
// 			FromAddress string `json:"from_address,omitempty"`
// 			ToAddress   string `json:"to_address,omitempty"`
// 			Amount      []struct {
// 				Denom  string `json:"denom"`
// 				Amount string `json:"amount"`
// 			} `json:"amount,omitempty"`
// 			SourcePort    string `json:"source_port,omitempty"`
// 			SourceChannel string `json:"source_channel,omitempty"`
// 			Token         struct {
// 				Denom  string `json:"denom"`
// 				Amount string `json:"amount"`
// 			} `json:"token,omitempty"`
// 			Sender        string `json:"sender,omitempty"`
// 			Receiver      string `json:"receiver,omitempty"`
// 			TimeoutHeight struct {
// 				RevisionNumber string `json:"revision_number"`
// 				RevisionHeight string `json:"revision_height"`
// 			} `json:"timeout_height,omitempty"`
// 			TimeoutTimestamp string `json:"timeout_timestamp,omitempty"`
// 			DelegatorAddress string `json:"delegator_address,omitempty"`
// 			ValidatorAddress string `json:"validator_address,omitempty"`
// 		} `json:"messages"`
// 		Memo                        string        `json:"memo"`
// 		TimeoutHeight               string        `json:"timeout_height"`
// 		ExtensionOptions            []interface{} `json:"extension_options"`
// 		NonCriticalExtensionOptions []interface{} `json:"non_critical_extension_options"`
// 	} `json:"body"`
// 	AuthInfo struct {
// 		SignerInfos []struct {
// 			PublicKey struct {
// 				Type string `json:"@type"`
// 				Key  string `json:"key"`
// 			} `json:"public_key"`
// 			ModeInfo struct {
// 				Single struct {
// 					Mode string `json:"mode"`
// 				} `json:"single"`
// 			} `json:"mode_info"`
// 			Sequence string `json:"sequence"`
// 		} `json:"signer_infos"`
// 		Fee struct {
// 			Amount []struct {
// 				Denom  string `json:"denom"`
// 				Amount string `json:"amount"`
// 			} `json:"amount,omitempty"`
// 			GasLimit string `json:"gas_limit"`
// 			Payer    string `json:"payer"`
// 			Granter  string `json:"granter"`
// 		} `json:"fee"`
// 	} `json:"auth_info"`
// 	Signatures []string `json:"signatures"`
// }

// UnmarshalJSON implements json.Unmarshaler on TxData.
// Signatures are canonically sent via JSON as base64-encoded strings.
// The rawTxData unmarshal the data type as sent from frontend, then we unmarshal into TxData as usual, but convert
// signatures from base64 strings to their bytes representation.
// func (t *TxData) UnmarshalJSON(bytes []byte) error {
// 	raw := rawTxData{}
// 	if err := json.Unmarshal(bytes, &raw); err != nil {
// 		return err
// 	}

// 	t.TxBody = raw.TxBody
// 	t.AuthInfo = raw.AuthInfo

// 	for i, s := range raw.Signatures {
// 		b, err := base64.StdEncoding.DecodeString(s)
// 		if err != nil {
// 			return fmt.Errorf("signature %d is not valid base64, %w", i, err)
// 		}

// 		t.Signatures = append(t.Signatures, b)
// 	}

// 	return nil
// }

// type rawTxData struct {
// 	TxBody     map[string]interface{} `json:"body"`
// 	AuthInfo   map[string]interface{} `json:"auth_info"`
// 	Signatures []string               `json:"signatures"`
// }

type TxMeta struct {
	RelayOnly      bool
	TxType         string
	Signer         string
	SignerSequence string
	FeePayer       string
	Valid          bool
	Chain          models.Chain
}

type TxResponse struct {
	Ticket string `json:"ticket"`
}
