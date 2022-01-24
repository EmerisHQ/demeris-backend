package tests

import (
	"fmt"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/stretchr/testify/require"
)

const (
	poolsEndpoint = "liquidity/cosmos/liquidity/v1beta1/pools"
)

func TestPoolsData(t *testing.T) {
	t.Parallel()

	url := fmt.Sprintf(baseUrl+poolsEndpoint, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath)

	resp, err := testCtx.client.Get(url)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, t)

	err = resp.Body.Close()
	require.NoError(t, err)

	require.NotNil(t, respValues)

	pools, _ := respValues["pools"]
	require.NotEmpty(t, pools)
}
