package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/allinbits/demeris-backend/test_utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	chainSupplyEndpoint = "chain/%s/supply"
	supplyKey           = "supply"
)

func TestChainSupply(t *testing.T) {
	t.Parallel()

	// arrange
	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	chains := utils.LoadChainsInfo(env, t)
	client := utils.CreateNetClient(env, t)

	for _, ch := range chains {
		t.Run(ch.Name, func(t *testing.T) {

			// arrange
			url := fmt.Sprintf(baseUrl+chainSupplyEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name)
			// act
			resp, err := client.Get(url)
			require.NoError(t, err)

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				data, err := json.Marshal(respValues[supplyKey])
				require.NoError(t, err)

				var coins sdk.Coins
				err = json.Unmarshal(data, &coins)
				require.NoError(t, err)

				//check if the repsonse is empty
				require.NotEmpty(t, coins)
			}
		})
	}
}
