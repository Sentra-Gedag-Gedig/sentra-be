package sentrapay

import (
	"ProjectGolang/pkg/response"
)

//var (
//	ErrInsufficientBalance = errors.New("insufficient balance")
//
//	ErrTransactionNotFound = errors.New("transaction not found")
//
//	ErrCreateTransaction = errors.New("failed to create transaction")
//
//	ErrUpdateTransaction = errors.New("failed to update transaction")
//
//	ErrDeleteTransaction = errors.New("failed to delete transaction")
//
//	ErrCreateVirtualAccount = errors.New("failed to create virtual account")
//
//	ErrInvalidBank = errors.New("invalid bank selection")
//
//	ErrInvalidAmount = errors.New("invalid amount")
//
//	ErrUserNotFound = errors.New("user not found")
//
//	ErrWalletNotFound = errors.New("wallet not found")
//
//	ErrInvalidCallback = errors.New("invalid callback data")
//
//	ErrInvalidTransactionState = errors.New("invalid transaction state")
//)

var (
	ErrInsufficientBalance     = response.NewError(400, "insufficient balance")
	ErrTransactionNotFound     = response.NewError(404, "transaction not found")
	ErrCreateTransaction       = response.NewError(500, "failed to create transaction")
	ErrUpdateTransaction       = response.NewError(500, "failed to update transaction")
	ErrDeleteTransaction       = response.NewError(500, "failed to delete transaction")
	ErrCreateVirtualAccount    = response.NewError(500, "failed to create virtual account")
	ErrInvalidBank             = response.NewError(400, "invalid bank selection")
	ErrInvalidAmount           = response.NewError(400, "invalid amount")
	ErrWalletNotFound          = response.NewError(404, "wallet not found")
	ErrInvalidCallback         = response.NewError(400, "invalid callback data")
	ErrInvalidTransactionState = response.NewError(400, "invalid transaction state")
)
