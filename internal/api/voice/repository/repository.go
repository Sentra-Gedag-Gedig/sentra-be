package voiceRepository

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
		VoiceCommands: &voiceRepository{q: sqlExecutor, log: r.log},
		Sessions:      &sessionRepository{q: sqlExecutor, log: r.log},
		PageMappings:  &pageMappingRepository{q: sqlExecutor, log: r.log},
		Commit:        commitFunc,
		Rollback:      rollbackFunc,
	}, nil
}

type Client struct {
	VoiceCommands interface {
		CreateVoiceCommand(ctx context.Context, cmd entity.VoiceCommand) error
		GetVoiceCommandByID(ctx context.Context, id string) (entity.VoiceCommand, error)
		GetVoiceCommandsByUserID(ctx context.Context, userID string, limit, offset int) ([]entity.VoiceCommand, int, error)
		UpdateVoiceCommand(ctx context.Context, cmd entity.VoiceCommand) error
		DeleteVoiceCommand(ctx context.Context, id string) error
	}

	Sessions interface {
		CreateSession(ctx context.Context, session entity.VoiceSession) error
		GetSessionByUserID(ctx context.Context, userID string) (entity.VoiceSession, error)
		UpdateSession(ctx context.Context, session entity.VoiceSession) error
		CleanupOldSessions(ctx context.Context) error
	}

	PageMappings interface {
		CreatePageMapping(ctx context.Context, mapping entity.PageMapping) error
		GetPageMappingByID(ctx context.Context, pageID string) (entity.PageMapping, error)
		GetAllPageMappings(ctx context.Context) ([]entity.PageMapping, error)
		UpdatePageMapping(ctx context.Context, mapping entity.PageMapping) error
	}

	Commit   func() error
	Rollback func() error
}

type voiceRepository struct {
	q   SQLExecutor
	log *logrus.Logger
}

type sessionRepository struct {
	q   SQLExecutor
	log *logrus.Logger
}

type pageMappingRepository struct {
	q   SQLExecutor
	log *logrus.Logger
}