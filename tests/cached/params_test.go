package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
	"github.com/stretchr/testify/require"
)

const (
	baseUrl                 = "%s://%s%s"
	cachedParamsEndPoint    = "cached/cosmos/liquidity/v1beta1/params"
	liquidityParamsEndPoint = "liquidity/cosmos/liquidity/v1beta1/params"
)

func TestCachedParams(t *testing.T) {
	t.Parallel()

	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	require.NotNil(t, emIngress)

	client := utils.CreateNetClient(env, t)
	require.NotNil(t, client)

	// get cached params
	url := fmt.Sprintf(baseUrl+cachedParamsEndPoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	cachedResp, err := client.Get(url)
	require.NoError(t, err)

	defer cachedResp.Body.Close()

	var cachedValues liquiditytypes.QueryParamsResponse
	body, err := ioutil.ReadAll(cachedResp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &cachedValues)
	require.NoError(t, err)

	// get liquidity params
	url = fmt.Sprintf(baseUrl+liquidityParamsEndPoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	liquidityResp, err := client.Get(url)
	require.NoError(t, err)

	defer liquidityResp.Body.Close()

	var liquidityValues liquiditytypes.QueryParamsResponse
	body, err = ioutil.ReadAll(liquidityResp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &liquidityValues)
	require.NoError(t, err)

	// adding missing feild in cached response
	cachedValues.Params.CircuitBreakerEnabled = liquidityValues.Params.CircuitBreakerEnabled

	require.Equal(t, liquidityValues.Params, cachedValues.Params)
}
