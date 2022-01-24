package tests

import (
	"encoding/json"
	"fmt"
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

	var chainsDenoms cns.DenomList
	for _, ch := range testCtx.chains {
		if ch.Enabled {
			var payload map[string]interface{}
			err := json.Unmarshal(ch.Payload, &payload)
			require.NoError(t, err)

			data, err := json.Marshal(payload["denoms"])
			require.NoError(t, err)

			var expectedDenoms cns.DenomList
			err = json.Unmarshal(data, &expectedDenoms)
			require.NoError(t, err)

			for _, denom := range expectedDenoms {
				if denom.Verified {
					chainsDenoms = append(chainsDenoms, denom)
				}
			}
		}
	}

	// arrange
	url := fmt.Sprintf(baseUrl+verifiedDenomsEndpoint, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath)
	// act
	resp, err := testCtx.client.Get(url)
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

	require.ElementsMatch(t, chainsDenoms, denoms)
}
