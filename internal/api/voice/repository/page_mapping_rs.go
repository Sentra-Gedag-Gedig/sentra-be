package voiceRepository

import (
	"ProjectGolang/internal/api/voice"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

type PageMappingDB struct {
	PageID      sql.NullString `db:"page_id"`
	URL         sql.NullString `db:"url"`
	DisplayName sql.NullString `db:"display_name"`
	Keywords    sql.NullString `db:"keywords"`
	Synonyms    sql.NullString `db:"synonyms"`
	Category    sql.NullString `db:"category"`
	Description sql.NullString `db:"description"`
	IsActive    sql.NullBool   `db:"is_active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}
func (r *pageMappingRepository) CreatePageMapping(ctx context.Context, mapping entity.PageMapping) error {
	requestID := contextPkg.GetRequestID(ctx)
	
	keywordsJSON, err := json.Marshal(mapping.Keywords)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to marshal keywords")
		return err
	}

	synonymsJSON, err := json.Marshal(mapping.Synonyms)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to marshal synonyms")
		return err
	}

	argsKV := map[string]interface{}{
		"page_id":      mapping.PageID,
		"url":          mapping.URL,
		"display_name": mapping.DisplayName,
		"keywords":     string(keywordsJSON),
		"synonyms":     string(synonymsJSON),
		"category":     mapping.Category,
		"description":  mapping.Description,
		"is_active":    mapping.IsActive,
		"created_at":   mapping.CreatedAt,
		"updated_at":   mapping.UpdatedAt,
	}

	query, args, err := sqlx.Named(queryCreatePageMapping, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to build SQL query for CreatePageMapping")
		return err
	}
	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Database error when creating page mapping")
		return err
	}

	return nil
}

func (r *pageMappingRepository) GetPageMappingByID(ctx context.Context, pageID string) (entity.PageMapping, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var mappingDB PageMappingDB

	argsKV := map[string]interface{}{
		"page_id": pageID,
	}

	query, args, err := sqlx.Named(queryGetPageMappingByID, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetPageMappingByID named query preparation err")
		return entity.PageMapping{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(ctx, query, args...).StructScan(&mappingDB); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"page_id":    pageID,
			}).Warn("GetPageMappingByID no rows found")
			return entity.PageMapping{}, voice.ErrPageMappingNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetPageMappingByID execution err")
		return entity.PageMapping{}, err
	}

	return r.makePageMapping(mappingDB), nil
}

func (r *pageMappingRepository) GetAllPageMappings(ctx context.Context) ([]entity.PageMapping, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var mappingsList []PageMappingDB

	query, args, err := sqlx.Named(queryGetAllPageMappings, map[string]interface{}{})
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetAllPageMappings named query preparation err")
		return nil, err
	}

	query = r.q.Rebind(query)

	if err := r.q.SelectContext(ctx, &mappingsList, query, args...); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetAllPageMappings execution err")
		return nil, err
	}

	var mappings []entity.PageMapping
	for _, mappingDB := range mappingsList {
		mappings = append(mappings, r.makePageMapping(mappingDB))
	}

	return mappings, nil
}

func (r *pageMappingRepository) UpdatePageMapping(ctx context.Context, mapping entity.PageMapping) error {
	requestID := contextPkg.GetRequestID(ctx)
	
	keywordsJSON, err := json.Marshal(mapping.Keywords)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to marshal keywords")
		return err
	}

	synonymsJSON, err := json.Marshal(mapping.Synonyms)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to marshal synonyms")
		return err
	}

	argsKV := map[string]interface{}{
		"page_id":      mapping.PageID,
		"url":          mapping.URL,
		"display_name": mapping.DisplayName,
		"keywords":     string(keywordsJSON),
		"synonyms":     string(synonymsJSON),
		"category":     mapping.Category,
		"description":  mapping.Description,
		"is_active":    mapping.IsActive,
		"updated_at":   time.Now(),
	}

	query, args, err := sqlx.Named(queryUpdatePageMapping, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdatePageMapping named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	result, err := r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdatePageMapping execution err")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdatePageMapping rows affected err")
		return err
	}

	if rowsAffected == 0 {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"page_id":    mapping.PageID,
		}).Warn("UpdatePageMapping no rows affected")
		return voice.ErrPageMappingNotFound
	}

	return nil
}

func (r *pageMappingRepository) makePageMapping(mappingDB PageMappingDB) entity.PageMapping {
	var keywords []string
	var synonyms []string

	if mappingDB.Keywords.Valid && mappingDB.Keywords.String != "" {
		json.Unmarshal([]byte(mappingDB.Keywords.String), &keywords)
	}

	if mappingDB.Synonyms.Valid && mappingDB.Synonyms.String != "" {
		json.Unmarshal([]byte(mappingDB.Synonyms.String), &synonyms)
	}

	return entity.PageMapping{
		PageID:      mappingDB.PageID.String,
		URL:         mappingDB.URL.String,
		DisplayName: mappingDB.DisplayName.String,
		Keywords:    keywords,
		Synonyms:    synonyms,
		Category:    mappingDB.Category.String,
		Description: mappingDB.Description.String,
		IsActive:    mappingDB.IsActive.Bool,
		CreatedAt:   mappingDB.CreatedAt,
		UpdatedAt:   mappingDB.UpdatedAt,
	}
}