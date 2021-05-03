package chains

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/allinbits/demeris-backend/models"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/chains", GetChains)
	router.GET("/chains/fee/addresses", GetFeeAddresses)

	chain := router.Group("/chain/:chain")

	chain.GET("", GetChain)
	chain.GET("/denom/verify-trace/:hash", VerifyTrace)
	chain.GET("/bech32", GetChainBech32Config)
	chain.GET("/primary_channel", GetPrimaryChannels)
	chain.GET("/primary_channel/:counterparty", GetPrimaryChannels)

	fee := chain.Group("/fee")

	fee.GET("", GetFee)
	fee.GET("/address", GetFeeAddress)
	fee.GET("/token", GetFeeToken)

}

// GetChains returns the list of all the chains supported by demeris.
func GetChains(c *gin.Context) {
	var res chainsResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"chains",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chains, err := d.Database.Chains()

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve chains"),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chains",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	for _, cc := range chains {
		res.SupportedChains = append(res.SupportedChains, cc.ChainName)
	}

	c.JSON(http.StatusOK, res)
}

// GetChain returns chain information by specifying its name.
func GetChain(c *gin.Context) {
	var res chainResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"chains",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chain",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	res.Chain = chain

	c.JSON(http.StatusOK, res)
}

// GetChainBech32Config returns bech32 configuration for a chain by specifying its name.
func GetChainBech32Config(c *gin.Context) {
	var res bech32ConfigResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"chains",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chain",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	res.ChainName = chain.ChainName
	res.Bech32Config = chain.NodeInfo.Bech32Config

	c.JSON(http.StatusOK, res)
}

// GetFee returns the fee average in dollar for the specified chain..
func GetFee(c *gin.Context) {
	var res feeResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"fee",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"fee",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chain",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	res = feeResponse{
		ChainName: chain.ChainName,
		Fee:       chain.BaseFee,
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeAddress returns the fee address for a given chain, looked up by the chain name attribute.
func GetFeeAddress(c *gin.Context) {
	var res feeAddressResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"feeaddress",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"feeaddress",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chain",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	res = feeAddressResponse{
		ChainName:  chain.ChainName,
		FeeAddress: chain.FeeAddress,
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeAddresses returns the fee address for all chains.
func GetFeeAddresses(c *gin.Context) {
	var res feeAddressesResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"feeaddress",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chains, err := d.Database.Chains()

	if err != nil {
		e := deps.NewError(
			"feeaddress",
			fmt.Errorf("cannot retrieve chains"),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chains",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	for _, c := range chains {
		res.FeeAddresses = append(
			res.FeeAddresses,
			feeAddressResponse{
				ChainName:  c.ChainName,
				FeeAddress: c.FeeAddress,
			},
		)
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeToken returns the fee token for a given chain, looked up by the chain name attribute.
func GetFeeToken(c *gin.Context) {
	var res feeTokenResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"feetoken",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"feetoken",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chain",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	for _, cc := range chain.VerifiedFeeTokens() {
		res.FeeTokens = append(res.FeeTokens, FeeToken{
			Name:     cc.Name,
			Verified: cc.Verified,
		})
	}

	c.JSON(http.StatusOK, res)
}

// GetPrimaryChannelWithCounterparty returns the primary channel of a chain by specifying the counterparty.
func GetPrimaryChannelWithCounterparty(c *gin.Context) {
	var res counterpartyResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"primarychannel",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chainName := c.Param("chain")
	counterparty := c.Param("counterparty")

	chain, err := d.Database.PrimaryChannelCounterparty(chainName, counterparty)

	if err != nil {
		e := deps.NewError(
			"primarychannel",
			fmt.Errorf("cannot retrieve primary channel between %v and %v", chainName, counterparty),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chain",
			"id",
			e.ID,
			"name",
			chainName,
			"counterparty",
			counterparty,
			"error",
			err,
		)

		return
	}

	res = counterpartyResponse{
		ChainName:    chainName,
		Counterparty: counterparty,
		ChannelName:  chain.ChannelName,
	}

	c.JSON(http.StatusOK, res)
}

// GetPrimaryChannels returns the primary channels of a chain.
func GetPrimaryChannels(c *gin.Context) {
	var res channelsResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"primarychannel",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chainName := c.Param("chain")

	chain, err := d.Database.PrimaryChannels(chainName)

	if err != nil {
		e := deps.NewError(
			"primarychannel",
			fmt.Errorf("cannot retrieve primary channels for %v", chainName),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chain",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	for _, cc := range chain {
		res.Channels = append(res.Channels, counterpartyResponse{
			ChainName:    chainName,
			Counterparty: cc.Counterparty,
			ChannelName:  cc.ChannelName,
		})
	}

	c.JSON(http.StatusOK, res)
}

func VerifyTrace(c *gin.Context) {
	var res VerifiedTraceResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"denom/verify-trace",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chain := c.Param("chain")
	hash := c.Param("hash")

	denomTrace, err := d.Database.DenomTrace(chain, hash)

	if err != nil {

		e := deps.NewError(
			"denom/verify-trace",
			fmt.Errorf("cannot query token hash %v on chain %v", hash, chain),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot query database for denom",
			"id",
			e.ID,
			"hash",
			hash,
			"chain",
			chain,
			"error",
			err,
		)

		return

	}

	res.IbcDenom = fmt.Sprintf("ibc/%s", hash)
	res.Path = denomTrace.Path

	// check if the path uses only the supported `transfer` port.

	channels := strings.Split(res.Path, "/transfer")

	for idx, channel := range channels {
		ch := strings.Trim(channel, "/")

		// port other than transfer being used
		if strings.Contains(ch, "/") {

			err = errors.New(fmt.Sprintf("Unsupported path %s", res.Path))

			e := deps.NewError(
				"denom/verify-trace",
				fmt.Errorf("invalid denom %v with path %v", hash, res.Path),
				http.StatusBadRequest,
			)

			c.Error(e)

			d.Logger.Errorw(
				"invalid denom",
				"id",
				e.ID,
				"hash",
				hash,
				"path",
				res.Path,
				"err",
				err,
			)
		}

		channels[idx] = ch
	}

	var client models.IbcClientInfo
	var chainInfo models.Chain
	var trace Trace

	for _, channel := range channels {

		client, _ = d.Database.QueryIBCClientTrace(chain, channel)

		chainInfo, _ = d.Database.Chain(chain)

		if counterparty := chainInfo.CounterpartyNames[client.ChannelId]; counterparty == "" {
			err = errors.New(fmt.Sprintf("Unsupported client id %s on chain %s", client.ChannelId, chain))

			e := deps.NewError(
				"denom/verify-trace",
				fmt.Errorf("Unsupported client id when resolving path for %s", hash),
				http.StatusBadRequest,
			)

			c.Error(e)

			d.Logger.Errorw(
				"invalid client id",
				"id",
				e.ID,
				"hash",
				hash,
				"path",
				res.Path,
				"client_id",
				client.ClientId,
				"chain",
				chain,
				"err",
				err,
			)
		} else {
			trace.ChainName = chain
			trace.Channel = client.ChannelId
			trace.ClientId = client.ClientId
			trace.CounterpartyName = counterparty

			// query counterparty chain name
			counterpartyConn, _ := d.Database.Connection(chain, client.CounterConnectionID)

			if counterpartyConn.CounterClientID != trace.ClientId {
				err = errors.New("Client ids do not match")

				e := deps.NewError(
					"denom/verify-trace",
					fmt.Errorf("Client ids do not match"),
					http.StatusBadRequest,
				)

				c.Error(e)

				d.Logger.Errorw(
					"invalid client id",
					"id",
					e.ID,
					"hash",
					hash,
					"path",
					res.Path,
					"client_id",
					client.ClientId,
					"chain",
					chain,
					"counter_client_id",
					counterpartyConn.CounterClientID,
					"counter_chain",
					counterparty,
					"err",
					err,
				)
			}
		}

		res.Trace = append(res.Trace, trace)
	}

	c.JSON(http.StatusOK, gin.H{
		"trace": res,
	})
}
