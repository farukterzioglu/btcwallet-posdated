// TODO : License related notes

package wallet

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcwallet/waddrmgr"
	"github.com/btcsuite/btcwallet/wallet/txauthor"
	"github.com/btcsuite/btcwallet/wallet/txrules"
	"github.com/btcsuite/btcwallet/walletdb"
)

// TODO : move to wallet/wallet.go
func (w *Wallet) SendPostDated(addrStr string, amount int64, lockTime uint32,
	account uint32) ([]*chainhash.Hash, error) {
	createdTx, err := w.createSimplePostDatedTx(addrStr, amount, lockTime, account)
	if err != nil {
		return nil, err
	}

	coincaseHash, err := w.publishTransaction(createdTx.CoincaseTx)
	if err != nil {
		return nil, err
	}

	postDatedhash, err := w.publishTransaction(createdTx.PostDatedTx)
	if err != nil {
		return nil, err
	}

	return []*chainhash.Hash{coincaseHash, postDatedhash}, nil
}

type AuthoredPostDatedTx struct {
	CoincaseTx  *wire.MsgTx
	PostDatedTx *wire.MsgTx
	// CoincaseTx *txauthor.AuthoredTx
	// PostDatedTx *txauthor.AuthoredTx
}

// wallet/wallet.go
type (
	createPostDatedTxRequest struct {
		account     uint32
		address     string
		amount      int64
		lockTime    uint32
		feeSatPerKB btcutil.Amount
		resp        chan createPostDatedTxResponse
	}
	createPostDatedTxResponse struct {
		tx  *AuthoredPostDatedTx
		err error
	}
)

// wallet/wallet.go
func (w *Wallet) createSimplePostDatedTx(address string, amount int64, lockTime uint32,
	account uint32) (*AuthoredPostDatedTx, error) {
	req := createPostDatedTxRequest{
		account:  account,
		address:  address,
		lockTime: lockTime,
		amount:   amount,
		resp:     make(chan createPostDatedTxResponse),
	}

	w.createPostDatedTxRequests <- req
	resp := <-req.resp
	return resp.tx, resp.err
}

// TODO : move to wallet/wallet.go
func (w *Wallet) postDatedTxCreator() {
	quit := w.quitChan()
out:
	for {
		select {
		case txr := <-w.createPostDatedTxRequests:
			heldUnlock, err := w.holdUnlock()
			if err != nil {
				txr.resp <- createPostDatedTxResponse{nil, err}
				continue
			}
			res := w.createPostDatedTx(txr)
			heldUnlock.release()
			txr.resp <- res
		case <-quit:
			break out
		}
	}
	w.wg.Done()
}

func NewUnsignedTransactionFromCoincase(coincaseTx *btcutil.Tx, output *wire.TxOut) (*txauthor.AuthoredTx, error) {
	// Create unsigned tx
	unsignedTransaction := &wire.MsgTx{
		Version:  wire.PostDatedTxVersion,
		LockTime: 0,
	}

	outpoint := wire.NewOutPoint(coincaseTx.Hash(), 0)
	txIn := wire.NewTxIn(outpoint, nil, nil)

	unsignedTransaction.AddTxIn(txIn)
	unsignedTransaction.AddTxOut(output)

	// Get amount from coincase
	amount := btcutil.Amount(coincaseTx.MsgTx().TxOut[0].Value)
	currentInputValues := []btcutil.Amount{amount}

	// Get pkScript from coincase
	currentScripts := [][]byte{coincaseTx.MsgTx().TxOut[0].PkScript}

	return &txauthor.AuthoredTx{
		Tx:              unsignedTransaction,
		PrevScripts:     currentScripts,
		PrevInputValues: currentInputValues,
		TotalInput:      amount,
		ChangeIndex:     -1,
	}, nil
}

//
func (w *Wallet) createPostDatedTx(req createPostDatedTxRequest) createPostDatedTxResponse {
	amount := btcutil.Amount(req.amount)
	redeemOutput, err := makeOutput(req.address, amount, w.ChainParams())
	if err != nil {
		return createPostDatedTxResponse{nil, err}
	}

	if err := txrules.CheckOutput(redeemOutput, req.feeSatPerKB); err != nil {
		return createPostDatedTxResponse{nil, err}
	}

	// Reference server.go:236
	var coincaseAddr btcutil.Address
	coincaseAddr, err = w.NewChangeAddress(req.account, waddrmgr.KeyScopeBIP0044)

	coincaseTx, err := w.createCoincase(coincaseAddr, req.amount)
	if err != nil {
		return createPostDatedTxResponse{nil, err}
	}

	var tx *txauthor.AuthoredTx
	err = walletdb.Update(w.db, func(dbtx walletdb.ReadWriteTx) error {
		addrmgrNs := dbtx.ReadWriteBucket(waddrmgrNamespaceKey)

		tx, err = NewUnsignedTransactionFromCoincase(coincaseTx, redeemOutput)
		if err != nil {
			return err
		}

		// TODO : How to set post dating feature. Transaction level lock time or script level lock time
		tx.Tx.LockTime = req.lockTime
		return tx.AddAllInputScripts(secretSource{w.Manager, addrmgrNs})
	})
	if err != nil {
		return createPostDatedTxResponse{nil, err}
	}

	err = validateMsgTx(tx.Tx, tx.PrevScripts, tx.PrevInputValues)
	if err != nil {
		return createPostDatedTxResponse{nil, err}
	}

	result := &AuthoredPostDatedTx{
		CoincaseTx:  coincaseTx.MsgTx(),
		PostDatedTx: tx.Tx,
	}

	return createPostDatedTxResponse{result, nil}
}
