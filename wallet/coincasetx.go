package wallet

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

const (
	coincaseTxFlags = "/POSTDATED/"

	// TODO : Need to be in wire package
	PostDatedTxVersion = 2
)

func createCoincaseScript() ([]byte, error) {
	return txscript.NewScriptBuilder().AddData([]byte(coincaseTxFlags)).Script()
}

// Reference : btcsuite\btcd\mining\mining.go:253
func newCoincaseTransaction(pkScript []byte, amount int64, lockTime uint32) (
	*btcutil.Tx, error) {
	var err error

	postDatedScript, err := createCoincaseScript()
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx(PostDatedTxVersion)

	tx.AddTxIn(&wire.TxIn{
		// Coincase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOutPoint: *wire.NewOutPoint(&chainhash.Hash{},
			wire.MaxPrevOutIndex),
		SignatureScript: postDatedScript,
		Sequence:        wire.MaxTxInSequenceNum,
	})
	tx.AddTxOut(&wire.TxOut{
		Value:    amount,
		PkScript: pkScript,
	})

	tx.LockTime = lockTime

	// Reference : btcsuite\btcd\mining\mining.go:805
	// TODO : Check segwit related codes

	return btcutil.NewTx(tx), nil
}

func (w *Wallet) createCoincase(
	coincaseAddr btcutil.Address,
	amount int64, lockTime uint32) (coincaseTx *btcutil.Tx, err error) {

	// Create coincase
	pkScript, err := txscript.PayToAddrScript(coincaseAddr)
	if err != nil {
		return
	}

	coincaseTx, err = newCoincaseTransaction(pkScript, amount, lockTime)
	return
}
