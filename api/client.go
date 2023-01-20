package api

import (
	"fmt"

	"go.sia.tech/core/consensus"
	"go.sia.tech/core/types"
	"go.sia.tech/explorer"
	"go.sia.tech/jape"
)

// A Client provides methods for interacting with a walletd API server.
type Client struct {
	c jape.Client
}

// TxpoolBroadcast broadcasts a transaction to the network.
func (c *Client) TxpoolBroadcast(txn types.Transaction, dependsOn []types.Transaction) (err error) {
	err = c.c.POST("/txpool/broadcast", TxpoolBroadcastRequest{dependsOn, txn}, nil)
	return
}

// TxpoolTransactions returns all transactions in the transaction pool.
func (c *Client) TxpoolTransactions() (resp []types.Transaction, err error) {
	err = c.c.GET("/txpool/transactions", &resp)
	return
}

// SyncerPeers returns the current peers of the syncer.
func (c *Client) SyncerPeers() (resp []SyncerPeerResponse, err error) {
	err = c.c.GET("/syncer/peers", &resp)
	return
}

// SyncerConnect adds the address as a peer of the syncer.
func (c *Client) SyncerConnect(addr string) (err error) {
	err = c.c.POST("/syncer/connect", addr, nil)
	return
}

// ChainStats returns stats about the chain at the given index.
func (c *Client) ChainStats(index types.ChainIndex) (resp explorer.ChainStats, err error) {
	err = c.c.GET(fmt.Sprintf("/explorer/chain/%s", index.String()), &resp)
	return
}

// ChainState returns the validation context at a given chain index.
func (c *Client) ChainState(index types.ChainIndex) (resp consensus.State, err error) {
	err = c.c.GET(fmt.Sprintf("/explorer/chain/%s/state", index.String()), &resp)
	return
}

// SiacoinElement returns the Siacoin element with the given ID.
func (c *Client) SiacoinElement(id types.ElementID) (resp types.SiacoinElement, err error) {
	err = c.c.GET(fmt.Sprintf("/explorer/element/siacoin/%s", id), &resp)
	return
}

// SiafundElement returns the Siafund element with the given ID.
func (c *Client) SiafundElement(id types.ElementID) (resp types.SiafundElement, err error) {
	err = c.c.GET(fmt.Sprintf("/explorer/element/siafund/%s", id), &resp)
	return
}

// FileContractElement returns the file contract element with the given ID.
func (c *Client) FileContractElement(id types.ElementID) (resp types.FileContractElement, err error) {
	err = c.c.GET(fmt.Sprintf("/explorer/element/contract/%s", id), &resp)
	return
}

// ElementSearch returns information about a given element.
func (c *Client) ElementSearch(id types.ElementID) (resp ExplorerSearchResponse, err error) {
	err = c.c.GET(fmt.Sprintf("/explorer/element/search/%s", id), &resp)
	return
}

// AddressBalance returns the siacoin and siafund balance of an address.
func (c *Client) AddressBalance(address types.Address) (resp ExplorerWalletBalanceResponse, err error) {
	err = c.c.GET(fmt.Sprintf("/explorer/address/%s/balance", address), &resp)
	return
}

// SiacoinOutputs returns the unspent siacoin elements of an address.
func (c *Client) SiacoinOutputs(address types.Address) (resp []types.ElementID, err error) {
	err = c.c.GET(fmt.Sprintf("/explorer/address/%s/siacoins", address), &resp)
	return
}

// SiafundOutputs returns the unspent siafunds elements of an address.
func (c *Client) SiafundOutputs(address types.Address) (resp []types.ElementID, err error) {
	err = c.c.GET(fmt.Sprintf("/explorer/address/%s/siafunds", address), &resp)
	return
}

// Transactions returns the latest transaction IDs the address was involved in.
func (c *Client) Transactions(address types.Address, amount, offset int) (resp []types.TransactionID, err error) {
	err = c.c.GET(fmt.Sprintf("/explorer/address/%s/transactions?amount=%d&offset=%d", address, amount, offset), &resp)
	return
}

// Transaction returns a transaction with the given ID.
func (c *Client) Transaction(id types.TransactionID) (resp types.Transaction, err error) {
	err = c.c.GET(fmt.Sprintf("/explorer/transaction/%s", id), &resp)
	return
}

// BatchBalance returns the siacoin and siafund balance of a list of addresses.
func (c *Client) BatchBalance(addresses []types.Address) (resp []ExplorerWalletBalanceResponse, err error) {
	err = c.c.POST("/explorer/batch/addresses/balance", addresses, &resp)
	return
}

// BatchSiacoins returns the unspent siacoin elements of the addresses.
func (c *Client) BatchSiacoins(addresses []types.Address) (resp [][]types.SiacoinElement, err error) {
	err = c.c.POST("/explorer/batch/addresses/siacoins", addresses, &resp)
	return
}

// BatchSiafunds returns the unspent siafund elements of the addresses.
func (c *Client) BatchSiafunds(addresses []types.Address) (resp [][]types.SiafundElement, err error) {
	err = c.c.POST("/explorer/batch/addresses/siafunds", addresses, &resp)
	return
}

// BatchTransactions returns the last n transactions of the addresses.
func (c *Client) BatchTransactions(addresses []ExplorerTransactionsRequest) (resp [][]types.Transaction, err error) {
	err = c.c.POST("/explorer/batch/addresses/transactions", addresses, &resp)
	return
}

// NewClient returns a client that communicates with a explorer server listening
// on the specified address.
func NewClient(addr, password string) *Client {
	return &Client{jape.Client{
		BaseURL:  addr,
		Password: password,
	}}
}
