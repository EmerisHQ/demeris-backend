package gaia_processor

import (
	"bytes"
	"encoding/hex"

	"github.com/allinbits/demeris-backend/models"

	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"go.uber.org/zap"
)

type validatorsProcessor struct {
	l               *zap.SugaredLogger
	validatorsCache map[string]models.ValidatorRow
}

func (*validatorsProcessor) TableSchema() string {
	return createValidatorsTable
}

func (p *validatorsProcessor) ModuleName() string {
	return "validators"
}

func (p *validatorsProcessor) FlushCache() []tracelistener.WritebackOp {
	if len(p.validatorsCache) == 0 {
		return nil
	}

	l := make([]models.DatabaseEntrier, 0, len(p.validatorsCache))

	for _, c := range p.validatorsCache {
		l = append(l, c)
	}

	p.validatorsCache = map[string]models.ValidatorRow{}

	return []tracelistener.WritebackOp{
		{
			DatabaseExec: insertValidator,
			Data:         l,
		},
	}
}
func (b *validatorsProcessor) OwnsKey(key []byte) bool {
	return bytes.HasPrefix(key, types.ValidatorsKey)
}

func (b *validatorsProcessor) Process(data tracelistener.TraceOperation) error {

	v := types.Validator{}

	if err := p.cdc.UnmarshalBinaryBare(data.Value, &v); err != nil {
		return err
	}

	val := string(v.ConsensusPubkey.GetValue())

	k := hex.EncodeToString(data.Key)

	b.l.Debugw("new validator write",
		"operator_address", v.OperatorAddress,
		"height", data.BlockHeight,
		"txHash", data.TxHash,
		"cons pub key type", data.TxHash,
		"cons pub key", val,
		"key", k,
	)

	b.validatorsCache[val] = models.ValidatorRow{
		OperatorAddress:      v.OperatorAddress,
		ConsensusPubKeyType:  v.ConsensusPubkey.GetTypeUrl(),
		ConsensusPubKeyValue: v.ConsensusPubkey.Value,
		Jailed:               v.Jailed,
		Status:               int32(v.Status),
		Tokens:               v.Tokens.String(),
		DelegatorShares:      v.DelegatorShares.String(),
		Moniker:              v.Description.Moniker,
		Identity:             v.Description.Identity,
		Website:              v.Description.Website,
		SecurityContact:      v.Description.SecurityContact,
		Details:              v.Description.Details,
		UnbondingHeight:      v.UnbondingHeight,
		UnbondingTime:        v.UnbondingTime.String(),
		CommissionRate:       v.Commission.CommissionRates.Rate.String(),
		MaxRate:              v.Commission.CommissionRates.MaxRate.String(),
		MaxChangeRate:        v.Commission.CommissionRates.MaxChangeRate.String(),
		UpdateTime:           v.Commission.UpdateTime.String(),
		MinSelfDelegation:    v.MinSelfDelegation.String(),
	}

	return nil
}
