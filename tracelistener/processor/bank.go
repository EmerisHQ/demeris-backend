package processor

import (
	"encoding/hex"

	"github.com/allinbits/demeris-backend/tracelistener/processor/sdk/bank"

	"github.com/allinbits/demeris-backend/models"

	"go.uber.org/zap"

	"github.com/allinbits/demeris-backend/tracelistener"
)

type bankCacheEntry struct {
	address string
	denom   string
}

type bankProcessor struct {
	l           *zap.SugaredLogger
	heightCache map[bankCacheEntry]models.BalanceRow
	parser      bank.Parser
}

func (*bankProcessor) TableSchema() string {
	return createBalancesTable
}

func (b *bankProcessor) ModuleName() string {
	return "bank"
}

func (b *bankProcessor) FlushCache() []tracelistener.WritebackOp {
	if len(b.heightCache) == 0 {
		return nil
	}

	l := make([]models.DatabaseEntrier, 0, len(b.heightCache))

	for _, v := range b.heightCache {
		l = append(l, v)
	}

	b.heightCache = map[bankCacheEntry]models.BalanceRow{}

	return []tracelistener.WritebackOp{
		{
			DatabaseExec: insertBalance,
			Data:         l,
		},
	}
}

func (b *bankProcessor) OwnsKey(key []byte) bool {
	return b.parser.OwnsKey(key)
}

func (b *bankProcessor) Process(data tracelistener.TraceOperation) error {
	addr, coins, err := b.parser.Process(p.cdc, data)
	if err != nil {
		return err
	}

	hAddr := hex.EncodeToString([]byte(addr))
	b.l.Debugw("new bank store write",
		"operation", data.Operation,
		"address", hAddr,
		"new_balance", coins.String(),
		"height", data.BlockHeight,
		"txHash", data.TxHash,
	)

	for _, coin := range coins {
		b.heightCache[bankCacheEntry{
			address: hAddr,
			denom:   coin.Denom,
		}] = models.BalanceRow{
			Address:     hAddr,
			Amount:      coin.String(),
			Denom:       coin.Denom,
			BlockHeight: data.BlockHeight,
		}
	}

	return nil
}
