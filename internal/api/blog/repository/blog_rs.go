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

type BlogDB struct {
	ID           sql.NullString `db:"id"`
	Title        sql.NullString `db:"title"`
	Body         sql.NullString `db:"body"`
	ImageURL     sql.NullString `db:"image_url"`
	Author       sql.NullString `db:"author"`
	BlogCategory sql.NullString `db:"blog_category"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}

func (r *blogsRepository) CreateBlog(ctx context.Context, blog entity.Blog) error {
	requestID := contextPkg.GetRequestID(ctx)
	argsKV := map[string]interface{}{
		"id":            blog.ID,
		"title":         blog.Title,
		"body":          blog.Body,
		"image_url":     blog.ImageURL,
		"author":        blog.Author,
		"blog_category": blog.BlogCategory,
		"created_at":    blog.CreatedAt,
		"updated_at":    blog.UpdatedAt,
	}

	query, args, err := sqlx.Named(queryCreateBlog, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to build SQL query for CreateBlog")
		return err
	}
	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Database error when creating blog")
		return err
	}

	return nil
}

func (r *blogsRepository) GetBlogByID(ctx context.Context, id string) (entity.Blog, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var blog BlogDB

	argsKV := map[string]interface{}{
		"id": id,
	}

	query, args, err := sqlx.Named(queryGetBlogByID, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetBlogByID named query preparation err")
		return entity.Blog{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(ctx, query, args...).StructScan(&blog); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("GetBlogByID no rows found")
			return entity.Blog{}, blogs.ErrBlogNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetBlogByID execution err")
		return entity.Blog{}, err
	}

	return r.makeBlog(blog), nil
}

func (r *blogsRepository) GetAllBlogs(ctx context.Context, limit, offset int) ([]entity.Blog, int, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var blogsList []BlogDB
	var total int

	countQuery, countArgs, err := sqlx.Named(queryCountAllBlogs, map[string]interface{}{})
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("CountAllBlogs named query preparation err")
		return nil, 0, err
	}

	countQuery = r.q.Rebind(countQuery)

	if err := r.q.QueryRowxContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("CountAllBlogs execution err")
		return nil, 0, err
	}

	argsKV := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}

	query, args, err := sqlx.Named(queryGetAllBlogs, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetAllBlogs named query preparation err")
		return nil, 0, err
	}

	query = r.q.Rebind(query)

	if err := r.q.SelectContext(ctx, &blogsList, query, args...); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetAllBlogs execution err")
		return nil, 0, err
	}

	var blog []entity.Blog
	for _, blogDB := range blogsList {
		blog = append(blog, r.makeBlog(blogDB))
	}

	return blog, total, nil
}

func (r *blogsRepository) GetBlogsByCategory(ctx context.Context, category string, limit, offset int) ([]entity.Blog, int, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var blogsList []BlogDB
	var total int

	countArgsKV := map[string]interface{}{
		"blog_category": category,
	}

	countQuery, countArgs, err := sqlx.Named(queryCountBlogsByCategory, countArgsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("CountBlogsByCategory named query preparation err")
		return nil, 0, err
	}

	countQuery = r.q.Rebind(countQuery)

	if err := r.q.QueryRowxContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("CountBlogsByCategory execution err")
		return nil, 0, err
	}

	argsKV := map[string]interface{}{
		"blog_category": category,
		"limit":         limit,
		"offset":        offset,
	}

	query, args, err := sqlx.Named(queryGetBlogsByCategory, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetBlogsByCategory named query preparation err")
		return nil, 0, err
	}

	query = r.q.Rebind(query)

	if err := r.q.SelectContext(ctx, &blogsList, query, args...); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetBlogsByCategory execution err")
		return nil, 0, err
	}

	var blog []entity.Blog
	for _, blogDB := range blogsList {
		blog = append(blog, r.makeBlog(blogDB))
	}

	return blog, total, nil
}

func (r *blogsRepository) UpdateBlog(ctx context.Context, blog entity.Blog) error {
	requestID := contextPkg.GetRequestID(ctx)
	argsKV := map[string]interface{}{
		"id":            blog.ID,
		"title":         blog.Title,
		"body":          blog.Body,
		"image_url":     blog.ImageURL,
		"blog_category": blog.BlogCategory,
		"updated_at":    time.Now(),
	}

	query, args, err := sqlx.Named(queryUpdateBlog, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateBlog named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	result, err := r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateBlog execution err")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateBlog rows affected err")
		return err
	}

	if rowsAffected == 0 {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         blog.ID,
		}).Warn("UpdateBlog no rows affected")
		return blogs.ErrBlogNotFound
	}

	return nil
}

func (r *blogsRepository) DeleteBlog(ctx context.Context, id string) error {
	requestID := contextPkg.GetRequestID(ctx)
	argsKV := map[string]interface{}{
		"id": id,
	}

	query, args, err := sqlx.Named(queryDeleteBlog, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("DeleteBlog named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	result, err := r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("DeleteBlog execution err")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("DeleteBlog rows affected err")
		return err
	}

	if rowsAffected == 0 {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         id,
		}).Warn("DeleteBlog no rows affected")
		return blogs.ErrBlogNotFound
	}

	return nil
}

func (r *blogsRepository) makeBlog(blog BlogDB) entity.Blog {
	return entity.Blog{
		ID:           blog.ID.String,
		Title:        blog.Title.String,
		Body:         blog.Body.String,
		ImageURL:     blog.ImageURL.String,
		Author:       blog.Author.String,
		BlogCategory: blog.BlogCategory.String,
		CreatedAt:    blog.CreatedAt,
		UpdatedAt:    blog.UpdatedAt,
	}
}
