package sentrapayRepository

import (
	sentrapay "ProjectGolang/internal/api/sentra_pay"
	contextPkg "ProjectGolang/pkg/context"
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

type WalletDB struct {
	ID        sql.NullString  `db:"id"`
	UserID    sql.NullString  `db:"user_id"`
	Balance   sql.NullFloat64 `db:"balance"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedAt time.Time       `db:"updated_at"`
}

type WalletTransactionDB struct {
	ID            sql.NullString  `db:"id"`
	UserID        sql.NullString  `db:"user_id"`
	Amount        sql.NullFloat64 `db:"amount"`
	Type          sql.NullString  `db:"type"`
	ReferenceNo   sql.NullString  `db:"reference_no"`
	PaymentMethod sql.NullString  `db:"payment_method"`
	Status        sql.NullString  `db:"status"`
	BankAccount   sql.NullString  `db:"bank_account"`
	BankName      sql.NullString  `db:"bank_name"`
	Description   sql.NullString  `db:"description"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
}

func (r *walletRepository) CreateWallet(ctx context.Context, userID string) error {
	requestID := contextPkg.GetRequestID(ctx)

	argsKV := map[string]interface{}{
		"id":         userID,
		"user_id":    userID,
		"balance":    0,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}

	query, args, err := sqlx.Named(queryCreateWallet, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to build SQL query for CreateWallet")
		return err
	}

	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Database error when creating wallet")
		return err
	}

	return nil
}

func (r *walletRepository) GetWallet(ctx context.Context, userID string) (sentrapay.WalletBalance, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var wallet WalletDB

	argsKV := map[string]interface{}{
		"user_id": userID,
	}

	query, args, err := sqlx.Named(queryGetWallet, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetWallet named query preparation err")
		return sentrapay.WalletBalance{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(ctx, query, args...).StructScan(&wallet); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("GetWallet no rows found")
			return sentrapay.WalletBalance{}, sentrapay.ErrWalletNotFound
		}

		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetWallet execution err")

		return sentrapay.WalletBalance{}, err
	}

	return sentrapay.WalletBalance{
		UserID:      wallet.UserID.String,
		Balance:     wallet.Balance.Float64,
		LastUpdated: wallet.UpdatedAt,
	}, nil
}

func (r *walletRepository) UpdateWalletBalance(ctx context.Context, userID string, amount float64) error {
	requestID := contextPkg.GetRequestID(ctx)

	argsKV := map[string]interface{}{
		"user_id":    userID,
		"balance":    amount,
		"updated_at": time.Now(),
	}

	query, args, err := sqlx.Named(queryUpdateWalletBalance, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateWalletBalance named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	result, err := r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateWalletBalance execution err")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateWalletBalance rows affected err")
		return err
	}

	if rowsAffected == 0 {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
		}).Warn("UpdateWalletBalance no rows affected")
		return sentrapay.ErrWalletNotFound
	}

	return nil
}

func (r *walletRepository) CreateTransaction(ctx context.Context, transaction sentrapay.WalletTransaction) error {
	requestID := contextPkg.GetRequestID(ctx)

	argsKV := map[string]interface{}{
		"id":             transaction.ID,
		"user_id":        transaction.UserID,
		"amount":         transaction.Amount,
		"type":           transaction.Type,
		"reference_no":   transaction.ReferenceNo,
		"payment_method": transaction.PaymentMethod,
		"status":         transaction.Status,
		"bank_account":   transaction.BankAccount,
		"bank_name":      transaction.BankName,
		"description":    transaction.Description,
		"created_at":     transaction.CreatedAt,
		"updated_at":     transaction.UpdatedAt,
	}

	query, args, err := sqlx.Named(queryCreateTransaction, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to build SQL query for CreateTransaction")
		return err
	}

	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Database error when creating transaction")
		return err
	}

	return nil
}

func (r *walletRepository) GetTransactionByID(ctx context.Context, id string) (sentrapay.WalletTransaction, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var transaction WalletTransactionDB

	argsKV := map[string]interface{}{
		"id": id,
	}

	query, args, err := sqlx.Named(queryGetTransactionByID, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionByID named query preparation err")
		return sentrapay.WalletTransaction{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(ctx, query, args...).StructScan(&transaction); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("GetTransactionByID no rows found")
			return sentrapay.WalletTransaction{}, sentrapay.ErrTransactionNotFound
		}

		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionByID execution err")

		return sentrapay.WalletTransaction{}, err
	}

	return r.makeWalletTransaction(transaction), nil
}

func (r *walletRepository) GetTransactionByReferenceNo(ctx context.Context, referenceNo string) (sentrapay.WalletTransaction, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var transaction WalletTransactionDB

	argsKV := map[string]interface{}{
		"reference_no": referenceNo,
	}

	query, args, err := sqlx.Named(queryGetTransactionByReferenceNo, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionByReferenceNo named query preparation err")
		return sentrapay.WalletTransaction{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(ctx, query, args...).StructScan(&transaction); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("GetTransactionByReferenceNo no rows found")
			return sentrapay.WalletTransaction{}, sentrapay.ErrTransactionNotFound
		}

		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionByReferenceNo execution err")

		return sentrapay.WalletTransaction{}, err
	}

	return r.makeWalletTransaction(transaction), nil
}

func (r *walletRepository) UpdateTransactionStatus(ctx context.Context, referenceNo string, status string) error {
	requestID := contextPkg.GetRequestID(ctx)

	argsKV := map[string]interface{}{
		"reference_no": referenceNo,
		"status":       status,
		"updated_at":   time.Now(),
	}

	query, args, err := sqlx.Named(queryUpdateTransactionStatus, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateTransactionStatus named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	result, err := r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateTransactionStatus execution err")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateTransactionStatus rows affected err")
		return err
	}

	if rowsAffected == 0 {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
		}).Warn("UpdateTransactionStatus no rows affected")
		return sentrapay.ErrTransactionNotFound
	}

	return nil
}

func (r *walletRepository) GetTransactionsByUserID(ctx context.Context, userID string, limit, offset int) ([]sentrapay.WalletTransaction, int, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var transactions []WalletTransactionDB
	var total int

	countArgsKV := map[string]interface{}{
		"user_id": userID,
	}

	countQuery, countArgs, err := sqlx.Named(queryCountTransactionsByUserID, countArgsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("CountTransactionsByUserID named query preparation err")
		return nil, 0, err
	}

	countQuery = r.q.Rebind(countQuery)

	if err := r.q.QueryRowxContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("CountTransactionsByUserID execution err")
		return nil, 0, err
	}

	argsKV := map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	}

	query, args, err := sqlx.Named(queryGetTransactionsByUserID, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionsByUserID named query preparation err")
		return nil, 0, err
	}

	query = r.q.Rebind(query)

	if err := r.q.SelectContext(ctx, &transactions, query, args...); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionsByUserID execution err")
		return nil, 0, err
	}

	result := make([]sentrapay.WalletTransaction, 0, len(transactions))
	for _, transaction := range transactions {
		result = append(result, r.makeWalletTransaction(transaction))
	}

	return result, total, nil
}

func (r *walletRepository) makeWalletTransaction(transaction WalletTransactionDB) sentrapay.WalletTransaction {
	return sentrapay.WalletTransaction{
		ID:            transaction.ID.String,
		UserID:        transaction.UserID.String,
		Amount:        transaction.Amount.Float64,
		Type:          transaction.Type.String,
		ReferenceNo:   transaction.ReferenceNo.String,
		PaymentMethod: transaction.PaymentMethod.String,
		Status:        transaction.Status.String,
		BankAccount:   transaction.BankAccount.String,
		BankName:      transaction.BankName.String,
		Description:   transaction.Description.String,
		CreatedAt:     transaction.CreatedAt,
		UpdatedAt:     transaction.UpdatedAt,
	}
}
