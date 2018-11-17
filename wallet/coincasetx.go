package wallet

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// Reference : btcsuite\btcd\mining\mining.go:253
func newCoincaseTransaction(pkScript []byte, amount int64) (
	*btcutil.Tx, error) {
	var err error

	postDatedScript, err := txscript.CreateCoincaseScript()
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx(wire.PostDatedTxVersion)

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

	// Reference : btcsuite\btcd\mining\mining.go:805
	// TODO : Check segwit related codes

	return btcutil.NewTx(tx), nil
}

func (w *Wallet) createCoincase(coincaseAddr btcutil.Address, amount int64) (coincaseTx *btcutil.Tx, err error) {

	// Create coincase
	pkScript, err := txscript.PayToAddrScript(coincaseAddr)
	if err != nil {
		return
	}

	coincaseTx, err = newCoincaseTransaction(pkScript, amount)
	return
}
