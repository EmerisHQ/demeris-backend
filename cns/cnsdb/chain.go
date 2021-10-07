package cnsdb

import (
	"github.com/allinbits/demeris-backend/models"
)

// VerifiedTokens returns a DenomList of native denoms that are verified.
func (c Chain) VerifiedTokens() models.DenomList {
	var ret models.DenomList
	for _, ft := range c.Denoms {
		if !ft.Verified {
			continue
		}

		ret = append(ret, ft)
	}

	return ret
}

// FeeTokens returns a DenomList of denoms that are usable as fee.
func (c Chain) FeeTokens() models.DenomList {
	var ret models.DenomList
	for _, ft := range c.Denoms {
		if !ft.FeeToken {
			continue
		}

		ret = append(ret, ft)
	}

	return ret
}

// RelayerToken returns the relayer token for a given chain.
func (c Chain) RelayerToken() models.Denom {
	for _, ft := range c.Denoms {
		if ft.RelayerDenom {
			return ft
		}
	}

	panic("relayer token not defined")
}
