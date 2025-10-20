
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
func (r *sessionRepository) CreateSession(ctx context.Context, session entity.VoiceSession) error {
	requestID := contextPkg.GetRequestID(ctx)
	
	contextJSON, err := json.Marshal(session.Context)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to marshal session context")
		return err
	}

	argsKV := map[string]interface{}{
		"id":                   session.ID,
		"user_id":              session.UserID,
		"pending_confirmation": session.PendingConfirmation,
		"pending_page_id":      session.PendingPageID,
		"context":              string(contextJSON),
		"created_at":           session.CreatedAt,
		"last_activity":        session.LastActivity,
	}

	query, args, err := sqlx.Named(queryCreateSession, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to build SQL query for CreateSession")
		return err
	}
	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Database error when creating session")
		return err
	}

	return nil
}

func (r *sessionRepository) GetSessionByUserID(ctx context.Context, userID string) (entity.VoiceSession, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var sessionDB VoiceSessionDB

	cutoffTime := time.Now().Add(-24 * time.Hour) 

	argsKV := map[string]interface{}{
		"user_id":     userID,
		"cutoff_time": cutoffTime,
	}

	query, args, err := sqlx.Named(queryGetSessionByUserID, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetSessionByUserID named query preparation err")
		return entity.VoiceSession{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(ctx, query, args...).StructScan(&sessionDB); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"user_id":    userID,
			}).Debug("GetSessionByUserID no active session found")
			return entity.VoiceSession{}, voice.ErrSessionNotFound
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetSessionByUserID execution err")
		return entity.VoiceSession{}, err
	}

	return r.makeVoiceSession(sessionDB), nil
}



func (r *sessionRepository) UpdateSession(ctx context.Context, session entity.VoiceSession) error {
	requestID := contextPkg.GetRequestID(ctx)
	
	
	contextJSON, err := json.Marshal(session.Context)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
			"context":    session.Context, 
		}).Error("Failed to marshal session context")
		return err
	}

	
	r.log.WithFields(logrus.Fields{
		"request_id": requestID,
		"context_json": string(contextJSON),
	}).Debug("Session context JSON")

	argsKV := map[string]interface{}{
		"id":                   session.ID,
		"pending_confirmation": session.PendingConfirmation,
		"pending_page_id":      session.PendingPageID,
		"context":              string(contextJSON),
		"last_activity":        time.Now(),
	}

	query, args, err := sqlx.Named(queryUpdateSession, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateSession named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	result, err := r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateSession execution err")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateSession rows affected err")
		return err
	}

	if rowsAffected == 0 {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         session.ID,
		}).Warn("UpdateSession no rows affected")
		return voice.ErrSessionNotFound
	}

	return nil
}

func (r *sessionRepository) CleanupOldSessions(ctx context.Context) error {
	requestID := contextPkg.GetRequestID(ctx)
	cutoffTime := time.Now().Add(-24 * time.Hour)

	argsKV := map[string]interface{}{
		"cutoff_time": cutoffTime,
	}

	query, args, err := sqlx.Named(queryDeleteOldSessions, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("CleanupOldSessions named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	result, err := r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("CleanupOldSessions execution err")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil {
		r.log.WithFields(logrus.Fields{
			"request_id":    requestID,
			"rows_affected": rowsAffected,
		}).Info("Cleaned up old sessions")
	}

	return err
}

func (r *sessionRepository) makeVoiceSession(sessionDB VoiceSessionDB) entity.VoiceSession {
	var context map[string]interface{}
	if sessionDB.Context.Valid && sessionDB.Context.String != "" {
		json.Unmarshal([]byte(sessionDB.Context.String), &context)
	}

	return entity.VoiceSession{
		ID:                  sessionDB.ID.String,
		UserID:              sessionDB.UserID.String,
		PendingConfirmation: sessionDB.PendingConfirmation.Bool,
		PendingPageID:       sessionDB.PendingPageID.String,
		Context:             context,
		CreatedAt:           sessionDB.CreatedAt,
		LastActivity:        sessionDB.LastActivity,
	}
}