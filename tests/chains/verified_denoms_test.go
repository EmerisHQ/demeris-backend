package tests

import (
	"encoding/json"

	"github.com/allinbits/demeris-backend-models/cns"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	verifiedDenomsEndpoint = "/verified_denoms"
)

func (suite *testCtx) TestVerifiedDenoms() {
	suite.T().Parallel()

	var chainsDenoms cns.DenomList
	for _, ch := range suite.Chains {
		if ch.Enabled {
			var payload map[string]interface{}
			err := json.Unmarshal(ch.Payload, &payload)
			suite.NoError(err)

			data, err := json.Marshal(payload["denoms"])
			suite.NoError(err)

			var expectedDenoms cns.DenomList
			err = json.Unmarshal(data, &expectedDenoms)
			suite.NoError(err)

			for _, denom := range expectedDenoms {
				if denom.Verified {
					chainsDenoms = append(chainsDenoms, denom)
				}
			}
		}
	}

	// arrange
	url := suite.Client.BuildUrl(verifiedDenomsEndpoint)
	// act
	resp, err := suite.Client.Get(url)
	suite.NoError(err)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, suite.T())

	defer resp.Body.Close()

	data, err := json.Marshal(respValues["verified_denoms"])
	suite.NoError(err)

	var denoms cns.DenomList
	err = json.Unmarshal(data, &denoms)
	suite.NoError(err)
	suite.NotNil(denoms)

	suite.Equal(len(chainsDenoms), len(denoms))

	suite.ElementsMatch(chainsDenoms, denoms)
}
