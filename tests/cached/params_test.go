package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
	"github.com/stretchr/testify/require"
)

const (
	cachedParamsEndPoint    = "cached/cosmos/liquidity/v1beta1/params"
	liquidityParamsEndPoint = "liquidity/cosmos/liquidity/v1beta1/params"
)

func TestCachedParams(t *testing.T) {
	t.Parallel()

	// get cached params
	urlPattern := strings.Join([]string{baseUrl, cachedParamsEndPoint}, "")

	url := fmt.Sprintf(urlPattern, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath)
	cachedResp, err := testCtx.client.Get(url)
	require.NoError(t, err)

	defer cachedResp.Body.Close()

	var cachedValues liquiditytypes.QueryParamsResponse
	body, err := ioutil.ReadAll(cachedResp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &cachedValues)
	require.NoError(t, err)

	// get liquidity params
	urlPattern = strings.Join([]string{baseUrl, liquidityParamsEndPoint}, "")

	url = fmt.Sprintf(urlPattern, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath)
	liquidityResp, err := testCtx.client.Get(url)
	require.NoError(t, err)

	defer liquidityResp.Body.Close()

	var liquidityValues liquiditytypes.QueryParamsResponse
	body, err = ioutil.ReadAll(liquidityResp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &liquidityValues)
	require.NoError(t, err)

	require.Equal(t, liquidityValues.Params, cachedValues.Params)
}
