package tx

import "github.com/allinbits/demeris-backend/models"

type TxData struct {
	TxBody     map[string]interface{} `json:"body"`
	AuthInfo   map[string]interface{} `json:"auth_info"`
	Signatures [][]byte               `json:"signatures"`
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
