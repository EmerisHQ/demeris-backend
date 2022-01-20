package tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
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

	var cachedValues liquiditytypes.QueryLiquidityPoolsResponse
	body, err := ioutil.ReadAll(cachedResp.Body)
	require.NoError(t, err)

	require.NoError(t, tmjson.Unmarshal(body, &cachedValues))

	// get liquidity pools
	url = fmt.Sprintf(baseUrl+liquidityPoolsEndPoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	liquidityResp, err := client.Get(url)
	require.NoError(t, err)

	defer liquidityResp.Body.Close()

	var liquidityValues liquiditytypes.QueryLiquidityPoolsResponse
	body, err = ioutil.ReadAll(liquidityResp.Body)
	require.NoError(t, err)

	require.NoError(t, tmjson.Unmarshal(body, &liquidityValues))

	require.Equal(t, liquidityValues.Pools, cachedValues.Pools)
}
