package wallet

type TransactionTransferError interface {
	error
	TransactionTransferError()
}

// Transaction not found error
type txNotFoundError struct{}

func (txNotFoundError) Error() string {
	return "transaction not found to transfer"
}
func (txNotFoundError) TransactionTransferError() {}

// Transaction in not mature error
type txNotMatureError struct{}

func (txNotMatureError) Error() string {
	return "transaction is not mature"
}
func (txNotMatureError) TransactionTransferError() {}

// Coinbase transaction in not mature error
type txCoinbaseNotMatureError struct{}

func (txCoinbaseNotMatureError) Error() string {
	return "Coin transaction is not mature"
}
func (txCoinbaseNotMatureError) TransactionTransferError() {}

// transaction is locked error
type txIsLockedError struct{}

func (txIsLockedError) Error() string {
	return "Transaction is locked"
}
func (txIsLockedError) TransactionTransferError() {}

// transaction is not owned by account error
type txIsNotOwnedError struct{}

func (txIsNotOwnedError) Error() string {
	return "Transaction is not owned by this account"
}
func (txIsNotOwnedError) TransactionTransferError() {}
