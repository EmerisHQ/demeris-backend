package tx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.POST("/tx/:chain", Tx)
}

// Tx relays a transaction to an internal node for the specified chain.
// @Summary Relays a transaction to the relevant chain.
// @Tags Tx
// @ID tx
// @Description Relays a transaction to the relevant chain.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} TxResponse
// @Failure 500,403 {object} deps.Error
// @Router /tx/{chainName} [post]
func Tx(c *gin.Context) {
	var tx TxData
	var meta TxMeta

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	err := c.BindJSON(&tx)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("failed to parse JSON"), http.StatusBadRequest)

		d.WriteError(c, e,
			"Failed to parse JSON",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	meta.Chain, err = d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("chain %s does not exist", chainName), http.StatusBadRequest)

		d.WriteError(c, e,
			"Invalid chain",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	err = validateTx(&tx, &meta, d)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("invalid transaction"), http.StatusBadRequest)

		d.WriteError(c, e,
			"invalid transaction",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	res, err := RelayTx(tx, meta)

	if err != nil {
		e := deps.NewError("tx", fmt.Errorf("relaying tx failed"), http.StatusBadRequest)

		d.WriteError(c, e,
			"relaying tx failed",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, res)
}

// validateTx populates metadata and
func validateTx(tx *TxData, meta *TxMeta, d *deps.Deps) error {

	err := validateSignatures(tx, meta, d)

	if err != nil {
		return err
	}

	err = validateBody(tx, meta, d)

	if err != nil {
		return err
	}

	// TODO: Fetch sequence for ticketing system
	err = validateAuthInfo(tx, meta, d)

	if err != nil {
		return err
	}

	return nil
}

// validateBody validates the data inside auth_info and populates the relevant metadata
func validateBody(tx *TxData, meta *TxMeta, d *deps.Deps) error {

	// check the body of the message

	for _, msg := range tx.TxBody["messages"].([]interface{}) {
		m := msg.(map[string]interface{})
		if msgType, _ := m["@type"]; msgType == "/ibc.applications.transfer.v1.MsgTransfer" {
			sourcePort := m["source_port"].(string)
			sourceChannel := m["source_channel"].(string)

			tokenDenom := m["token"].(map[string]interface{})["denom"].(string)

			if sourcePort != "transfer" {
				return fmt.Errorf("Invalid IBC Port %s", sourcePort)
			}

			if "ibc/" == tokenDenom[:4] {
				tokenHash := tokenDenom[4:]
				denomTrace, err := d.Database.DenomTrace(meta.Chain.ChainName, tokenHash)
				if err != nil {
					return fmt.Errorf("Invalid denom trace")
				}

				channels := strings.Split(denomTrace.Path, "/transfer")
				if channels[0] != sourceChannel {
					return fmt.Errorf("IBC forward is disabled for multi-hop tokens. Try sending it back through the original channel.")
				}
			}
		}
	}

	return nil
}

// validateAuthInfo validates the data inside auth_info and populates the relevant metadata
func validateAuthInfo(tx *TxData, meta *TxMeta, d *deps.Deps) error {

	if infos := tx.AuthInfo["signer_infos"].([]interface{}); len(infos) == 1 {
		// Fetch signer sequence
		signerInfo := tx.AuthInfo["signer_infos"].([]interface{})[0].(map[string]interface{})
		meta.SignerSequence = signerInfo["sequence"].(string)
	} else {
		return fmt.Errorf("Invalid number of signatures. Expected 1, got %d", len(infos))
	}

	return nil
}

// validateSignatures ensures the signature exists
func validateSignatures(tx *TxData, meta *TxMeta, d *deps.Deps) error {
	if len(tx.Signatures) != 1 {
		return fmt.Errorf("Invalid number of signatures")
	}

	return nil
}

// RelayTx relays the tx to the specifc endpoint
// RelayTx will also perform the ticketing mechanism
func RelayTx(tx TxData, meta TxMeta) (TxResponse, error) {

	var res TxResponse

	data, err := json.Marshal(tx)

	if err != nil {
		return res, err
	}

	b := bytes.NewReader(data)

	_, err = http.Post(meta.Chain.NodeInfo.Endpoint, "application/json", b)

	if err != nil {
		return res, err
	}

	// todo: ticketing

	res.Key = ""
	res.Sequence = meta.SignerSequence
	res.Status = "ok"

	return res, nil
}
