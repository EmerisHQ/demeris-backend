package tx

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/allinbits/demeris-backend/models"
)

type TxData struct {
	TxBody     map[string]interface{} `json:"body"`
	AuthInfo   map[string]interface{} `json:"auth_info"`
	Signatures [][]byte               `json:"signatures"`
}

// UnmarshalJSON implements json.Unmarshaler on TxData.
// Signatures are canonically sent via JSON as base64-encoded strings.
// The rawTxData unmarshal the data type as sent from frontend, then we unmarshal into TxData as usual, but convert
// signatures from base64 strings to their bytes representation.
func (t *TxData) UnmarshalJSON(bytes []byte) error {
	raw := rawTxData{}
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return err
	}

	t.TxBody = raw.TxBody
	t.AuthInfo = raw.AuthInfo

	for i, s := range raw.Signatures {
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return fmt.Errorf("signature %d is not valid base64, %w", i, err)
		}

		t.Signatures = append(t.Signatures, b)
	}

	return nil
}

type rawTxData struct {
	TxBody     map[string]interface{} `json:"body"`
	AuthInfo   map[string]interface{} `json:"auth_info"`
	Signatures []string               `json:"signatures"`
}

type TxMeta struct {
	RelayOnly      bool
	TxType         string
	SignerSequence string
	FeePayer       string
	Valid          bool
	Chain          models.Chain
}

type TxResponse struct {
	Status   string `json:"status"`
	Sequence string `json:"sequence"`
	Key      string `json:"key"`
}
