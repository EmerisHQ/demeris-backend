package relayer

import (
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/allinbits/demeris-backend/models"

	"github.com/allinbits/demeris-backend/api/database"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/allinbits/demeris-backend/utils/k8s"
	v1 "github.com/allinbits/starport-operator/api/v1"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	rel := router.Group("/relayer")

	rel.GET("/status", getRelayerStatus)
	rel.GET("/balance", getRelayerBalance)
}

// getRelayerStatus returns status of relayer.
// @Summary Gets relayer  status
// @Tags Relayer
// @ID relayer-status
// @Description gets relayer status
// @Produce json
// @Success 200 {object} relayerStatusResponse
// @Failure 500,403 {object} deps.Error
// @Router /relayer/status [get]
func getRelayerStatus(c *gin.Context) {
	var res relayerStatusResponse

	d := deps.GetDeps(c)

	running, err := k8s.Querier{
		Client:    *d.K8S,
		Namespace: d.KubeNamespace,
	}.Relayer()

	if err != nil && !errors.Is(err, k8s.ErrNotFound) {
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

	res.Running = true

	if errors.Is(err, k8s.ErrNotFound) || running.Status.Phase != v1.RelayerPhaseRunning {
		res.Running = false
	}

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

	running, err := k8s.Querier{
		Client:    *d.K8S,
		Namespace: d.KubeNamespace,
	}.Relayer()

	if err != nil && !errors.Is(err, k8s.ErrNotFound) {
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

	chains := []string{}
	addresses := []string{}

	for _, cs := range running.Status.ChainStatuses {
		chains = append(chains, cs.ID)
		addresses = append(addresses, cs.AccountAddress)
	}

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

		rThresStr := fmt.Sprintf("%v%s", *denom.MinimumThreshRelayerBalance, parsedAmt.Denom)
		rThresAmt, err := types.ParseCoinNormalized(rThresStr)
		if err != nil {
			return false, fmt.Errorf("cannot ParseCoinNormalized() %s", rThresStr)
		}

		status = parsedAmt.IsGTE(rThresAmt)
		break
	}

	return status, nil
}
