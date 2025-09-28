package voiceRepository

import (
	"ProjectGolang/internal/api/voice"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/sirupsen/logrus"
)

type VoiceSessionDB struct {
	ID                  sql.NullString `db:"id"`
	UserID              sql.NullString `db:"user_id"`
	PendingConfirmation sql.NullBool   `db:"pending_confirmation"`
	PendingPageID       sql.NullString `db:"pending_page_id"`
	Context             sql.NullString `db:"context"`
	CreatedAt           time.Time      `db:"created_at"`
	LastActivity        time.Time      `db:"last_activity"`
}

type VoiceCommandDB struct {
	ID         sql.NullString  `db:"id"`
	UserID     sql.NullString  `db:"user_id"`
	AudioFile  sql.NullString  `db:"audio_file"`
	Transcript sql.NullString  `db:"transcript"`
	Command    sql.NullString  `db:"command"`
	Response   sql.NullString  `db:"response"`
	AudioURL   sql.NullString  `db:"audio_url"`
	Confidence sql.NullFloat64 `db:"confidence"`
	Metadata   sql.NullString  `db:"metadata"`
	CreatedAt  time.Time       `db:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at"`
}

// Voice Commands Repository Implementation
func (r *voiceRepository) CreateVoiceCommand(ctx context.Context, cmd entity.VoiceCommand) error {
	requestID := contextPkg.GetRequestID(ctx)
	
	metadataJSON, err := json.Marshal(cmd.Metadata)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to marshal metadata")
		return err
	}

	argsKV := map[string]interface{}{
		"id":          cmd.ID,
		"user_id":     cmd.UserID,
		"audio_file":  cmd.AudioFile,
		"transcript":  cmd.Transcript,
		"command":     cmd.Command,
		"response":    cmd.Response,
		"audio_url":   cmd.AudioURL,
		"confidence":  cmd.Confidence,
		"metadata":    string(metadataJSON),
		"created_at":  cmd.CreatedAt,
		"updated_at":  cmd.UpdatedAt,
	}

	query, args, err := sqlx.Named(queryCreateVoiceCommand, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to build SQL query for CreateVoiceCommand")
		return err
	}
	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Database error when creating voice command")
		return err
	}

	return nil
}

func (r *voiceRepository) GetVoiceCommandByID(ctx context.Context, id string) (entity.VoiceCommand, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var cmdDB VoiceCommandDB

	argsKV := map[string]interface{}{
		"id": id,
	}

	query, args, err := sqlx.Named(queryGetVoiceCommandByID, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetVoiceCommandByID named query preparation err")
		return entity.VoiceCommand{}, err
	}

	query = r.q.Rebind(query)

	if err := r.q.QueryRowxContext(ctx, query, args...).StructScan(&cmdDB); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("GetVoiceCommandByID no rows found")
			return entity.VoiceCommand{}, voice.ErrCommandNotRecognized
		}
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetVoiceCommandByID execution err")
		return entity.VoiceCommand{}, err
	}

	return r.makeVoiceCommand(cmdDB), nil
}

func (r *voiceRepository) GetVoiceCommandsByUserID(ctx context.Context, userID string, limit, offset int) ([]entity.VoiceCommand, int, error) {
	requestID := contextPkg.GetRequestID(ctx)
	var commandsList []VoiceCommandDB
	var total int

	// Count total
	countArgsKV := map[string]interface{}{
		"user_id": userID,
	}

	countQuery, countArgs, err := sqlx.Named(queryCountVoiceCommandsByUserID, countArgsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("CountVoiceCommandsByUserID named query preparation err")
		return nil, 0, err
	}

	countQuery = r.q.Rebind(countQuery)

	if err := r.q.QueryRowxContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("CountVoiceCommandsByUserID execution err")
		return nil, 0, err
	}

	// Get commands
	argsKV := map[string]interface{}{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	}

	query, args, err := sqlx.Named(queryGetVoiceCommandsByUserID, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetVoiceCommandsByUserID named query preparation err")
		return nil, 0, err
	}

	query = r.q.Rebind(query)

	if err := r.q.SelectContext(ctx, &commandsList, query, args...); err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("GetVoiceCommandsByUserID execution err")
		return nil, 0, err
	}

	var commands []entity.VoiceCommand
	for _, cmdDB := range commandsList {
		commands = append(commands, r.makeVoiceCommand(cmdDB))
	}

	return commands, total, nil
}

func (r *voiceRepository) UpdateVoiceCommand(ctx context.Context, cmd entity.VoiceCommand) error {
	requestID := contextPkg.GetRequestID(ctx)
	
	metadataJSON, err := json.Marshal(cmd.Metadata)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to marshal metadata")
		return err
	}

	argsKV := map[string]interface{}{
		"id":          cmd.ID,
		"transcript":  cmd.Transcript,
		"command":     cmd.Command,
		"response":    cmd.Response,
		"audio_url":   cmd.AudioURL,
		"confidence":  cmd.Confidence,
		"metadata":    string(metadataJSON),
		"updated_at":  time.Now(),
	}

	query, args, err := sqlx.Named(queryUpdateVoiceCommand, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateVoiceCommand named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	result, err := r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateVoiceCommand execution err")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("UpdateVoiceCommand rows affected err")
		return err
	}

	if rowsAffected == 0 {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         cmd.ID,
		}).Warn("UpdateVoiceCommand no rows affected")
		return voice.ErrCommandNotRecognized
	}

	return nil
}

func (r *voiceRepository) DeleteVoiceCommand(ctx context.Context, id string) error {
	requestID := contextPkg.GetRequestID(ctx)
	argsKV := map[string]interface{}{
		"id": id,
	}

	query, args, err := sqlx.Named(queryDeleteVoiceCommand, argsKV)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("DeleteVoiceCommand named query preparation err")
		return err
	}

	query = r.q.Rebind(query)

	result, err := r.q.ExecContext(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("DeleteVoiceCommand execution err")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("DeleteVoiceCommand rows affected err")
		return err
	}

	if rowsAffected == 0 {
		r.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         id,
		}).Warn("DeleteVoiceCommand no rows affected")
		return voice.ErrCommandNotRecognized
	}

	return nil
}

func (r *voiceRepository) makeVoiceCommand(cmdDB VoiceCommandDB) entity.VoiceCommand {
	var metadata map[string]interface{}
	if cmdDB.Metadata.Valid && cmdDB.Metadata.String != "" {
		json.Unmarshal([]byte(cmdDB.Metadata.String), &metadata)
	}

	return entity.VoiceCommand{
		ID:         cmdDB.ID.String,
		UserID:     cmdDB.UserID.String,
		AudioFile:  cmdDB.AudioFile.String,
		Transcript: cmdDB.Transcript.String,
		Command:    cmdDB.Command.String,
		Response:   cmdDB.Response.String,
		AudioURL:   cmdDB.AudioURL.String,
		Confidence: cmdDB.Confidence.Float64,
		Metadata:   metadata,
		CreatedAt:  cmdDB.CreatedAt,
		UpdatedAt:  cmdDB.UpdatedAt,
	}
}
