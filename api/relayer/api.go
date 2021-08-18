package relayer

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/allinbits/demeris-backend/models"

	"github.com/allinbits/demeris-backend/api/database"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, store *persistence.InMemoryStore) {
	rel := router.Group("/relayer")

	rel.GET("/status", cache.CachePage(store, 10*time.Second, getRelayerStatus))
	rel.GET("/balance", cache.CachePage(store, 10*time.Second, getRelayerBalance))
}

// getRelayerStatus returns status of relayer.
// @Summary Gets relayer status
// @Tags Relayer
// @ID relayer-status
// @Description gets relayer status
// @Produce json
// @Success 200 {object} relayerStatusResponse
// @Failure 500,403 {object} deps.Error
// @Router /relayer/status [get]
func getRelayerStatus(c *gin.Context) {
	var res relayerStatusResponse

	//d := deps.GetDeps(c)

	res.Running = true

	//if errors.Is(err, k8s.ErrNotFound) || running.Status.Phase != v1.RelayerPhaseRunning {
	//	res.Running = false
	//}

	c.JSON(http.StatusOK, res)
}

// getRelayerBalance returns the balance of the various relayer accounts.
// @Summary Gets relayer balance for the various relayer accounts
// @Tags Relayer
// @ID relayer-balance
// @Description gets relayer balance for the various relayer accounts
// @Produce json
// @Success 200 {object} relayerBalances
// @Failure 500,403 {object} deps.Error
// @Router /relayer/balance [get]
func getRelayerBalance(c *gin.Context) {
	var res relayerBalances

	d := deps.GetDeps(c)

	chains := []string{"irishub-1", "akashnet-2", "regen-1", "sentinelhub-2", "crypto-org-chain-mainnet-1", "cosmoshub-4", "core-1", "osmosis-1"}
	addresses := []string{"iaa1nl68echj5q5xqmelt9njaz82ehe7dgre2vfu93", "akash1da0ad723jhv0k7zgrv3szupngdmtqhkkad6dkz", "regen1uft3gakywrrw52qafu7ej8cl0avd308qraqg89",
		"sent13t6r4eczclzce75y99haetc787dxnr6zg3sfmn", "cro18llfnp6sd9qwhc4km0cf6e043jjr30lpkd4akh", "cosmos19nwjcna4vc7nf5f5ykfnkenvpj7cng3drcejkt",
		"persistence1l0u02hdzl27rf00707vqaat7tswpxqewc9zkw8", "osmo1kcm2rgsk5q9cj72y2wfuym4a9ka96njxx85zlp"}

	thresh, err := relayerThresh(chains, d.Database)
	if err != nil {
		e := deps.NewError(
			"status",
			fmt.Errorf("cannot retrieve relayer status"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve relayer status",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	for i := 0; i < len(addresses); i++ {
		t, found := thresh[chains[i]]
		if !found {
			continue
		}

		enough, err := enoughBalance(addresses[i], t, d.Database)
		if err != nil {
			e := deps.NewError(
				"status",
				fmt.Errorf("cannot retrieve relayer status"),
				http.StatusBadRequest,
			)

			d.WriteError(c, e,
				"cannot retrieve relayer status",
				"id",
				e.ID,
				"error",
				err,
			)

			return
		}

		res.Balances = append(res.Balances, relayerBalance{
			Address:       addresses[i],
			ChainName:     chains[i],
			EnoughBalance: enough,
		})

	}

	c.JSON(http.StatusOK, res)

}

func relayerThresh(chains []string, db *database.Database) (map[string]models.Denom, error) {
	res := map[string]models.Denom{}

	for _, cn := range chains {
		chain, err := db.ChainFromChainID(cn)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue // probably the chain isn't enabled yet
			}
			return nil, fmt.Errorf("cannot retrieve chain %s, %w", cn, err)
		}

		res[cn] = chain.RelayerToken()
	}

	return res, nil
}

func enoughBalance(address string, denom models.Denom, db *database.Database) (bool, error) {
	_, hb, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return false, err
	}

	addrHex := hex.EncodeToString(hb)

	balance, err := db.Balances(addrHex)
	if err != nil {
		return false, err
	}

	status := false

	for _, bal := range balance {
		if bal.Denom != denom.Name {
			continue
		}

		parsedAmt, err := types.ParseCoinNormalized(bal.Amount)
		if err != nil {
			return false, fmt.Errorf("found relayeramount denom but failed to parse amount, %w", err)
		}

		status = parsedAmt.Amount.Int64() >= *denom.MinimumThreshRelayerBalance
		break
	}

	return status, nil
}
