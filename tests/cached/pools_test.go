package tests

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

const (
	cachedPoolsEndPoint    = "cached/cosmos/liquidity/v1beta1/pools"
	liquidityPoolsEndPoint = "liquidity/cosmos/liquidity/v1beta1/pools"
)

func TestCachedPools(t *testing.T) {
	t.Parallel()

	// get cached pools
	urlPattern := strings.Join([]string{baseUrl, cachedPoolsEndPoint}, "")
	url := fmt.Sprintf(urlPattern, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath)
	cachedResp, err := testCtx.client.Get(url)
	require.NoError(t, err)

	defer cachedResp.Body.Close()

	var cachedValues liquiditytypes.QueryLiquidityPoolsResponse
	body, err := ioutil.ReadAll(cachedResp.Body)
	require.NoError(t, err)

	require.NoError(t, tmjson.Unmarshal(body, &cachedValues))

	// get liquidity pools
	urlPattern = strings.Join([]string{baseUrl, liquidityPoolsEndPoint}, "")

	url = fmt.Sprintf(urlPattern, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath)
	liquidityResp, err := testCtx.client.Get(url)
	require.NoError(t, err)

	defer liquidityResp.Body.Close()

	var liquidityValues liquiditytypes.QueryLiquidityPoolsResponse
	body, err = ioutil.ReadAll(liquidityResp.Body)
	require.NoError(t, err)

	require.NoError(t, tmjson.Unmarshal(body, &liquidityValues))

	require.Equal(t, liquidityValues.Pools, cachedValues.Pools)
}
