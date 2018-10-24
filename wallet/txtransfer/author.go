package txtransfer

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

const (
	// TODO : This one should be in btcd/wire
	TransferTxVersion = 99
	coincaseTxFlags   = "/POSTDATED/"

	// TODO : This one should be in btcd/blockchain
	CoincaseWitnessDataLen = 32
)

type AuthoredTx struct {
	Tx *wire.MsgTx
}

func createCoincaseScript() ([]byte, error) {
	return txscript.NewScriptBuilder().AddData([]byte(coincaseTxFlags)).Script()
}

// Reference : btcsuite\btcd\mining\mining.go:253
func newCoincaseTransaction(addr btcutil.Address, amount int64, lockTime uint32) (*btcutil.Tx, error) {
	var pkScript []byte
	var err error

	pkScript, err = txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, err
	}

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
		PkScript: pkScript,
	})

	tx.LockTime = lockTime

	// Reference : btcsuite\btcd\mining\mining.go:805
	// TODO : Check segwit related codes

	return btcutil.NewTx(tx), nil
}
