package tests

import (
	"encoding/json"

	"github.com/emerishq/demeris-backend-models/cns"
	utils "github.com/emerishq/demeris-backend/test_utils"
)

const (
	verifiedDenomsEndpoint = "verified_denoms"
)

func (suite *testCtx) TestVerifiedDenoms() {
	// have removed gravity-1 and ion for cosmos and osmosis from chains json files. but getting chains denoms count as 6 and tests is failing.
	// we have removed those denoms bcz got error while testing chain fee.
	var chainsDenoms cns.DenomList
	for _, ch := range suite.Chains {
		if ch.Enabled {
			expectedDenoms := ch.Denoms

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
