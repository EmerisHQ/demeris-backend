package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	baseUrl               = "%s://%s%s"
	liquidityNodeEndpoint = "/liquidity/node_info"
	chainName             = "cosmos-hub"
)

func TestLiquidityStatus(t *testing.T) {
	t.Parallel()

	// arrange
	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	require.NotNil(t, emIngress)

	chains := utils.LoadChainsInfo(env, t)
	require.NotNil(t, chains)

	client := utils.CreateNetClient(env, t)
	require.NotNil(t, client)

	for _, ch := range chains {
		if ch.Name == chainName {
			t.Run(ch.Name, func(t *testing.T) {

				// arrange
				url := fmt.Sprintf(baseUrl+liquidityNodeEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
				// act
				resp, err := client.Get(url)
				require.NoError(t, err)
				
				defer resp.Body.Close()

				// assert
				require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var values map[string]interface{}
				utils.RespBodyToMap(resp.Body, &values, t)

				v, ok := values["node_info"].(map[string]interface{})
				require.True(t, ok)
				networkName := v["network"]

				var fileResp map[string]interface{}
				utils.StringToMap(ch.Payload, &fileResp, t)

				fv, ok := fileResp["node_info"].(map[string]interface{})
				require.True(t, ok)

				expectedName := fv["chain_id"]

				require.Equal(t, expectedName, networkName)
			})
		}
	}
}
