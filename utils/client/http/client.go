package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vechain/networkhub/utils/common"
	"github.com/vechain/thor/api/accounts"
	"github.com/vechain/thor/api/blocks"
	"github.com/vechain/thor/api/events"
	"github.com/vechain/thor/api/node"
	"github.com/vechain/thor/api/transactions"
	"github.com/vechain/thor/api/transfers"
	"github.com/vechain/thor/thor"
)

type Client struct {
	url string
	c   *http.Client
}

func NewClient(url string) *Client {
	return &Client{
		url: url,
		//c:   newLogWriter(),
		c: &http.Client{},
	}
}

func (c *Client) GetTransactionReceipt(txID *thor.Bytes32) (*transactions.Receipt, error) {
	body, err := c.httpGET(c.url + "/transactions/" + txID.String() + "/receipt")
	if err != nil {
		return nil, fmt.Errorf("unable to fetch receipt - %w", err)
	}

	if string(body) == "null\n" {
		return nil, common.NotFoundErr
	}

	var receipt transactions.Receipt
	if err = json.Unmarshal(body, &receipt); err != nil {
		return nil, fmt.Errorf("unable to unmarshall receipt - %w", err)
	}

	return &receipt, nil
}

func (c *Client) InspectClauses(calldata *accounts.BatchCallData) ([]*accounts.CallResult, error) {
	body, err := c.httpPOSTObj(c.url+"/accounts/*", calldata)
	if err != nil {
		return nil, fmt.Errorf("unable to request inspect clauses - %w", err)
	}

	var inspectionRes []*accounts.CallResult
	if err = json.Unmarshal(body, &inspectionRes); err != nil {
		return nil, fmt.Errorf("unable to unmarshall inspection - %w", err)
	}

	return inspectionRes, nil
}

func (c *Client) SendTransaction(obj *transactions.RawTx) (*common.TxSendResult, error) {
	body, err := c.httpPOSTObj(c.url+"/transactions", obj)
	if err != nil {
		return nil, fmt.Errorf("unable to send raw transaction - %w", err)
	}

	var txID common.TxSendResult
	if err = json.Unmarshal(body, &txID); err != nil {
		return nil, fmt.Errorf("unable to unmarshall inspection - %w", err)
	}

	return &txID, nil
}

func (c *Client) GetLogs(eventEndpoint string, req map[string]interface{}) ([]events.FilteredEvent, error) {
	body, err := c.httpPOSTObj(eventEndpoint, req)
	if err != nil {
		return nil, fmt.Errorf("unable to send raw transaction - %w", err)
	}

	var filteredEvents []events.FilteredEvent
	if err = json.Unmarshal(body, &filteredEvents); err != nil {
		return nil, fmt.Errorf("unable to unmarshall events - %w", err)
	}

	return filteredEvents, nil
}

func (c *Client) GetLogTransfer(req map[string]interface{}) ([]*transfers.FilteredTransfer, error) {
	body, err := c.httpPOSTObj(c.url+"/logs/transfer", req)
	if err != nil {
		return nil, fmt.Errorf("unable to send retrieve transfer logs - %w", err)
	}

	var filteredEvents []*transfers.FilteredTransfer
	if err = json.Unmarshal(body, &filteredEvents); err != nil {
		return nil, fmt.Errorf("unable to unmarshall events - %w", err)
	}

	return filteredEvents, nil
}

func (c *Client) GetLogsEvent(req map[string]interface{}) ([]events.FilteredEvent, error) {
	return c.GetLogs(c.url+"/logs/event", req)
}

func (c *Client) GetAccount(addr *thor.Address) (*accounts.Account, error) {
	body, err := c.httpGET(c.url + "/accounts/" + addr.String())
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve account - %w", err)
	}

	var account accounts.Account
	if err = json.Unmarshal(body, &account); err != nil {
		return nil, fmt.Errorf("unable to unmarshall events - %w", err)
	}

	return &account, nil
}

func (c *Client) GetContractByteCode(addr *thor.Address) ([]byte, error) {
	return c.httpGET(c.url + "/accounts/" + addr.String() + "/code")
}

func (c *Client) GetStorage(addr *thor.Address, key *thor.Bytes32) ([]byte, error) {
	return c.httpGET(c.url + "/accounts/" + addr.String() + "/key/" + key.String())
}

func (c *Client) GetExpandedBlock(blockID string) (*blocks.JSONExpandedBlock, error) {
	body, err := c.httpGET(c.url + "/blocks/" + blockID + "?expanded=true")
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve block - %w", err)
	}

	var block blocks.JSONExpandedBlock
	if err = json.Unmarshal(body, &block); err != nil {
		return nil, fmt.Errorf("unable to unmarshall events - %w", err)
	}

	return &block, nil
}
func (c *Client) GetBlock(blockID string) (*blocks.JSONBlockSummary, error) {
	body, err := c.httpGET(c.url + "/blocks/" + blockID)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve block - %w", err)
	}

	var block blocks.JSONBlockSummary
	if err = json.Unmarshal(body, &block); err != nil {
		return nil, fmt.Errorf("unable to unmarshall events - %w", err)
	}

	return &block, nil
}

func (c *Client) GetTransaction(txID *thor.Bytes32) (*transactions.Transaction, error) {
	body, err := c.httpGET(c.url + "/transactions/" + txID.String())
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve transaction - %w", err)
	}

	var tx transactions.Transaction
	if err = json.Unmarshal(body, &tx); err != nil {
		return nil, fmt.Errorf("unable to unmarshall events - %w", err)
	}

	return &tx, nil
}

func (c *Client) RawHTTPPost(url string, calldata interface{}) ([]byte, error) {
	return c.httpPOSTObj(c.url+url, calldata)
}

func (c *Client) RawHTTPGet(url string) ([]byte, error) {
	return c.httpGET(c.url + url)
}

func (c *Client) GetPeers() ([]*node.PeerStats, error) {
	body, err := c.httpGET(c.url + "/node/network/peers")
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve peers - %w", err)
	}

	var peers []*node.PeerStats
	if err = json.Unmarshal(body, &peers); err != nil {
		return nil, fmt.Errorf("unable to unmarshall events - %w", err)
	}

	return peers, nil
}
