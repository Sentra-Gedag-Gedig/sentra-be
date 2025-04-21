package blogRepository

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
		Blogs:      &blogsRepository{q: sqlExecutor, log: r.log},
		Categories: &categoriesRepository{q: sqlExecutor, log: r.log},
		Commit:     commitFunc,
		Rollback:   rollbackFunc,
	}, nil
}

type Client struct {
	Blogs interface {
		CreateBlog(ctx context.Context, blog entity.Blog) error
		GetBlogByID(ctx context.Context, id string) (entity.Blog, error)
		GetAllBlogs(ctx context.Context, limit, offset int) ([]entity.Blog, int, error)
		GetBlogsByCategory(ctx context.Context, category string, limit, offset int) ([]entity.Blog, int, error)
		UpdateBlog(ctx context.Context, blog entity.Blog) error
		DeleteBlog(ctx context.Context, id string) error
	}

	Categories interface {
		GetAllCategories(ctx context.Context) ([]entity.BlogCategory, error)
		GetCategoryByID(ctx context.Context, id string) (entity.BlogCategory, error)
		GetCategoryByName(ctx context.Context, name string) (entity.BlogCategory, error)
	}

	Commit   func() error
	Rollback func() error
}

type blogsRepository struct {
	q   SQLExecutor
	log *logrus.Logger
}

type categoriesRepository struct {
	q   SQLExecutor
	log *logrus.Logger
}
