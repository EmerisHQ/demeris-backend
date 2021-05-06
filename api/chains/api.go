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
	chain.GET("/denom/verify_trace/:hash", VerifyTrace)
	chain.GET("/bech32", GetChainBech32Config)
	chain.GET("/primary_channels", GetPrimaryChannels)
	chain.GET("/primary_channel/:counterparty", GetPrimaryChannelWithCounterparty)

	fee := chain.Group("/fee")

	fee.GET("", GetFee)
	fee.GET("/address", GetFeeAddress)
	fee.GET("/token", GetFeeToken)

}

// GetChains returns the list of all the chains supported by demeris.
// @Summary Gets list of supported chains.
// @Tags Chain
// @ID chains
// @Description Gets list of supported chains.
// @Produce json
// @Success 200 {object} chainsResponse
// @Failure 500,403 {object} deps.Error
// @Router /chains [get]
func GetChains(c *gin.Context) {
	var res chainsResponse

	d := deps.GetDeps(c)

	chains, err := d.Database.Chains()

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve chains"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chains",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	for _, cc := range chains {
		res.Chains = append(res.Chains, supportedChain{
			ChainName:   cc.ChainName,
			DisplayName: cc.DisplayName,
			Logo:        cc.Logo,
		})
	}

	c.JSON(http.StatusOK, res)
}

// GetChain returns chain information by specifying its name.
// @Summary Gets chain by name.
// @Tags Chain
// @ID chain
// @Description Gets chain by name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} chainResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName} [get]
func GetChain(c *gin.Context) {
	var res chainResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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
// @Summary Gets chain bech32 configuration by chain name.
// @Tags Chain
// @ID bech32config
// @Description Gets chain bech32 configuration by chain name..
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} bech32ConfigResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/bech32 [get]
func GetChainBech32Config(c *gin.Context) {
	var res bech32ConfigResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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

	res.Bech32Config = chain.NodeInfo.Bech32Config

	c.JSON(http.StatusOK, res)
}

// GetFee returns the fee average in dollar for the specified chain.
// @Summary Gets average fee in dollar by chain name.
// @Tags Chain
// @ID fee
// @Description Gets average fee in dollar by chain name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} feeResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/fee [get]
func GetFee(c *gin.Context) {
	var res feeResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"fee",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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
		Fee: chain.BaseFee,
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeAddress returns the fee address for a given chain, looked up by the chain name attribute.
// @Summary Gets address to pay fee for by chain name.
// @Tags Chain
// @ID feeaddress
// @Description Gets address to pay fee for by chain name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} feeAddressResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/address [get]
func GetFeeAddress(c *gin.Context) {
	var res feeAddressResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"feeaddress",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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
		FeeAddress: chain.FeeAddress,
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeAddresses returns the fee address for all chains.
// @Summary Gets all addresses to pay fee for.
// @Tags Chain
// @ID feeaddresses
// @Description Gets all addresses to pay fee for.
// @Produce json
// @Success 200 {object} feeAddressesResponse
// @Failure 500,403 {object} deps.Error
// @Router /chains/fee/addresses [get]
func GetFeeAddresses(c *gin.Context) {
	var res feeAddressesResponse

	d := deps.GetDeps(c)

	chains, err := d.Database.Chains()

	if err != nil {
		e := deps.NewError(
			"feeaddress",
			fmt.Errorf("cannot retrieve chains"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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
			feeAddress{
				ChainName:  c.ChainName,
				FeeAddress: c.FeeAddress,
			},
		)
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeToken returns the fee token for a given chain, looked up by the chain name attribute.
// @Summary Gets token used to pay fees by chain name.
// @Tags Chain
// @ID feetoken
// @Description Gets token used to pay fees by chain name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} feeTokenResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/token [get]
func GetFeeToken(c *gin.Context) {
	var res feeTokenResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"feetoken",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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
		res.FeeTokens = append(res.FeeTokens, cc)
	}

	c.JSON(http.StatusOK, res)
}

// GetPrimaryChannelWithCounterparty returns the primary channel of a chain by specifying the counterparty.
// @Summary Gets the channel name that connects two chains.
// @Tags Chain
// @ID counterparty
// @Description Gets the channel name that connects two chains.
// @Param chainName path string true "chain name"
// @Param counterparty path string true "counterparty chain name"
// @Produce json
// @Success 200 {object} primaryChannelResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/primary_channel/{counterparty} [get]
func GetPrimaryChannelWithCounterparty(c *gin.Context) {
	var res primaryChannelResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")
	counterparty := c.Param("counterparty")

	chain, err := d.Database.PrimaryChannelCounterparty(chainName, counterparty)

	if err != nil {
		e := deps.NewError(
			"primarychannel",
			fmt.Errorf("cannot retrieve primary channel between %v and %v", chainName, counterparty),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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

	res.Channel = primaryChannel{
		Counterparty: counterparty,
		ChannelName:  chain.ChannelName,
	}

	c.JSON(http.StatusOK, res)
}

// GetPrimaryChannels returns the primary channels of a chain.
// @Summary Gets the channel mapping of a chain with all the other chains it is connected to.
// @Tags Chain
// @ID channels
// @Description Gets the channel mapping of a chain with all the other chains it is connected to.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} primaryChannelsResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/primary_channel [get]
func GetPrimaryChannels(c *gin.Context) {
	var res primaryChannelsResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.PrimaryChannels(chainName)

	if err != nil {
		e := deps.NewError(
			"primarychannel",
			fmt.Errorf("cannot retrieve primary channels for %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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
		res.Channels = append(res.Channels, primaryChannel{
			Counterparty: cc.Counterparty,
			ChannelName:  cc.ChannelName,
		})
	}

	c.JSON(http.StatusOK, res)
}

// VerifyTrace verifies that a trace hash is valid against a chain name.
// @Summary Verifies that a trace hash is valid against a chain name.
// @Tags Chain
// @ID verifyTrace
// @Description Verifies that a trace hash is valid against a chain name.
// @Param chainName path string true "chain name"
// @Param hash path string true "trace hash"
// @Produce json
// @Success 200 {object} verifiedTraceResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/denom/verify_trace/{hash} [get]
func VerifyTrace(c *gin.Context) {
	var res verifiedTraceResponse

	d := deps.GetDeps(c)

	chain := c.Param("chain")
	hash := c.Param("hash")

	denomTrace, err := d.Database.DenomTrace(chain, hash)

	if err != nil {

		e := deps.NewError(
			"denom/verify-trace",
			fmt.Errorf("cannot query token hash %v on chain %v", hash, chain),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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

	res.VerifiedTrace.IbcDenom = fmt.Sprintf("ibc/%s", hash)
	res.VerifiedTrace.Path = denomTrace.Path

	// check if the path uses only the supported `transfer` port.

	channels := strings.Split(res.VerifiedTrace.Path, "/transfer")

	for idx, channel := range channels {
		ch := strings.Trim(channel, "/")

		// port other than transfer being used
		if strings.Contains(ch, "/") {

			err = errors.New(fmt.Sprintf("Unsupported path %s", res.VerifiedTrace.Path))

			e := deps.NewError(
				"denom/verify-trace",
				fmt.Errorf("invalid denom %v with path %v", hash, res.VerifiedTrace.Path),
				http.StatusBadRequest,
			)

			d.WriteError(c, e,
				"invalid denom",
				"id",
				e.ID,
				"hash",
				hash,
				"path",
				res.VerifiedTrace.Path,
				"err",
				err,
			)
		}

		channels[idx] = ch
	}

	var client models.IbcClientInfo
	var chainInfo models.Chain
	var trace trace

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

			d.WriteError(c, e,
				"invalid client id",
				"id",
				e.ID,
				"hash",
				hash,
				"path",
				res.VerifiedTrace.Path,
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

				d.WriteError(c, e,
					"invalid client id",
					"id",
					e.ID,
					"hash",
					hash,
					"path",
					res.VerifiedTrace.Path,
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

		res.VerifiedTrace.Trace = append(res.VerifiedTrace.Trace, trace)

	}

	c.JSON(http.StatusOK, res)
}
