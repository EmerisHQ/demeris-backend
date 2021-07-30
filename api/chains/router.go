package chains

import "github.com/gin-gonic/gin"

const grpcPort = 9090

func Register(router *gin.Engine) {
	router.GET("/chains", GetChains)
	router.GET("/chains/fee/addresses", GetFeeAddresses)

	chain := router.Group("/chain/:chain")

	chain.GET("", GetChain)
	chain.GET("/denom/verify_trace/:hash", VerifyTrace)
	chain.GET("/bech32", GetChainBech32Config)
	chain.GET("/primary_channels", GetPrimaryChannels)
	chain.GET("/primary_channel/:counterparty", GetPrimaryChannelWithCounterparty)
	chain.GET("/status", GetChainStatus)
	chain.GET("/supply", GetChainSupply)
	chain.GET("/:tx", GetChainTx)

	fee := chain.Group("/fee")

	fee.GET("", GetFee)
	fee.GET("/address", GetFeeAddress)
	fee.GET("/token", GetFeeToken)

}
