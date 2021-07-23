package block

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/allinbits/demeris-backend/utils/store"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/block_results", GetBlock)
}

// GetBlock returns a Tendermint block data at a given height.
// @Summary Returns block data at a given height.
// @Tags Block
// @ID get-block
// @Description returns block data at a given height
// @Produce json
// @Param height query string true "height to query for"
// @Success 200 {object} blockHeightResp
// @Failure 500,403 {object} deps.Error
// @Router /block [get]
func GetBlock(c *gin.Context) {
	d := deps.GetDeps(c)

	h := c.Query("height")
	if h == "" {
		e := deps.NewError(
			"block",
			fmt.Errorf("missing height"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query block, missing height",
			"id",
			e.ID,
		)
		return
	}

	hh, err := strconv.ParseInt(h, 10, 64)
	if err != nil {
		e := deps.NewError(
			"block",
			fmt.Errorf("malformed height"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query block, malformed height",
			"id",
			e.ID,
			"height_string",
			h,
			"error",
			err,
		)
		return
	}

	bs := store.NewBlocks(d.Store)

	bd, err := bs.Block(hh)
	if err != nil {
		e := deps.NewError(
			"block",
			fmt.Errorf("cannot get block at height %v", hh),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query block from redis",
			"id",
			e.ID,
			"height",
			hh,
			"error",
			err,
		)
		return
	}

	c.Data(http.StatusOK, "application/json", bd)
}
