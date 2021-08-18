package chains

import (
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
)

const grpcPort = 9090

func Register(router *gin.Engine, store *persistence.InMemoryStore) {
	router.GET("/chains", cache.CachePage(store, 10*time.Second, GetChains))
	router.GET("/chains/fee/addresses", cache.CachePage(store, 10*time.Second, GetFeeAddresses))

	chain := router.Group("/chain/:chain")

	chain.GET("", cache.CachePage(store, 10*time.Second, GetChain))
	chain.GET("/denom/verify_trace/:hash", cache.CachePage(store, 10*time.Second, VerifyTrace))
	chain.GET("/bech32", cache.CachePage(store, 10*time.Second, GetChainBech32Config))
	chain.GET("/primary_channels", cache.CachePage(store, 10*time.Second, GetPrimaryChannels))
	chain.GET("/primary_channel/:counterparty", cache.CachePage(store, 10*time.Second, GetPrimaryChannelWithCounterparty))
	chain.GET("/status", cache.CachePage(store, 10*time.Second, GetChainStatus))
	chain.GET("/supply", cache.CachePage(store, 10*time.Second, GetChainSupply))
	chain.GET("/txs/:tx", GetChainTx)
	chain.GET("/numbers/:address", GetNumbersByAddress)

	fee := chain.Group("/fee")

	fee.GET("", GetFee)
	fee.GET("/address", GetFeeAddress)
	fee.GET("/token", cache.CachePage(store, 10*time.Second, GetFeeToken))
}
