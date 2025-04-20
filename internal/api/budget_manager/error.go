package budget_manager

import "ProjectGolang/pkg/response"

var (
	ErrTransactionNotFound    = response.NewError(404, "transaction not found")
	ErrInvalidTransaction     = response.NewError(400, "invalid transaction data")
	ErrInvalidUserID          = response.NewError(400, "invalid user id")
	ErrInvalidCategory        = response.NewError(400, "invalid category")
	ErrInvalidTransactionType = response.NewError(400, "invalid transaction type")
	ErrInvalidAmount          = response.NewError(400, "invalid transaction amount")
	ErrCreateTransaction      = response.NewError(500, "failed to create transaction")
	ErrUpdateTransaction      = response.NewError(500, "failed to update transaction")
	ErrDeleteTransaction      = response.NewError(500, "failed to delete transaction")
	ErrTransactionNotOwned    = response.NewError(403, "transaction does not belong to user")
	ErrInvalidAudioFile       = response.NewError(400, "invalid audio file type")
	ErrFailedToUploadAudio    = response.NewError(500, "failed to upload audio file")
)
