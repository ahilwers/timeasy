package model

type Transaction interface {
	Commit() error
	Rollback() error
}

type TransactionHandler interface {
	BeginTransaction() (Transaction, error)
}
