// TODO : License related notes

package wallet

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcwallet/waddrmgr"
	"github.com/btcsuite/btcwallet/wallet/txauthor"
	"github.com/btcsuite/btcwallet/wallet/txrules"
	"github.com/btcsuite/btcwallet/walletdb"
)

//// Post-dated related codes

type SendPostDatedToAddressCmd struct {
	Address  string
	Amount   int64
	LockTime uint32
}

func NewSendPostDatedToAddressCmd(address string, amount int64, lockTime uint32) *SendPostDatedToAddressCmd {
	return &SendPostDatedToAddressCmd{
		Address:  address,
		Amount:   amount,
		LockTime: lockTime,
	}
}

func sendPostDatedTransaction(icmd interface{}, w *Wallet) (interface{}, error) {
	cmd := icmd.(*SendPostDatedToAddressCmd)
	return sendPostDated(w, cmd.Address, cmd.Amount, cmd.LockTime, waddrmgr.DefaultAccountNum)
}

func sendPostDated(w *Wallet, addrStr string, amount int64, lockTime uint32,
	account uint32) (string, error) {

	redeemTxHash, _ := w.SendPostDated(addrStr, amount, lockTime, account) //txHash
	// TODO : Check for error

	txHashStr := redeemTxHash.String()
	log.Infof("Successfully transferred transaction %v", txHashStr)
	return txHashStr, nil
}

// Entry point
func (w *Wallet) SendPostDated(addrStr string, amount int64, lockTime uint32,
	account uint32) (*chainhash.Hash, error) {
	createdTx, err := w.CreateSimplePostDatedTransfer(addrStr, amount, lockTime, account)
	if err != nil {
		return nil, err
	}

	return w.publishTransaction(createdTx.Tx)
}

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
		tx  *txauthor.AuthoredTx
		err error
	}
)

func (w *Wallet) CreateSimplePostDatedTransfer(address string, amount int64, lockTime uint32,
	account uint32) (*txauthor.AuthoredTx, error) {
	req := createPostDatedTxRequest{
		account:  account,
		address:  address,
		lockTime: lockTime,
		amount:   amount,
		resp:     make(chan createPostDatedTxResponse),
	}

	// TODO : use channels instead of direct call
	//w.createPostDatedTxRequests <- req
	//resp := <-req.resp
	resp := w.createPostDatedTx(req)
	return resp.tx, resp.err
}

// TODO : move & call this
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

const (
	// TODO : This one should be in btcd/wire
	TransferTxVersion = 99
	coincaseTxFlags   = "/POSTDATED/"

	// TODO : This one should be in btcd/blockchain
	CoincaseWitnessDataLen = 32

	// TODO : Need to be in wire package
	PostDatedTxVersion = 2
)

func createCoincaseScript() ([]byte, error) {
	return txscript.NewScriptBuilder().AddData([]byte(coincaseTxFlags)).Script()
}

// Reference : btcsuite\btcd\mining\mining.go:253
func newCoincaseTransaction(coincaseScript []byte, amount int64, lockTime uint32) (
	*btcutil.Tx, error) {
	var err error

	postDatedScript, err := createCoincaseScript()
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx(TransferTxVersion)

	tx.AddTxIn(&wire.TxIn{
		// Transfer transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOutPoint: *wire.NewOutPoint(&chainhash.Hash{},
			wire.MaxPrevOutIndex),
		SignatureScript: postDatedScript,
		Sequence:        wire.MaxTxInSequenceNum,
	})
	tx.AddTxOut(&wire.TxOut{
		Value:    amount,
		PkScript: coincaseScript,
	})

	tx.LockTime = lockTime

	// Reference : btcsuite\btcd\mining\mining.go:805
	// TODO : Check segwit related codes

	return btcutil.NewTx(tx), nil
}

func (w *Wallet) createCoincase(
	dbtx walletdb.ReadWriteTx, account uint32,
	amount int64, lockTime uint32) (coincaseTx *btcutil.Tx, err error) {

	addrmgrNs := dbtx.ReadWriteBucket(waddrmgrNamespaceKey)

	// Create coincase
	coincaseSource := func() ([]byte, error) {
		var coincaseAddr btcutil.Address
		var err error
		if account == waddrmgr.ImportedAddrAccount {
			coincaseAddr, err = w.newChangeAddress(addrmgrNs, 0)
		} else {
			coincaseAddr, err = w.newChangeAddress(addrmgrNs, account)
		}
		if err != nil {
			return nil, err
		}
		return txscript.PayToAddrScript(coincaseAddr)
	}
	coincaseScript, err := coincaseSource()
	if err != nil {
		return
	}

	coincaseTx, err = newCoincaseTransaction(coincaseScript, amount, lockTime)
	return
}

type AuthoredPostDatedTx struct {
	Tx              *wire.MsgTx
	PrevScripts     [][]byte
	PrevInputValues []btcutil.Amount
}

func NewUnsignedTransactionFromCoincase(coincaseTx *btcutil.Tx, output *wire.TxOut) (*AuthoredPostDatedTx, error) {
	// Create unsigned tx
	unsignedTransaction := &wire.MsgTx{
		Version: PostDatedTxVersion,
	}

	outpoint := wire.NewOutPoint(coincaseTx.Hash(), 0)
	txIn := wire.NewTxIn(outpoint, nil, nil)

	unsignedTransaction.AddTxIn(txIn)
	unsignedTransaction.AddTxOut(output)
	unsignedTransaction.LockTime = coincaseTx.MsgTx().LockTime

	// Get amount from coincase
	amount := btcutil.Amount(coincaseTx.MsgTx().TxOut[0].Value)
	currentInputValues := make([]btcutil.Amount, 0, 1)
	currentInputValues = append(currentInputValues, amount)

	// Get pkScript from coincase
	currentScripts := make([][]byte, 0, 1)
	currentScripts = append(currentScripts, coincaseTx.MsgTx().TxOut[0].PkScript)

	return &AuthoredPostDatedTx{
		Tx:              unsignedTransaction,
		PrevScripts:     currentScripts,
		PrevInputValues: currentInputValues,
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

	err = walletdb.Update(w.db, func(dbtx walletdb.ReadWriteTx) error {
		// Create coincase tx
		coincaseTx, err := w.createCoincase(dbtx, req.account, req.amount, req.lockTime)
		if err != nil {
			return err
		}

		NewUnsignedTransactionFromCoincase(coincaseTx, redeemOutput)

		// TODO : Continue to implementation
		return nil
	})
	if err != nil {
		return createPostDatedTxResponse{nil, err}
	}

	// TODO : Continue to implementation
	return createPostDatedTxResponse{nil, nil}
}
