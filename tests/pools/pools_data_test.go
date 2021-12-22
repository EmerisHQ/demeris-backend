package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/stretchr/testify/require"
)

const (
	poolsEndpoint = "liquidity/cosmos/liquidity/v1beta1/pools"
	baseUrl       = "%s://%s%s"
)

func TestPoolsData(t *testing.T) {
	t.Parallel()

	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	client := utils.CreateNetClient(env, t)

	url := fmt.Sprintf(baseUrl+poolsEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)

	resp, err := client.Get(url)
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
