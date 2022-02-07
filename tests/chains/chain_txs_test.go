package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	chainClient "github.com/allinbits/demeris-backend/chain_client"
	utils "github.com/allinbits/demeris-backend/test_utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	chainTxsEndpoint = "chain/%s/txs/%s"
	randomTxHash     = "56FF608A76A01D9178039D17949F53ED8E3969752D546E5474605A67B13A42A0"
)

func TestTxsEndpoint(t *testing.T) {
	// t.Parallel()

	// arrange
	env := "dev"
	emIngress, _, err := utils.LoadIngressInfo(env)
	require.NoError(t, err)
	chains := utils.LoadClientChainsInfo(env, t)
	client, err := utils.CreateNetClient(env)
	require.NoError(t, err)

	for _, ch := range chains {
		t.Run(ch.Name, func(t *testing.T) {
			var cc chainClient.Client
			err := json.Unmarshal(ch.Payload, &cc)
			require.NoError(t, err)
			cli := chainClient.GetClient(t, env, ch.Name, cc)
			// assert
			if !cli.Enabled {
				// arrange
				url := fmt.Sprintf(baseUrl+chainTxsEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name, randomTxHash)
				// act
				resp, err := client.Get(url)
				require.NoError(t, err)
				require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
				err = resp.Body.Close()
				require.NoError(t, err)
			} else {
				account, err := cli.AccountGet(cc.Key)
				require.NoError(t, err)

				fromAddr, err := sdk.AccAddressFromBech32(account.Address)
				require.NoError(t, err)

				account2, err := cli.AccountCreate("key2", "", cli.HDPath)
				require.NoError(t, err)

				toAddr, err := sdk.AccAddressFromBech32(account2.Address)
				require.NoError(t, err)

				msg := banktypes.NewMsgSend(fromAddr, toAddr, sdk.NewCoins(sdk.NewCoin(cli.Denom, sdk.NewInt(100))))

				txRes, err := cli.Broadcast(cc.Key, context.Background(), cli.GetContext(), msg)
				require.NoError(t, err)

				hash := txRes.TxHash

				time.Sleep(time.Second * 8)

				// arrange
				url := fmt.Sprintf(baseUrl+chainTxsEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name, hash)
				// act
				resp, err := client.Get(url)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				err = resp.Body.Close()
				require.NoError(t, err)

				txhashMap := respValues["tx_response"].(map[string]interface{})
				require.Equal(t, hash, txhashMap["txhash"])
			}
		})
	}
}
