package tests

import (
	"fmt"
	"os"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/stretchr/testify/require"
)

const (
	baseUrl                = "%s://%s%s"
	cachedPoolsEndPoint    = "cached/cosmos/liquidity/v1beta1/pools"
	liquidityPoolsEndPoint = "liquidity/cosmos/liquidity/v1beta1/pools"
)

func TestCachedPools(t *testing.T) {
	t.Parallel()

	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	require.NotNil(t, emIngress)

	client := utils.CreateNetClient(env, t)
	require.NotNil(t, client)

	// get cached pools
	url := fmt.Sprintf(baseUrl+cachedPoolsEndPoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	cachedResp, err := client.Get(url)
	require.NoError(t, err)

	defer cachedResp.Body.Close()

	var cachedValues map[string]interface{}
	utils.RespBodyToMap(cachedResp.Body, &cachedValues, t)

	// get liquidity pools
	url = fmt.Sprintf(baseUrl+liquidityPoolsEndPoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	liquidityResp, err := client.Get(url)
	require.NoError(t, err)

	defer liquidityResp.Body.Close()

	var liquidityValues map[string]interface{}
	utils.RespBodyToMap(liquidityResp.Body, &liquidityValues, t)

	// Fix needed: type mistach of pool id (uint64 in cached resp and string in liquidity resp)

	require.Equal(t, liquidityValues["pools"], cachedValues["pools"])
}
