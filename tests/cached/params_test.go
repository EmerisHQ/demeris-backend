package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
	"github.com/stretchr/testify/require"
)

const (
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
	urlPattern := strings.Join([]string{baseUrl, cachedParamsEndPoint}, "")

	url := fmt.Sprintf(urlPattern, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	cachedResp, err := client.Get(url)
	require.NoError(t, err)

	defer cachedResp.Body.Close()

	var cachedValues liquiditytypes.QueryParamsResponse
	body, err := ioutil.ReadAll(cachedResp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &cachedValues)
	require.NoError(t, err)

	// get liquidity params
	urlPattern = strings.Join([]string{baseUrl, liquidityParamsEndPoint}, "")

	url = fmt.Sprintf(urlPattern, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	liquidityResp, err := client.Get(url)
	require.NoError(t, err)

	defer liquidityResp.Body.Close()

	var liquidityValues liquiditytypes.QueryParamsResponse
	body, err = ioutil.ReadAll(liquidityResp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &liquidityValues)
	require.NoError(t, err)

	require.Equal(t, liquidityValues.Params, cachedValues.Params)
}
