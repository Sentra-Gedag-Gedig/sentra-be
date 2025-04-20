package budgetRepository

import (
	"ProjectGolang/internal/api/budget_manager"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

type BudgetTransactionDB struct {
	ID          sql.NullString  `db:"id"`
	UserID      sql.NullString  `db:"user_id"`
	Title       sql.NullString  `db:"title"`
	Description sql.NullString  `db:"description"`
	Nominal     sql.NullFloat64 `db:"nominal"`
	Type        sql.NullString  `db:"type"`
	Category    sql.NullString  `db:"category"`
	AudioLink   sql.NullString  `db:"audio_link"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at"`
}

func (r *budgetRepository) CreateTransaction(c context.Context, transaction entity.BudgetTransaction) error {
	requestID := contextPkg.GetRequestID(c)
	argsKV := map[string]interface{}{
		"id":          transaction.ID,
		"user_id":     transaction.UserID,
		"title":       transaction.Title,
		"description": transaction.Description,
		"nominal":     transaction.Nominal,
		"type":        transaction.Type,
		"category":    transaction.Category,
		"audio_link":  transaction.AudioLink,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
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

	_, err = r.q.ExecContext(c, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Database error when creating transaction")

		return err
	}

	return nil
}

func (r *budgetRepository) GetTransactionByID(c context.Context, id string) (entity.BudgetTransaction, error) {
	requestID := contextPkg.GetRequestID(c)
	var transaction BudgetTransactionDB

	argsKV := map[string]interface{}{
		"id": id,
	}

	query, args, err := sqlx.Named(queryGetTransactionById, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionByID named query preparation err")

		return entity.BudgetTransaction{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(c, query, args...).StructScan(&transaction); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("GetTransactionByID no rows found")
			return entity.BudgetTransaction{}, budget_manager.ErrTransactionNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionByID execution err")
		return entity.BudgetTransaction{}, err
	}

	transactionRes := r.makeBudgetTransaction(transaction)

	return transactionRes, nil
}

func (r *budgetRepository) GetTransactionsByUserID(c context.Context, userID string) ([]entity.BudgetTransaction, error) {
	requestID := contextPkg.GetRequestID(c)
	var transactions []BudgetTransactionDB

	argsKV := map[string]interface{}{
		"user_id": userID,
	}

	query, args, err := sqlx.Named(queryGetTransactionsByUserID, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionsByUserID named query preparation err")
		return nil, err
	}

	query = r.q.Rebind(query)

	if err := r.q.SelectContext(c, &transactions, query, args...); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionsByUserID execution err")
		return nil, err
	}

	result := make([]entity.BudgetTransaction, 0, len(transactions))
	for _, transaction := range transactions {
		result = append(result, r.makeBudgetTransaction(transaction))
	}

	return result, nil
}

func (r *budgetRepository) UpdateTransaction(c context.Context, transaction entity.BudgetTransaction) error {
	requestID := contextPkg.GetRequestID(c)
	argsKV := map[string]interface{}{
		"id":          transaction.ID,
		"title":       transaction.Title,
		"description": transaction.Description,
		"nominal":     transaction.Nominal,
		"type":        transaction.Type,
		"category":    transaction.Category,
		"audio_link":  transaction.AudioLink,
		"updated_at":  time.Now(),
	}

	query, args, err := sqlx.Named(queryUpdateTransaction, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateTransaction named query preparation err")

		return err
	}

	query = r.q.Rebind(query)

	result, err := r.q.ExecContext(c, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateTransaction execution err")

		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateTransaction rows affected err")

		return err
	}

	if rowsAffected == 0 {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
		}).Warn("UpdateTransaction no rows affected")

		return budget_manager.ErrTransactionNotFound
	}

	return nil
}

func (r *budgetRepository) DeleteTransaction(ctx context.Context, id string) error {
	requestID := contextPkg.GetRequestID(ctx)
	argsKV := map[string]interface{}{
		"id": id,
	}

	query, args, err := sqlx.Named(queryDeleteTransaction, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("DeleteTransaction named query preparation err")

		return err
	}

	query = r.q.Rebind(query)

	result, err := r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("DeleteTransaction execution err")

		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("DeleteTransaction rows affected err")

		return err
	}

	if rowsAffected == 0 {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
		}).Warn("DeleteTransaction no rows affected")

		return budget_manager.ErrTransactionNotFound
	}

	return nil
}

func (r *budgetRepository) GetTransactionsByTypeAndCategory(ctx context.Context, userID string, transactionType string, category string) ([]entity.BudgetTransaction, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var transactions []BudgetTransactionDB

	argsKV := map[string]interface{}{
		"user_id":  userID,
		"type":     transactionType,
		"category": category,
	}

	query, args, err := sqlx.Named(queryGetTransactionsByTypeAndCategory, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionsByTypeAndCategory named query preparation err")
		return nil, err
	}

	query = r.q.Rebind(query)

	if err := r.q.SelectContext(ctx, &transactions, query, args...); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetTransactionsByTypeAndCategory execution err")
		return nil, err
	}

	result := make([]entity.BudgetTransaction, 0, len(transactions))
	for _, transaction := range transactions {
		result = append(result, r.makeBudgetTransaction(transaction))
	}

	return result, nil
}

func (r *budgetRepository) GetTransactionsByPeriod(ctx context.Context, userID string, period string) ([]entity.BudgetTransaction, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var transactions []BudgetTransactionDB
	var queryToUse string

	argsKV := map[string]interface{}{
		"user_id": userID,
	}

	switch period {
	case "all":
		queryToUse = queryGetAllTransactions
	case "week":
		queryToUse = queryGetCurrentWeekTransactions
	case "month":
		queryToUse = queryGetCurrentMonthTransactions
	default:
		queryToUse = queryGetAllTransactions
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"period":     period,
		}).Warn("Invalid period provided, defaulting to 'all'")
	}

	query, args, err := sqlx.Named(queryToUse, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"period":     period,
		}).Error("GetTransactionsByPeriod named query preparation err")
		return nil, err
	}

	query = r.q.Rebind(query)

	if err := r.q.SelectContext(ctx, &transactions, query, args...); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"period":     period,
		}).Error("GetTransactionsByPeriod execution err")
		return nil, err
	}

	result := make([]entity.BudgetTransaction, 0, len(transactions))
	for _, transaction := range transactions {
		result = append(result, r.makeBudgetTransaction(transaction))
	}

	return result, nil
}

func (r *budgetRepository) makeBudgetTransaction(transaction BudgetTransactionDB) entity.BudgetTransaction {
	return entity.BudgetTransaction{
		ID:          transaction.ID.String,
		UserID:      transaction.UserID.String,
		Title:       transaction.Title.String,
		Description: transaction.Description.String,
		Nominal:     transaction.Nominal.Float64,
		Type:        transaction.Type.String,
		Category:    transaction.Category.String,
		AudioLink:   transaction.AudioLink.String,
		CreatedAt:   transaction.CreatedAt,
		UpdatedAt:   transaction.UpdatedAt,
	}
}
