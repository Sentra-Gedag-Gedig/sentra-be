package budgetRepository

import (
	"ProjectGolang/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type SQLExecutor interface {
	sqlx.ExtContext
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	Rebind(query string) string
}

func New(db *sqlx.DB, log *logrus.Logger) Repository {
	return &repository{
		DB:  db,
		log: log,
	}
}

type repository struct {
	DB  *sqlx.DB
	log *logrus.Logger
}

type Repository interface {
	NewClient(tx bool) (Client, error)
}

func (r *repository) NewClient(tx bool) (Client, error) {
	var sqlExecutor SQLExecutor
	var commitFunc, rollbackFunc func() error

	sqlExecutor = r.DB

	if tx {
		var err error
		txx, err := r.DB.Beginx()
		if err != nil {
			return Client{}, err
		}

		sqlExecutor = txx
		commitFunc = txx.Commit
		rollbackFunc = txx.Rollback
	} else {
		commitFunc = func() error { return nil }
		rollbackFunc = func() error { return nil }
	}

	return Client{
		Budget:   &budgetRepository{q: sqlExecutor, log: r.log},
		Commit:   commitFunc,
		Rollback: rollbackFunc,
	}, nil
}

type Client struct {
	Budget interface {
		CreateTransaction(c context.Context, transaction entity.BudgetTransaction) error
		GetTransactionByID(c context.Context, id string) (entity.BudgetTransaction, error)
		GetTransactionsByUserID(c context.Context, userID string) ([]entity.BudgetTransaction, error)
		GetTransactionsByPeriod(ctx context.Context, userID string, period string) ([]entity.BudgetTransaction, error)
		UpdateTransaction(c context.Context, transaction entity.BudgetTransaction) error
		DeleteTransaction(ctx context.Context, id string) error
		GetTransactionsByTypeAndCategory(ctx context.Context, userID string, transactionType string, category string) ([]entity.BudgetTransaction, error)
	}

	Commit   func() error
	Rollback func() error
}

type budgetRepository struct {
	q   SQLExecutor
	log *logrus.Logger
}
