package voiceRepository

const (
	queryCreateVoiceCommand = `
		INSERT INTO voice_commands (
			id, user_id, audio_file, transcript, command, 
			response, audio_url, confidence, metadata, 
			created_at, updated_at
		) VALUES (
			:id, :user_id, :audio_file, :transcript, :command,
			:response, :audio_url, :confidence, :metadata,
			:created_at, :updated_at
		)
	`

	queryGetVoiceCommandByID = `
		SELECT 
			id, user_id, audio_file, transcript, command,
			response, audio_url, confidence, metadata,
			created_at, updated_at
		FROM voice_commands 
		WHERE id = :id
	`

	queryGetVoiceCommandsByUserID = `
		SELECT 
			id, user_id, audio_file, transcript, command,
			response, audio_url, confidence, metadata,
			created_at, updated_at
		FROM voice_commands 
		WHERE user_id = :user_id 
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset
	`

	queryCountVoiceCommandsByUserID = `
		SELECT COUNT(*)
		FROM voice_commands 
		WHERE user_id = :user_id
	`

	queryUpdateVoiceCommand = `
		UPDATE voice_commands 
		SET 
			transcript = :transcript,
			command = :command,
			response = :response,
			audio_url = :audio_url,
			confidence = :confidence,
			metadata = :metadata,
			updated_at = :updated_at
		WHERE id = :id
	`

	queryDeleteVoiceCommand = `
		DELETE FROM voice_commands 
		WHERE id = :id
	`

	queryGetVoiceAnalytics = `
		SELECT 
			command,
			confidence,
			metadata,
			created_at
		FROM voice_commands 
		WHERE user_id = :user_id 
		AND created_at >= :start_date
		ORDER BY created_at DESC
		LIMIT :limit
	`

	queryCreateSession = `
		INSERT INTO voice_sessions (
			id, user_id, pending_confirmation, pending_page_id,
			context, created_at, last_activity
		) VALUES (
			:id, :user_id, :pending_confirmation, :pending_page_id,
			:context, :created_at, :last_activity
		)
	`

	queryGetSessionByUserID = `
		SELECT 
			id, user_id, pending_confirmation, pending_page_id,
			context, created_at, last_activity
		FROM voice_sessions 
		WHERE user_id = :user_id
		AND last_activity >= :cutoff_time
		ORDER BY last_activity DESC
		LIMIT 1
	`

	queryUpdateSession = `
		UPDATE voice_sessions 
		SET 
			pending_confirmation = :pending_confirmation,
			pending_page_id = :pending_page_id,
			context = :context,
			last_activity = :last_activity
		WHERE id = :id
	`

	queryDeleteOldSessions = `
		DELETE FROM voice_sessions 
		WHERE last_activity < :cutoff_time
	`

	queryCreatePageMapping = `
		INSERT INTO page_mappings (
			page_id, url, display_name, keywords, synonyms,
			category, description, is_active, created_at, updated_at
		) VALUES (
			:page_id, :url, :display_name, :keywords, :synonyms,
			:category, :description, :is_active, :created_at, :updated_at
		)
	`

	queryGetPageMappingByID = `
		SELECT 
			page_id, url, display_name, keywords, synonyms,
			category, description, is_active, created_at, updated_at
		FROM page_mappings 
		WHERE page_id = :page_id AND is_active = true
	`

	queryGetAllPageMappings = `
		SELECT 
			page_id, url, display_name, keywords, synonyms,
			category, description, is_active, created_at, updated_at
		FROM page_mappings 
		WHERE is_active = true
		ORDER BY category, display_name
	`

	queryUpdatePageMapping = `
		UPDATE page_mappings 
		SET 
			url = :url,
			display_name = :display_name,
			keywords = :keywords,
			synonyms = :synonyms,
			category = :category,
			description = :description,
			is_active = :is_active,
			updated_at = :updated_at
		WHERE page_id = :page_id
	`
)