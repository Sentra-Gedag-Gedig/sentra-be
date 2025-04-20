package sentrapayRepository

import (
	sentrapay "ProjectGolang/internal/api/sentra_pay"
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
		Wallet:   &walletRepository{q: sqlExecutor, log: r.log},
		Commit:   commitFunc,
		Rollback: rollbackFunc,
	}, nil
}

type Client struct {
	Wallet interface {
		CreateWallet(ctx context.Context, userID string) error
		GetWallet(ctx context.Context, userID string) (sentrapay.WalletBalance, error)
		UpdateWalletBalance(ctx context.Context, userID string, amount float64) error
		CreateTransaction(ctx context.Context, transaction sentrapay.WalletTransaction) error
		GetTransactionByID(ctx context.Context, id string) (sentrapay.WalletTransaction, error)
		GetTransactionByReferenceNo(ctx context.Context, referenceNo string) (sentrapay.WalletTransaction, error)
		UpdateTransactionStatus(ctx context.Context, referenceNo string, status string) error
		GetTransactionsByUserID(ctx context.Context, userID string, limit, offset int) ([]sentrapay.WalletTransaction, int, error)
	}

	Commit   func() error
	Rollback func() error
}

type walletRepository struct {
	q   SQLExecutor
	log *logrus.Logger
}
