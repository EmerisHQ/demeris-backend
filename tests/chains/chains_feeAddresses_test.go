package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainsFeeAddressesEndpoint = "chains/fee/addresses"

func TestChainsFeeAddresses(t *testing.T) {
	t.Parallel()

	// arrange
	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	chains := utils.LoadChainsInfo(env, t)
	client := utils.CreateNetClient(env, t)

	// arrange
	url := fmt.Sprintf(baseUrl+chainsFeeAddressesEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	// act
	resp, err := client.Get(url)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, t)

	err = resp.Body.Close()
	require.NoError(t, err)

	expValues := make(map[string][]map[string]interface{}, 0)
	for _, ch := range chains {
		if ch.Enabled {
			expValues["fee_addresses"] = append(expValues["fee_addresses"], map[string]interface{}{
				"chain_name":  ch.Name,
				"fee_address": []string{"feeaddress"},
			})
		}
	}

	expValuesData, err := json.Marshal(expValues)
	require.NoError(t, err)

	var expValuesInterface map[string]interface{}
	err = json.Unmarshal(expValuesData, &expValuesInterface)
	require.NoError(t, err)

	require.Equal(t, expValuesInterface, respValues)
}
