package navigator_cns

type Chain struct {
	ID          uint64 `db:"id" json:"-"`
	ClientID    string `db:"client_id" json:"client_id,omitempty" binding:"required"`
	ChainName   string `db:"chain_name" json:"chain_name,omitempty" binding:"required"`
	ChainID     string `db:"chain_id" json:"chain_id,omitempty" binding:"required"`
	NativeToken string `db:"native_token" json:"native_token,omitempty" binding:"required"`
}
