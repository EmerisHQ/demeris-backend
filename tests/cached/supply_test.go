package tests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/stretchr/testify/require"
)

const (
	cachedSupplyEndPoint = "cached/cosmos/bank/v1beta1/supply"
	supplyEndPoint       = "liquidity/cosmos/bank/v1beta1/supply"
)

func TestCachedSupply(t *testing.T) {
	t.Parallel()

	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	require.NotNil(t, emIngress)

	client := utils.CreateNetClient(env, t)
	require.NotNil(t, client)

	// get cached supply
	urlPattern := strings.Join([]string{baseUrl, cachedSupplyEndPoint}, "")

	url := fmt.Sprintf(urlPattern, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	cachedResp, err := client.Get(url)
	require.NoError(t, err)

	defer cachedResp.Body.Close()

	var cachedValues map[string]interface{}
	utils.RespBodyToMap(cachedResp.Body, &cachedValues, t)

	// get supply
	urlPattern = strings.Join([]string{baseUrl, supplyEndPoint}, "")
	url = fmt.Sprintf(urlPattern, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	supplyResp, err := client.Get(url)
	require.NoError(t, err)

	defer supplyResp.Body.Close()

	var supplyValues map[string]interface{}
	utils.RespBodyToMap(supplyResp.Body, &supplyValues, t)

	require.Equal(t, supplyValues["supply"], cachedValues["supply"])
}
