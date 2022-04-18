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
	var chainsDenoms cns.DenomList
	for _, ch := range suite.Chains {
		if ch.Enabled {
			var payload map[string]interface{}
			err := json.Unmarshal(ch.Payload, &payload)
			suite.Require().NoError(err)

			data, err := json.Marshal(payload["denoms"])
			suite.Require().NoError(err)

			var expectedDenoms cns.DenomList
			err = json.Unmarshal(data, &expectedDenoms)
			suite.Require().NoError(err)

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
	suite.Require().NoError(err)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, suite.T())

	defer resp.Body.Close()

	data, err := json.Marshal(respValues["verified_denoms"])
	suite.Require().NoError(err)

	var denoms cns.DenomList
	err = json.Unmarshal(data, &denoms)
	suite.Require().NoError(err)
	suite.Require().NotNil(denoms)

	suite.Require().Equal(len(chainsDenoms), len(denoms))

	suite.Require().ElementsMatch(chainsDenoms, denoms)
}
