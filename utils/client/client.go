package client

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vechain/networkhub/utils/client/http"
	"github.com/vechain/networkhub/utils/common"
	"github.com/vechain/thor/v2/api/accounts"
	"github.com/vechain/thor/v2/api/blocks"
	"github.com/vechain/thor/v2/api/events"
	"github.com/vechain/thor/v2/api/node"
	"github.com/vechain/thor/v2/api/transactions"
	"github.com/vechain/thor/v2/api/transfers"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/tx"
)

type Client struct {
	conn *http.Client
}

func NewClient(url string) *Client {
	// todo, depending on the url return a http or a ws client
	return &Client{
		conn: http.NewClient(url),
	}
}

func (c *Client) RawClient() *http.Client {
	return c.conn
}
func (c *Client) GetTransactionReceipt(id *thor.Bytes32) (*transactions.Receipt, error) {
	return c.conn.GetTransactionReceipt(id)
}

func (c *Client) InspectClauses(calldata *accounts.BatchCallData) ([]*accounts.CallResult, error) {
	return c.conn.InspectClauses(calldata)
}

func (c *Client) InspectTxClauses(tx *tx.Transaction, senderAddr *thor.Address) ([]*accounts.CallResult, error) {
	clauses := convertToBatchCallData(tx, senderAddr)
	return c.InspectClauses(clauses)
}

func (c *Client) SendTransaction(tx *tx.Transaction) (*common.TxSendResult, error) {
	rlpTx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, fmt.Errorf("unable to encode transaction - %w", err)
	}

	return c.SendEncodedTransaction(rlpTx)
}

func (c *Client) SendEncodedTransaction(rlpTx []byte) (*common.TxSendResult, error) {
	return c.conn.SendTransaction(&transactions.RawTx{Raw: hexutil.Encode(rlpTx)})
}

func (c *Client) GetLogEvents(req map[string]interface{}) ([]events.FilteredEvent, error) {
	return c.conn.GetLogsEvent(req)
}

func (c *Client) GetLogTransfer(req map[string]interface{}) ([]*transfers.FilteredTransfer, error) {
	return c.conn.GetLogTransfer(req)
}

func (c *Client) GetAccount(addr *thor.Address) (*accounts.Account, error) {
	return c.conn.GetAccount(addr)
}

func (c *Client) GetContractByteCode(addr *thor.Address) ([]byte, error) {
	return c.conn.GetContractByteCode(addr)
}

func (c *Client) GetStorage(addr *thor.Address, key *thor.Bytes32) ([]byte, error) {
	return c.conn.GetStorage(addr, key)
}

func (c *Client) GetExpandedBlock(block string) (blocks *blocks.JSONExpandedBlock, err error) {
	return c.conn.GetExpandedBlock(block)
}
func (c *Client) GetBlock(block string) (blocks *blocks.JSONBlockSummary, err error) {
	return c.conn.GetBlock(block)
}

func (c *Client) GetBestBlock() (blocks *blocks.JSONExpandedBlock, err error) {
	return c.GetExpandedBlock("best")
}

func (c *Client) RawHTTPPost(url string, calldata interface{}) ([]byte, error) {
	return c.conn.RawHTTPPost(url, calldata)
}
func (c *Client) RawHTTPGet(url string) ([]byte, error) {
	return c.conn.RawHTTPGet(url)
}

func (c *Client) GetTransaction(id *thor.Bytes32) (*transactions.Transaction, error) {
	return c.conn.GetTransaction(id)
}

func (c *Client) GetPeers() ([]*node.PeerStats, error) {
	return c.conn.GetPeers()
}

func (c *Client) ChainTag() (byte, error) {
	return c.conn.ChainTag()
}

func convertToBatchCallData(tx *tx.Transaction, addr *thor.Address) *accounts.BatchCallData {
	cls := make(accounts.Clauses, len(tx.Clauses()))
	for i, c := range tx.Clauses() {
		cls[i] = convertClauseAccounts(c)
	}

	return &accounts.BatchCallData{
		Clauses: cls,
		Gas:     tx.Gas(),
		//GasPrice:   (*math.HexOrDecimal256)(big.NewInt(1)),
		ProvedWork: nil,
		Caller:     addr,
		GasPayer:   nil,
		Expiration: 0,
		BlockRef:   "",
	}
}

func convertClauseAccounts(c *tx.Clause) accounts.Clause {
	value := math.HexOrDecimal256(*c.Value())
	return accounts.Clause{
		To:    c.To(),
		Value: &value,
		Data:  hexutil.Encode(c.Data()),
	}
}
