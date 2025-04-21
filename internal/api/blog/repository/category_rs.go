package blogRepository

import (
	"ProjectGolang/internal/api/blog"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

type CategoryDB struct {
	ID        sql.NullString `db:"id"`
	Name      sql.NullString `db:"name"`
	CreatedAt time.Time      `db:"created_at"`
}

func (r *categoriesRepository) GetAllCategories(ctx context.Context) ([]entity.BlogCategory, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var categoriesList []CategoryDB

	query, args, err := sqlx.Named(queryGetAllCategories, map[string]interface{}{})
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetAllCategories named query preparation err")
		return nil, err
	}

	query = r.q.Rebind(query)

	if err := r.q.SelectContext(ctx, &categoriesList, query, args...); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetAllCategories execution err")
		return nil, err
	}

	var categories []entity.BlogCategory
	for _, categoryDB := range categoriesList {
		categories = append(categories, r.makeCategory(categoryDB))
	}

	return categories, nil
}

func (r *categoriesRepository) GetCategoryByID(ctx context.Context, id string) (entity.BlogCategory, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var category CategoryDB

	argsKV := map[string]interface{}{
		"id": id,
	}

	query, args, err := sqlx.Named(queryGetCategoryByID, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetCategoryByID named query preparation err")
		return entity.BlogCategory{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(ctx, query, args...).StructScan(&category); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("GetCategoryByID no rows found")
			return entity.BlogCategory{}, blogs.ErrCategoryNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetCategoryByID execution err")
		return entity.BlogCategory{}, err
	}

	return r.makeCategory(category), nil
}

func (r *categoriesRepository) GetCategoryByName(ctx context.Context, name string) (entity.BlogCategory, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var category CategoryDB

	argsKV := map[string]interface{}{
		"name": name,
	}

	query, args, err := sqlx.Named(queryGetCategoryByName, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetCategoryByName named query preparation err")
		return entity.BlogCategory{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(ctx, query, args...).StructScan(&category); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("GetCategoryByName no rows found")
			return entity.BlogCategory{}, blogs.ErrCategoryNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetCategoryByName execution err")
		return entity.BlogCategory{}, err
	}

	return r.makeCategory(category), nil
}

func (r *categoriesRepository) makeCategory(category CategoryDB) entity.BlogCategory {
	return entity.BlogCategory{
		ID:        category.ID.String,
		Name:      category.Name.String,
		CreatedAt: category.CreatedAt,
	}
}
