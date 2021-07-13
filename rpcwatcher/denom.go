package rpcwatcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/allinbits/demeris-backend/models"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type poolCoinDenomsInfo []denomInfo
type denomInfo struct {
	displayName string
	denom       string
	baseDenom   string
	ticker      string
	verified    bool
}

func (d denomInfo) isPoolCoin() bool {
	if len(d.denom) < 4 {
		return false
	}
	return d.denom[:4] == "pool"
}

func (d denomInfo) isIBCToken() bool {
	if len(d.denom) < 4 {
		return false
	}
	return d.denom[:4] == "ibc/"
}

func formatDenom(w *Watcher, data coretypes.ResultEvent) (models.Denom, error) {
	d := models.Denom{}

	poolCoinDenom, ok := data.Events["create_pool.pool_coin_denom"]

	if !ok {
		return d, fmt.Errorf("failed to read pool coin denom")
	}

	d.Name = poolCoinDenom[0]

	depositCoins, ok := data.Events["create_pool.deposit_coins"]
	if !ok {
		return d, fmt.Errorf("failed to read deposit coins")
	}

	coins, err := sdktypes.ParseCoinsNormalized(depositCoins[0])
	cosmoshub, err := w.cns.Chain("cosmos-hub")

	if err != nil {
		return d, err
	}

	var denomInfos poolCoinDenomsInfo

	for _, coin := range coins {
		denom := denomInfo{
			denom: coin.Denom,
		}

		if denom.isIBCToken() {

			verifiedTrace := VerifyTraceResponse{}
			w.l.Debugw("querying verified trace for coin", "coin", denom.denom)

			u, err := url.Parse(w.apiUrl)
			u.Path = fmt.Sprintf("chain/%s/denom/verify_trace/%s", "cosmos-hub", denom.denom[4:])

			endpoint := u.String()

			resp, err := http.Get(endpoint)

			if err != nil {
				return d, err
			}

			if resp.StatusCode == 200 {
				resp, err = http.Get(endpoint)

				if err != nil {
					return d, err
				}
			}

			dc := json.NewDecoder(resp.Body)

			err = dc.Decode(&verifiedTrace)

			if err != nil {
				return d, err
			}

			b, err := json.Marshal(verifiedTrace)

			if err != nil {
				return d, err
			}

			w.l.Debugw("got trace", "trace", string(b))

			if !verifiedTrace.VerifyTrace.Verified {
				return d, fmt.Errorf("not a verified denom")
			}
			if l := len(verifiedTrace.VerifyTrace.Trace); l != 1 {
				return d, fmt.Errorf("trace too long, expected 1, got %d", l)
			}

			denom.baseDenom = verifiedTrace.VerifyTrace.BaseDenom

			sourceChainName := verifiedTrace.VerifyTrace.Trace[0].CounterpartyName
			w.l.Debugw("checking base denom in chain", "denom", denom.baseDenom, "chain", sourceChainName)

			sourceChain, err := w.cns.Chain(sourceChainName)

			if err != nil {
				return d, err
			}

			found := false

			for _, dd := range sourceChain.Denoms {
				if dd.Name == denom.baseDenom {

					if !dd.Verified {
						return d, fmt.Errorf("denom not verified in source chain")
					}

					if dd.Ticker == "" {
						if dd.DisplayName == "" {
							denom.ticker = dd.Name
							denom.displayName = dd.Name
						} else {
							denom.ticker = dd.DisplayName
							denom.displayName = dd.DisplayName
						}
					} else {
						denom.ticker = dd.Ticker
						denom.displayName = dd.DisplayName
					}

					denom.verified = dd.Verified

					found = true
				}
			}

			if !found {
				return d, fmt.Errorf("denom not found")
			}
		} else {
			// check if token exists & is verified on cosmos hub
			denom.baseDenom = coin.Denom

			for _, dd := range cosmoshub.Denoms {
				if dd.Name == denom.baseDenom {

					if !dd.Verified {
						return d, fmt.Errorf("denom not verified in source chain")
					}

					if dd.Ticker == "" {
						if dd.DisplayName == "" {
							denom.ticker = dd.Name
							denom.displayName = dd.Name
						} else {
							denom.ticker = dd.DisplayName
							denom.displayName = dd.DisplayName
						}
					} else {
						denom.ticker = dd.Ticker
						denom.displayName = dd.DisplayName
					}

					denom.verified = dd.Verified

				}
			}

		}

		w.l.Debugw("adding verified denom", "denom", denom.displayName)

		denomInfos = append(denomInfos, denom)

	}

	if denomInfos[0].isPoolCoin() || denomInfos[0].isPoolCoin() {
		d.DisplayName = fmt.Sprintf("GDEX %s LP", poolCoinDenom[0])
		d.Ticker = fmt.Sprintf("G-%s", poolCoinDenom[0])
		d.Verified = true
	} else {
		d.DisplayName = fmt.Sprintf("GDEX %s/%s LP", denomInfos[0].displayName, denomInfos[1].displayName)
		d.Ticker = fmt.Sprintf("G-%s-%s", denomInfos[0].ticker, denomInfos[1].ticker)
		d.Verified = true
	}

	w.l.Debugw("verified lp denom", "displayname", d.DisplayName, "ticker", d.Ticker)

	return d, nil
}
