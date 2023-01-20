package api

import (
	"net/http"

	"go.sia.tech/core/consensus"
	"go.sia.tech/core/types"
	"go.sia.tech/explorer"
	"go.sia.tech/jape"
)

type (
	// A Syncer can connect to other peers and synchronize the blockchain.
	Syncer interface {
		Addr() string
		Peers() []string
		Connect(addr string) error
		BroadcastTransaction(txn types.Transaction, dependsOn []types.Transaction)
	}

	// A TransactionPool can validate and relay unconfirmed transactions.
	TransactionPool interface {
		Transactions() []types.Transaction
		AddTransaction(txn types.Transaction) error
	}

	// A ChainManager manages blockchain state.
	ChainManager interface {
		TipState() consensus.State
	}

	// An Explorer contains a database storing information about blocks, outputs,
	// contracts.
	Explorer interface {
		SiacoinElement(id types.ElementID) (types.SiacoinElement, error)
		SiafundElement(id types.ElementID) (types.SiafundElement, error)
		FileContractElement(id types.ElementID) (types.FileContractElement, error)
		ChainStats(index types.ChainIndex) (explorer.ChainStats, error)
		ChainStatsLatest() (explorer.ChainStats, error)
		SiacoinBalance(address types.Address) (types.Currency, error)
		SiafundBalance(address types.Address) (uint64, error)
		Transaction(id types.TransactionID) (types.Transaction, error)
		UnspentSiacoinElements(address types.Address) ([]types.ElementID, error)
		UnspentSiafundElements(address types.Address) ([]types.ElementID, error)
		Transactions(address types.Address, amount, offset int) ([]types.TransactionID, error)
		State(index types.ChainIndex) (context consensus.State, err error)
	}
)

type server struct {
	s  Syncer
	e  Explorer
	cm ChainManager
	tp TransactionPool
}

func (s *server) txpoolBroadcastHandler(jc jape.Context) {
	var tbr TxpoolBroadcastRequest
	if jc.Decode(&tbr) != nil {
		return
	}

	for _, txn := range tbr.DependsOn {
		if jc.Check("couldn't broadcast transaction dependency", s.tp.AddTransaction(txn)) != nil {
			return
		}
	}
	if jc.Check("couldn't broadcast transaction dependency", s.tp.AddTransaction(tbr.Transaction)) != nil {
		return
	}
	s.s.BroadcastTransaction(tbr.Transaction, tbr.DependsOn)
}

func (s *server) txpoolTransactionsHandler(jc jape.Context) {
	jc.Encode(s.tp.Transactions())
}

func (s *server) syncerPeersHandler(jc jape.Context) {
	ps := s.s.Peers()
	sps := make([]SyncerPeerResponse, len(ps))
	for i, peer := range ps {
		sps[i] = SyncerPeerResponse{
			NetAddress: peer,
		}
	}
	jc.Encode(sps)
}

func (s *server) syncerConnectHandler(jc jape.Context) {
	var addr string
	if jc.Decode(&addr) != nil {
		return
	}
	if jc.Check("failed to connect to peer", s.s.Connect(addr)) != nil {
		return
	}
}

func (s *server) elementSiacoinHandler(jc jape.Context) {
	var id types.ElementID
	if jc.DecodeParam("id", &id) != nil {
		return
	}

	elem, err := s.e.SiacoinElement(id)
	if jc.Check("failed to load siacoin element", err) != nil {
		return
	}
	jc.Encode(elem)
}

func (s *server) elementSiafundHandler(jc jape.Context) {
	var id types.ElementID
	if jc.DecodeParam("id", &id) != nil {
		return
	}

	elem, err := s.e.SiafundElement(id)
	if jc.Check("failed to load siafund element", err) != nil {
		return
	}
	jc.Encode(elem)
}

func (s *server) elementContractHandler(jc jape.Context) {
	var id types.ElementID
	if jc.DecodeParam("id", &id) != nil {
		return
	}

	elem, err := s.e.FileContractElement(id)
	if jc.Check("failed to load siafund element", err) != nil {
		return
	}
	jc.Encode(elem)
}

func (s *server) chainStatsHandler(jc jape.Context) {
	if jc.PathParam("index") == "tip" {
		facts, err := s.e.ChainStatsLatest()
		if jc.Check("failed to load latest chain stats", err) != nil {
			return
		}
		jc.Encode(facts)
		return
	}

	index, err := types.ParseChainIndex(jc.PathParam("index"))
	if jc.Check("failed to parse chain index", err) != nil {
		return
	}

	facts, err := s.e.ChainStats(index)
	if jc.Check("failed to load chain stats", err) != nil {
		return
	}
	jc.Encode(facts)
}

func (s *server) chainStateHandler(jc jape.Context) {
	index, err := types.ParseChainIndex(jc.PathParam("index"))
	if jc.Check("failed to parse chain index", err) != nil {
		return
	}

	vc, err := s.e.State(index)
	if jc.Check("failed to load chain state", err) != nil {
		return
	}
	jc.Encode(vc)
}

func (s *server) elementSearchHandler(jc jape.Context) {
	var id types.ElementID
	if jc.DecodeParam("id", &id) != nil {
		return
	}

	var response ExplorerSearchResponse
	if elem, err := s.e.SiacoinElement(id); err == nil {
		response.Type = "siacoin"
		response.SiacoinElement = elem
	} else if elem, err := s.e.SiafundElement(id); err == nil {
		response.Type = "siafund"
		response.SiafundElement = elem
	} else if elem, err := s.e.FileContractElement(id); err == nil {
		response.Type = "contract"
		response.FileContractElement = elem
	}
	jc.Encode(response)
}

func (s *server) addressBalanceHandler(jc jape.Context) {
	var address types.Address
	if jc.DecodeParam("address", &address) != nil {
		return
	}

	scBalance, err := s.e.SiacoinBalance(address)
	if jc.Check("failed to get siacoin balance", err) != nil {
		return
	}

	sfBalance, err := s.e.SiafundBalance(address)
	if jc.Check("failed to get siafund balance", err) != nil {
		return
	}

	jc.Encode(ExplorerWalletBalanceResponse{scBalance, sfBalance})
}

func (s *server) addressSiacoinsHandler(jc jape.Context) {
	var address types.Address
	if jc.DecodeParam("address", &address) != nil {
		return
	}

	outputs, err := s.e.UnspentSiacoinElements(address)
	if jc.Check("failed to get unspent siacoin elements", err) != nil {
		return
	}
	jc.Encode(outputs)
}

func (s *server) addressSiafundsHandler(jc jape.Context) {
	var address types.Address
	if jc.DecodeParam("address", &address) != nil {
		return
	}

	outputs, err := s.e.UnspentSiafundElements(address)
	if jc.Check("failed to get unspent siafund elements", err) != nil {
		return
	}
	jc.Encode(outputs)
}

func (s *server) addressTransactionsHandler(jc jape.Context) {
	var address types.Address
	if jc.DecodeParam("address", &address) != nil {
		return
	}

	var amount int
	if jc.DecodeForm("amount", &amount) != nil {
		return
	}

	var offset int
	if jc.DecodeForm("offset", &amount) != nil {
		return
	}

	ids, err := s.e.Transactions(address, amount, offset)
	if jc.Check("failed to get address' transactions", err) != nil {
		return
	}

	jc.Encode(ids)
}

func (s *server) transactionHandler(jc jape.Context) {
	var id types.TransactionID
	if jc.DecodeParam("id", &id) != nil {
		return
	}

	txn, err := s.e.Transaction(id)
	if jc.Check("failed to load transaction", err) != nil {
		return
	}
	jc.Encode(txn)
}

func (s *server) batchAddressesBalanceHandler(jc jape.Context) {
	var addresses []types.Address
	if jc.Decode(&addresses) != nil {
		return
	}

	var balances []ExplorerWalletBalanceResponse
	for _, address := range addresses {
		scBalance, err := s.e.SiacoinBalance(address)
		if jc.Check("failed to get siacoin balance", err) != nil {
			return
		}

		sfBalance, err := s.e.SiafundBalance(address)
		if jc.Check("failed to get siafund balance", err) != nil {
			return
		}

		balances = append(balances, ExplorerWalletBalanceResponse{scBalance, sfBalance})
	}
	jc.Encode(balances)
}

func (s *server) batchAddressesSiacoinsHandler(jc jape.Context) {
	var addresses []types.Address
	if jc.Decode(&addresses) != nil {
		return
	}

	var elems [][]types.SiacoinElement
	for _, address := range addresses {
		ids, err := s.e.UnspentSiacoinElements(address)
		if jc.Check("failed to load unspent siacoin elements", err) != nil {
			return
		}

		var elemsList []types.SiacoinElement
		for _, id := range ids {
			elem, err := s.e.SiacoinElement(id)
			if jc.Check("failed to load siacoin elements", err) != nil {
				return
			}
			elemsList = append(elemsList, elem)
		}
		elems = append(elems, elemsList)
	}
	jc.Encode(elems)
}

func (s *server) batchAddressesSiafundsHandler(jc jape.Context) {
	var addresses []types.Address
	if jc.Decode(&addresses) != nil {
		return
	}

	var elems [][]types.SiafundElement
	for _, address := range addresses {
		ids, err := s.e.UnspentSiafundElements(address)
		if jc.Check("failed to load unspent siafund elements", err) != nil {
			return
		}

		var elemsList []types.SiafundElement
		for _, id := range ids {
			elem, err := s.e.SiafundElement(id)
			if jc.Check("failed to load siafund elements", err) != nil {
				return
			}
			elemsList = append(elemsList, elem)
		}
		elems = append(elems, elemsList)
	}
	jc.Encode(elems)
}

func (s *server) batchAddressesTransactionsHandler(jc jape.Context) {
	var etrs []ExplorerTransactionsRequest
	if jc.Decode(&etrs) != nil {
		return
	}

	var txns [][]types.Transaction
	for _, etr := range etrs {
		ids, err := s.e.Transactions(etr.Address, etr.Amount, etr.Offset)
		if jc.Check("failed to load transactions", err) != nil {
			return
		}

		var txnsList []types.Transaction
		for _, id := range ids {
			txn, err := s.e.Transaction(id)
			if jc.Check("failed to load transaction", err) != nil {
				return
			}
			txnsList = append(txnsList, txn)
		}
		txns = append(txns, txnsList)
	}
	jc.Encode(txns)
}

// NewServer returns an HTTP handler that serves the explorerd API.
func NewServer(cm ChainManager, s Syncer, tp TransactionPool, e Explorer) http.Handler {
	srv := server{
		cm: cm,
		s:  s,
		tp: tp,
		e:  e,
	}
	return jape.Mux(map[string]jape.Handler{
		"GET /txpool/transactions": srv.txpoolTransactionsHandler,
		"POST /txpool/broadcast":   srv.txpoolBroadcastHandler,

		"GET /syncer/peers":    srv.syncerPeersHandler,
		"POST /syncer/connect": srv.syncerConnectHandler,

		"GET /explorer/element/search/:id":   srv.elementSearchHandler,
		"GET /explorer/element/siacoin/:id":  srv.elementSiacoinHandler,
		"GET /explorer/element/siafund/:id":  srv.elementSiafundHandler,
		"GET /explorer/element/contract/:id": srv.elementContractHandler,

		"GET /explorer/chain/:index":       srv.chainStatsHandler,
		"GET /explorer/chain/:index/state": srv.chainStateHandler,

		"GET /explorer/transaction/:id": srv.transactionHandler,

		"GET /explorer/address/:address/balance":      srv.addressBalanceHandler,
		"GET /explorer/address/:address/siacoins":     srv.addressSiacoinsHandler,
		"GET /explorer/address/:address/siafunds":     srv.addressSiacoinsHandler,
		"GET /explorer/address/:address/transactions": srv.addressTransactionsHandler,

		"POST /explorer/batch/addresses/balance":      srv.batchAddressesBalanceHandler,
		"POST /explorer/batch/addresses/siacoins":     srv.batchAddressesSiacoinsHandler,
		"POST /explorer/batch/addresses/siafunds":     srv.batchAddressesSiafundsHandler,
		"POST /explorer/batch/addresses/transactions": srv.batchAddressesTransactionsHandler,
	})
}

// AuthMiddleware enforces HTTP Basic Authentication on the provided handler.
func AuthMiddleware(handler http.Handler, requiredPass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, password, ok := req.BasicAuth(); !ok || password != requiredPass {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, req)
	})
}
