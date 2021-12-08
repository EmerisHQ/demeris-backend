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
	baseUrl        = "%s://%s%s"
	statusEndpoint = "chain/%s/status"
	onlineKey      = "online"
)

func TestChainStatus(t *testing.T) {
	t.Parallel()

	// arrange
	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	chains := utils.LoadChainsInfo(env, t)
	client := utils.CreateNetClient(env, t)

	for _, ch := range chains {
		t.Run(ch.Name, func(t *testing.T) {
			t.Parallel()

			// arrange
			url := fmt.Sprintf(baseUrl+statusEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name)
			// act
			resp, err := client.Get(url)
			require.NoError(t, err)

			// assert
			require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

			var values map[string]interface{}
			utils.RespBodyToMap(resp.Body, &values, t)

			require.Equal(t, ch.Enabled, values[onlineKey].(bool), fmt.Sprintf("Chain %s Online %t", ch.Name, values[onlineKey].(bool)))
		})
	}
}
