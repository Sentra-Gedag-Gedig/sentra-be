package authRepository

import (
	"ProjectGolang/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

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
	var db sqlx.ExtContext
	var commitFunc, rollbackFunc func() error

	db = r.DB

	if tx {
		var err error
		txx, err := r.DB.Beginx()
		if err != nil {
			return Client{}, err
		}

		db = txx
		commitFunc = txx.Commit
		rollbackFunc = txx.Rollback
	} else {
		commitFunc = func() error { return nil }
		rollbackFunc = func() error { return nil }
	}

	return Client{
		Users:    &userRepository{q: db, log: r.log},
		Commit:   commitFunc,
		Rollback: rollbackFunc,
	}, nil
}

type Client struct {
	Users interface {
		CreateUser(ctx context.Context, user entity.User) error
		GetByID(ctx context.Context, id string) (entity.User, error)
		GetByPhoneNumber(ctx context.Context, phoneNumber string) (entity.User, error)
		GetByEmail(ctx context.Context, email string) (entity.User, error)
		UpdateUser(ctx context.Context, user entity.User) error
		UpdateUserPIN(ctx context.Context, phoneNum string, pin string) error
		UpdateUserPassword(ctx context.Context, phoneNum string, password string) error
		DeleteUser(ctx context.Context, id string) error
		EnableTouchID(ctx context.Context, id string, hash string) error
		UpdateProfilePhoto(ctx context.Context, id string, photoURL string) error
		UpdateFacePhoto(ctx context.Context, id string, facePhotoURL string) error
	}

	Commit   func() error
	Rollback func() error
}

type userRepository struct {
	q   sqlx.ExtContext
	log *logrus.Logger
}

type sessionRepository struct {
	q sqlx.ExtContext
}

type userOauthRepository struct {
	q sqlx.ExtContext
}
