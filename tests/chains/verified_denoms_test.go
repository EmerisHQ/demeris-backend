package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/allinbits/demeris-backend-models/cns"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	verifiedDenomsEndpoint = "/verified_denoms"
)

func TestVerifiedDenoms(t *testing.T) {
	t.Parallel()

	// arrange
	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	require.NotNil(t, emIngress)

	chains := utils.LoadChainsInfo(env, t)
	require.NotNil(t, chains)

	client := utils.CreateNetClient(env, t)
	require.NotNil(t, client)

	var chainsDenoms []cns.DenomList
	for _, ch := range chains {
		var payload map[string]interface{}
		err := json.Unmarshal(ch.Payload, &payload)
		require.NoError(t, err)

		data, err := json.Marshal(payload["denoms"])
		require.NoError(t, err)

		var expectedDenoms cns.DenomList
		err = json.Unmarshal(data, &expectedDenoms)
		require.NoError(t, err)
		chainsDenoms = append(chainsDenoms, expectedDenoms)
	}

	// arrange
	url := fmt.Sprintf(baseUrl+verifiedDenomsEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	// act
	resp, err := client.Get(url)
	require.NoError(t, err)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, t)

	defer resp.Body.Close()

	data, err := json.Marshal(respValues["verified_denoms"])
	require.NoError(t, err)

	var denoms cns.DenomList
	err = json.Unmarshal(data, &denoms)
	require.NoError(t, err)
	require.NotNil(t, denoms)

	require.Equal(t, len(chainsDenoms), len(denoms))

}
