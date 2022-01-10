package client

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/stretchr/testify/require"
)

const (
	getBalanceEndpoint = "/account/%s/balance"
	baseUrl            = "%s://%s%s"
)

func (c Client) TestGetBalanceOfAnyAccount(t *testing.T) {
	t.Parallel()

	//env := os.Getenv("ENV")
	env := "dev"
	emIngress, _ := utils.LoadIngressInfo(env, t)
	client := utils.CreateNetClient(env, t)

	list, err := c.GetkeysList()
	require.NoError(t, err)
	require.NotNil(t, list)

	chains := utils.LoadChainsInfo(env, t)

	for _, ch := range chains {
		t.Run(ch.Name, func(t *testing.T) {
			for _, k := range list {
				url := fmt.Sprintf(baseUrl+getBalanceEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, k.GetAddress().String())

				log.Println("urlllllllllll..........", url)

				resp, err := client.Get(url)
				require.NoError(t, err)

				if !ch.Enabled {
					require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
				} else {
					require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

					var respValues map[string]interface{}
					utils.RespBodyToMap(resp.Body, &respValues, t)

					err = resp.Body.Close()
					require.NoError(t, err)

					require.NotNil(t, respValues)
				}
			}
		})
	}
}
