package voiceService

import (
	"ProjectGolang/internal/api/voice"
	voiceRepository "ProjectGolang/internal/api/voice/repository"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/nlp"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

func (s *voiceService) ProcessVoiceCommand(ctx context.Context, userID string, req voice.ProcessVoiceRequest) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.voiceRepo.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return nil, err
	}
	defer repo.Rollback()

	// Validate audio file
	if err := s.validateAudioFile(req.AudioFile); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Invalid audio file")
		return nil, err
	}

	// Get or create session
	session, err := s.getOrCreateSession(ctx, repo, userID)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get or create session")
		return nil, err
	}

	// Save and process audio file
	audioFilename, err := s.saveAudioFile(req.AudioFile)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to save audio file")
		return nil, err
	}

	// Transcribe audio using OpenAI Whisper
	transcript, err := s.transcribeAudio(audioFilename)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to transcribe audio")
		return nil, voice.ErrTranscriptionFailed
	}

	// Process with NLP
	nlpResult, err := s.nlpProcessor.ProcessCommand(transcript)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to process NLP")
		return nil, voice.ErrNLPProcessingFailed
	}

	// Handle based on confidence and session state
	response := s.handleIntentResult(ctx, nlpResult, transcript, session)

	// Generate audio response
	audioURL, err := s.generateAudioResponse(response.Text)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to generate audio response, continuing without audio")
	} else {
		response.AudioURL = audioURL
	}

	// Save voice command to database
	commandID, err := s.utils.NewULIDFromTimestamp(time.Now())
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to generate command ID")
		return nil, err
	}

	now := time.Now()
	voiceCommand := entity.VoiceCommand{
		ID:         commandID,
		UserID:     userID,
		AudioFile:  audioFilename,
		Transcript: transcript,
		Command:    response.Action,
		Response:   response.Text,
		AudioURL:   response.AudioURL,
		Confidence: response.Confidence,
		Metadata: map[string]interface{}{
			"nlp_result":   nlpResult,
			"session_id":   session.ID,
			"processing_time": time.Since(now).Milliseconds(),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repo.VoiceCommands.CreateVoiceCommand(ctx, voiceCommand); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to save voice command")
		return nil, voice.ErrVoiceCommandFailed
	}

	// Update session state
	s.updateSessionFromResponse(session, response)
	if err := repo.Sessions.UpdateSession(ctx, *session); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to update session")
	}

	if err := repo.Commit(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction")
		return nil, voice.ErrVoiceCommandFailed
	}

	// Set session state in response
	response.SessionState = &voice.SessionState{
		PendingConfirmation: session.PendingConfirmation,
		PendingPageID:       session.PendingPageID,
		Context:             session.Context,
	}

	return response, nil
}

func (s *voiceService) ProcessConfirmation(ctx context.Context, userID string, req voice.ConfirmationRequest) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.voiceRepo.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return nil, err
	}
	defer repo.Rollback()

	// Get session
	session, err := repo.Sessions.GetSessionByUserID(ctx, userID)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get session")
		return nil, voice.ErrSessionNotFound
	}

	if !session.PendingConfirmation || session.PendingPageID == "" {
		return nil, voice.ErrInvalidSession
	}

	// Validate and process confirmation audio
	if err := s.validateAudioFile(req.AudioFile); err != nil {
		return nil, err
	}

	audioFilename, err := s.saveAudioFile(req.AudioFile)
	if err != nil {
		return nil, err
	}

	transcript, err := s.transcribeAudio(audioFilename)
	if err != nil {
		return nil, voice.ErrTranscriptionFailed
	}

	// Check if it's a confirmation
	isConfirmed := s.isConfirmation(transcript)
	
	var response *voice.VoiceResponse

	if isConfirmed {
		// Get page mapping for confirmed page
		pageMapping, err := repo.PageMappings.GetPageMappingByID(ctx, session.PendingPageID)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"page_id":    session.PendingPageID,
				"error":      err.Error(),
			}).Error("Failed to get page mapping")
			return nil, voice.ErrPageMappingNotFound
		}

		responseText := fmt.Sprintf("Baik, menuju ke %s", pageMapping.DisplayName)
		response = &voice.VoiceResponse{
			Text:       responseText,
			Action:     "navigate",
			Target:     pageMapping.URL,
			Success:    true,
			Confidence: 1.0,
		}

		// Clear pending state
		session.PendingConfirmation = false
		session.PendingPageID = ""
	} else {
		responseText := "Baik, silakan ulangi perintah Anda dengan lebih jelas."
		response = &voice.VoiceResponse{
			Text:    responseText,
			Action:  "retry",
			Success: false,
		}

		// Clear pending state
		session.PendingConfirmation = false
		session.PendingPageID = ""
	}

	// Generate audio response
	audioURL, err := s.generateAudioResponse(response.Text)
	if err == nil {
		response.AudioURL = audioURL
	}

	// Update session
	if err := repo.Sessions.UpdateSession(ctx, session); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to update session")
	}

	if err := repo.Commit(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction")
		return nil, voice.ErrVoiceCommandFailed
	}

	response.SessionState = &voice.SessionState{
		PendingConfirmation: session.PendingConfirmation,
		PendingPageID:       session.PendingPageID,
		Context:             session.Context,
	}

	return response, nil
}

// UPDATE GetVoiceHistory method in voice_sv.go
func (s *voiceService) GetVoiceHistory(ctx context.Context, userID string, page, limit int) ([]voice.VoiceCommandHistory, int, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.voiceRepo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	commands, total, err := repo.VoiceCommands.GetVoiceCommandsByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get voice commands")
		return nil, 0, err
	}

	var history []voice.VoiceCommandHistory
	for _, cmd := range commands {
		// Generate presigned URL for response audio if it exists
		audioURL := cmd.AudioURL
		if audioURL != "" {
			presignedURL, err := s.s3Client.PresignUrl(audioURL)
			if err == nil {
				audioURL = presignedURL
			}
		}

		// Also presign the input audio file URL
		inputAudioURL := cmd.AudioFile
		if inputAudioURL != "" {
			presignedURL, err := s.s3Client.PresignUrl(inputAudioURL)
			if err == nil {
				inputAudioURL = presignedURL
			}
		}

		history = append(history, voice.VoiceCommandHistory{
			ID:         cmd.ID,
			UserID:     cmd.UserID,
			AudioFile:  inputAudioURL,
			Transcript: cmd.Transcript,
			Command:    cmd.Command,
			Response:   cmd.Response,
			AudioURL:   audioURL,
			Confidence: cmd.Confidence,
			Metadata:   cmd.Metadata,
			CreatedAt:  cmd.CreatedAt,
			UpdatedAt:  cmd.UpdatedAt,
		})
	}

	return history, total, nil
}

func (s *voiceService) GetVoiceAnalytics(ctx context.Context, userID string) (*voice.VoiceAnalytics, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.voiceRepo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return nil, err
	}

	// Get last 100 commands for analytics
	commands, _, err := repo.VoiceCommands.GetVoiceCommandsByUserID(ctx, userID, 100, 0)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get voice commands for analytics")
		return nil, err
	}

	return s.generateAnalytics(commands), nil
}

// Helper methods
func (s *voiceService) validateAudioFile(file *multipart.FileHeader) error {
	if file == nil {
		return voice.ErrInvalidAudioFile
	}

	if file.Size > s.config.MaxFileSize {
		return voice.ErrAudioFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	for _, allowedExt := range s.config.AllowedFormats {
		if ext == allowedExt {
			return nil
		}
	}

	return voice.ErrUnsupportedFormat
}

func (s *voiceService) getOrCreateSession(ctx context.Context, repo voiceRepository.Client, userID string) (*entity.VoiceSession, error) {
	session, err := repo.Sessions.GetSessionByUserID(ctx, userID)
	if err != nil {
		if err == voice.ErrSessionNotFound {
			// Create new session
			sessionID, _ := s.utils.NewULIDFromTimestamp(time.Now())
			now := time.Now()
			
			session = entity.VoiceSession{
				ID:           sessionID,
				UserID:       userID,
				Context:      make(map[string]interface{}),
				CreatedAt:    now,
				LastActivity: now,
			}

			if err := repo.Sessions.CreateSession(ctx, session); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &session, nil
}

func (s *voiceService) handleUnknownCommand(transcript string) *voice.VoiceResponse {
	suggestions := s.generateSuggestions(transcript)
	
	responseText := fmt.Sprintf("Maaf, saya tidak memahami '%s'.", transcript)
	if len(suggestions) > 0 {
		responseText += " Apakah Anda maksud: " + strings.Join(suggestions, ", ") + "?"
	} else {
		responseText += " Katakan 'bantuan' untuk mendengar daftar perintah yang tersedia."
	}

	return &voice.VoiceResponse{
		Text:       responseText,
		Action:     "clarify",
		Success:    false,
		Confidence: 0.0,
		Metadata: map[string]interface{}{
			"suggestions":        suggestions,
			"need_clarification": true,
		},
	}
}

func (s *voiceService) handleLowConfidence(nlpResult *nlp.IntentResult, transcript string) *voice.VoiceResponse {
	responseText := fmt.Sprintf("Saya kurang yakin dengan perintah '%s'. Bisa dijelaskan lebih spesifik?", transcript)

	return &voice.VoiceResponse{
		Text:       responseText,
		Action:     "confirm",
		Success:    false,
		Confidence: nlpResult.Confidence,
		Metadata: map[string]interface{}{
			"pending_navigation":  nlpResult.Page,
			"need_confirmation":   true,
			"original_transcript": transcript,
		},
	}
}

func (s *voiceService) handleMediumConfidence(nlpResult *nlp.IntentResult) *voice.VoiceResponse {
	responseText := fmt.Sprintf("Menuju ke %s", nlpResult.PageDisplayName)

	return &voice.VoiceResponse{
		Text:       responseText,
		Action:     "navigate",
		Target:     nlpResult.PageURL,
		Success:    true,
		Confidence: nlpResult.Confidence,
		Metadata: map[string]interface{}{
			"page_id":    nlpResult.Page,
			"confidence": nlpResult.Confidence,
		},
	}
}

func (s *voiceService) isConfirmation(transcript string) bool {
	transcript = strings.ToLower(strings.TrimSpace(transcript))
	
	positiveResponses := []string{
		"ya", "iya", "yes", "benar", "betul", "correct", "ok", "oke", 
		"baik", "setuju", "agree", "sip", "yup", "yep",
	}
	
	for _, positive := range positiveResponses {
		if strings.Contains(transcript, positive) {
			return true
		}
	}

	return false
}

func (s *voiceService) generateAnalytics(commands []entity.VoiceCommand) *voice.VoiceAnalytics {
	analytics := &voice.VoiceAnalytics{
		TotalCommands:    len(commands),
		MostUsedCommands: make(map[string]int),
		ConfidenceStats:  make(map[string]float64),
		UsageByTime:      make(map[string]int),
		CategoryUsage:    make(map[string]int),
	}

	if len(commands) == 0 {
		return analytics
	}

	successCount := 0
	confidenceSum := 0.0
	confidenceCount := 0

	for _, cmd := range commands {
		// Command frequency
		analytics.MostUsedCommands[cmd.Command]++
		
		// Success rate
		if cmd.Response != "" && cmd.Confidence > 0.5 {
			successCount++
		}

		// Confidence stats
		if cmd.Confidence > 0 {
			confidenceSum += cmd.Confidence
			confidenceCount++
		}

		// Time-based usage
		hour := cmd.CreatedAt.Hour()
		var timeSlot string
		switch {
		case hour >= 6 && hour < 12:
			timeSlot = "morning"
		case hour >= 12 && hour < 18:
			timeSlot = "afternoon"
		case hour >= 18 && hour < 22:
			timeSlot = "evening"
		default:
			timeSlot = "night"
		}
		analytics.UsageByTime[timeSlot]++
	}

	analytics.SuccessRate = float64(successCount) / float64(len(commands)) * 100
	
	if confidenceCount > 0 {
		analytics.ConfidenceStats["average"] = confidenceSum / float64(confidenceCount)
		analytics.ConfidenceStats["total_samples"] = float64(confidenceCount)
	}

	return analytics
}

func (s *voiceService) generateSuggestions(transcript string) []string {
	var suggestions []string
	
	// Get all page mappings
	allMappings := s.nlpProcessor.GetAllMappings()
	
	// Find similar keywords
	cleanTranscript := strings.ToLower(transcript)
	for _, mapping := range allMappings {
		for _, keyword := range mapping.Keywords {
			if s.calculateSimilarity(cleanTranscript, keyword) > 0.4 {
				suggestions = append(suggestions, mapping.DisplayName)
				break
			}
		}
		
		// Limit suggestions
		if len(suggestions) >= 3 {
			break
		}
	}
	
	return suggestions
}

func (s *voiceService) calculateSimilarity(text1, text2 string) float64 {
	text1 = strings.ToLower(text1)
	text2 = strings.ToLower(text2)
	
	if strings.Contains(text1, text2) || strings.Contains(text2, text1) {
		return 0.8
	}
	
	// Simple similarity - can be enhanced
	return 0.0
}

// Fix metadata handling
func (s *voiceService) updateSessionFromResponse(session *entity.VoiceSession, response *voice.VoiceResponse) {
	if response.Action == "confirm" {
		session.PendingConfirmation = true
		if response.Metadata != nil {
			if pendingNav, ok := response.Metadata["pending_navigation"].(string); ok {
				session.PendingPageID = pendingNav
			}
		}
	} else {
		session.PendingConfirmation = false
		session.PendingPageID = ""
	}

	session.LastActivity = time.Now()
}

// Missing service methods for interface compliance
func (s *voiceService) GetSmartSuggestions(ctx context.Context, userID string) (*voice.SuggestionsResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	_, err := s.voiceRepo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return nil, err
	}

	// Get user context (time of day, recent commands, etc.)
	context := s.getUserContext(userID)
	
	// Generate contextual suggestions
	suggestions := s.getContextualSuggestions(context)

	return &voice.SuggestionsResponse{
		Suggestions: suggestions,
		Context:     context,
	}, nil
}

func (s *voiceService) TestNLPProcessing(ctx context.Context, req voice.NLPTestRequest) (*voice.NLPTestResponse, error) {
	startTime := time.Now()
	
	// Process with NLP
	result, err := s.nlpProcessor.ProcessCommand(req.Text)
	if err != nil {
		return nil, err
	}
	
	processingTime := time.Since(startTime)

	// Convert nlp.MatchResult to voice.MatchResult
	var matches []voice.MatchResult
	for _, match := range result.Matches {
		matches = append(matches, voice.MatchResult{
			Keyword: match.Keyword,
			Score:   match.Score,
			Type:    match.Type,
		})
	}
	
	return &voice.NLPTestResponse{
		Input:      req.Text,
		Intent:     result.Intent,
		Page:       result.Page,
		Confidence: result.Confidence,
		Matches:    matches, // Use converted matches
		Processing: voice.ProcessingDetail{
			CleanedText:    strings.ToLower(req.Text),
			ProcessingTime: processingTime.String(),
		},
	}, nil
}

// Juga perlu fix handleIntentResult method untuk consistency
func (s *voiceService) handleIntentResult(ctx context.Context, nlpResult *nlp.IntentResult, transcript string, session *entity.VoiceSession) *voice.VoiceResponse {
	switch {
	case nlpResult.Intent == "unknown":
		return s.handleUnknownCommand(transcript)
	case nlpResult.Confidence < 0.4:
		return s.handleLowConfidence(nlpResult, transcript)
	case nlpResult.Confidence < 0.7:
		return s.handleMediumConfidence(nlpResult)
	default:
		return s.handleHighConfidence(nlpResult)
	}
}

func (s *voiceService) handleHighConfidence(nlpResult *nlp.IntentResult) *voice.VoiceResponse {
	responseText := fmt.Sprintf("%s. Menuju ke %s.", nlpResult.PageDescription, nlpResult.PageDisplayName)

	// Convert nlp.MatchResult to voice.MatchResult for metadata
	var matches []voice.MatchResult
	for _, match := range nlpResult.Matches {
		matches = append(matches, voice.MatchResult{
			Keyword: match.Keyword,
			Score:   match.Score,
			Type:    match.Type,
		})
	}

	return &voice.VoiceResponse{
		Text:       responseText,
		Action:     "navigate",
		Target:     nlpResult.PageURL,
		Success:    true,
		Confidence: nlpResult.Confidence,
		Metadata: map[string]interface{}{
			"page_id":    nlpResult.Page,
			"confidence": nlpResult.Confidence,
			"matches":    matches, // Use converted matches
		},
	}
}
func (s *voiceService) GetPageMappings(ctx context.Context) ([]voice.PageMapping, error) {
	repo, err := s.voiceRepo.NewClient(false)
	if err != nil {
		return nil, err
	}

	mappings, err := repo.PageMappings.GetAllPageMappings(ctx)
	if err != nil {
		return nil, err
	}

	var result []voice.PageMapping
	for _, mapping := range mappings {
		result = append(result, voice.PageMapping{
			PageID:      mapping.PageID,
			URL:         mapping.URL,
			DisplayName: mapping.DisplayName,
			Keywords:    mapping.Keywords,
			Synonyms:    mapping.Synonyms,
			Category:    mapping.Category,
			Description: mapping.Description,
			IsActive:    mapping.IsActive,
		})
	}

	return result, nil
}

func (s *voiceService) CreatePageMapping(ctx context.Context, mapping voice.PageMapping) error {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.voiceRepo.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}
	defer repo.Rollback()

	now := time.Now()
	entityMapping := entity.PageMapping{
		PageID:      mapping.PageID,
		URL:         mapping.URL,
		DisplayName: mapping.DisplayName,
		Keywords:    mapping.Keywords,
		Synonyms:    mapping.Synonyms,
		Category:    mapping.Category,
		Description: mapping.Description,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := repo.PageMappings.CreatePageMapping(ctx, entityMapping); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create page mapping")
		return err
	}

	// Add to NLP processor
	nlpMapping := nlp.PageMappingData{
		PageID:      mapping.PageID,
		URL:         mapping.URL,
		DisplayName: mapping.DisplayName,
		Keywords:    mapping.Keywords,
		Synonyms:    mapping.Synonyms,
		Category:    mapping.Category,
		Description: mapping.Description,
	}
	s.nlpProcessor.AddPageMapping(mapping.PageID, nlpMapping)

	return repo.Commit()
}

func (s *voiceService) UpdatePageMapping(ctx context.Context, pageID string, mapping voice.PageMapping) error {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.voiceRepo.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}
	defer repo.Rollback()

	entityMapping := entity.PageMapping{
		PageID:      pageID,
		URL:         mapping.URL,
		DisplayName: mapping.DisplayName,
		Keywords:    mapping.Keywords,
		Synonyms:    mapping.Synonyms,
		Category:    mapping.Category,
		Description: mapping.Description,
		IsActive:    mapping.IsActive,
		UpdatedAt:   time.Now(),
	}

	if err := repo.PageMappings.UpdatePageMapping(ctx, entityMapping); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to update page mapping")
		return err
	}

	// Update NLP processor
	nlpMapping := nlp.PageMappingData{
		PageID:      pageID,
		URL:         mapping.URL,
		DisplayName: mapping.DisplayName,
		Keywords:    mapping.Keywords,
		Synonyms:    mapping.Synonyms,
		Category:    mapping.Category,
		Description: mapping.Description,
	}
	s.nlpProcessor.AddPageMapping(pageID, nlpMapping)

	return repo.Commit()
}

// REPLACE ServeAudioFile method
func (s *voiceService) ServeAudioFile(ctx context.Context, filename string) ([]byte, error) {
	// Security check
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		return nil, voice.ErrInvalidAudioFile
	}

	// Get presigned URL from S3
	s3URL := filename
	if !strings.HasPrefix(filename, "http") {
		// If filename is not a full URL, construct S3 URL
		// You might need to adjust this based on your S3 bucket structure
		s3URL = fmt.Sprintf("https://%s.s3.amazonaws.com/%s", 
			os.Getenv("AWS_BUCKET_NAME"), 
			filename)
	}

	presignedURL, err := s.s3Client.PresignUrl(s3URL)
	if err != nil {
		return nil, voice.ErrInvalidAudioFile
	}

	// Download file from S3
	resp, err := http.Get(presignedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download from S3: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, voice.ErrInvalidAudioFile
	}

	// Read and return file content
	return io.ReadAll(resp.Body)
}

// Helper methods
func (s *voiceService) getUserContext(userID string) map[string]interface{} {
	now := time.Now()
	hour := now.Hour()
	
	var timeContext string
	switch {
	case hour >= 5 && hour < 12:
		timeContext = "morning"
	case hour >= 12 && hour < 17:
		timeContext = "afternoon"
	case hour >= 17 && hour < 21:
		timeContext = "evening"
	default:
		timeContext = "night"
	}

	return map[string]interface{}{
		"time_of_day": timeContext,
		"day_of_week": now.Weekday().String(),
		"user_id":     userID,
		"timestamp":   now.Unix(),
	}
}

func (s *voiceService) getContextualSuggestions(context map[string]interface{}) []voice.SmartSuggestion {
	suggestions := []voice.SmartSuggestion{}
	
	timeOfDay, _ := context["time_of_day"].(string)
	
	// Time-based suggestions
	switch timeOfDay {
	case "morning":
		suggestions = append(suggestions,
			voice.SmartSuggestion{
				Text:        "Lihat notifikasi",
				Command:     "notifikasi",
				Description: "Periksa pesan dan update terbaru",
				Priority:    "high",
				Category:    "communication",
			},
			voice.SmartSuggestion{
				Text:        "Cek saldo dompet",
				Command:     "dompet",
				Description: "Periksa saldo dan transaksi hari ini",
				Priority:    "medium",
				Category:    "finance",
			},
		)
	case "afternoon":
		suggestions = append(suggestions,
			voice.SmartSuggestion{
				Text:        "Riwayat transaksi",
				Command:     "riwayat transaksi",
				Description: "Lihat aktivitas keuangan hari ini",
				Priority:    "high",
				Category:    "finance",
			},
		)
	case "evening":
		suggestions = append(suggestions,
			voice.SmartSuggestion{
				Text:        "Pengaturan aplikasi",
				Command:     "pengaturan",
				Description: "Atur preferensi dan konfigurasi",
				Priority:    "medium",
				Category:    "system",
			},
		)
	}

	return suggestions
}

// REPLACE the existing saveAudioFile method with this:
func (s *voiceService) saveAudioFile(audioFile *multipart.FileHeader) (string, error) {
	// Upload directly to S3
	s3URL, err := s.s3Client.UploadFile(audioFile)
	if err != nil {
		return "", fmt.Errorf("failed to upload audio to S3: %w", err)
	}
	
	// Extract filename from S3 URL for reference
	parts := strings.Split(s3URL, "/")
	filename := parts[len(parts)-1]
	
	s.log.WithFields(logrus.Fields{
		"filename": filename,
		"s3_url":   s3URL,
	}).Debug("Audio file uploaded to S3")
	
	return s3URL, nil
}

// UPDATE transcribeAudio to work with S3 URLs
func (s *voiceService) transcribeAudio(s3URL string) (string, error) {
	// Download file from S3 to temporary location for transcription
	presignedURL, err := s.s3Client.PresignUrl(s3URL)
	if err != nil {
		return "", fmt.Errorf("failed to presign S3 URL: %w", err)
	}

	// Download file content
	resp, err := http.Get(presignedURL)
	if err != nil {
		return "", fmt.Errorf("failed to download audio from S3: %w", err)
	}
	defer resp.Body.Close()

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "voice-*.mp3")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy S3 content to temp file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	// Transcribe using OpenAI Whisper
	file, err := os.Open(tmpFile.Name())
	if err != nil {
		return "", err
	}
	defer file.Close()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: tmpFile.Name(),
		Language: "id",
	}

	client := openai.NewClient(s.config.OpenAIAPIKey)
	transcribeResp, err := client.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", err
	}

	return transcribeResp.Text, nil
}

// UPDATE generateAudioResponse to save to S3
func (s *voiceService) generateAudioResponse(text string) (string, error) {
	// Call ElevenLabs API
	url := "https://api.elevenlabs.io/v1/text-to-speech/" + s.config.ElevenLabsVoiceID

	requestBody := map[string]interface{}{
		"text":     text,
		"model_id": "eleven_multilingual_v2",
		"voice_settings": map[string]interface{}{
			"stability":         0.5,
			"similarity_boost":  0.8,
			"style":             0.0,
			"use_speaker_boost": true,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Accept", "audio/mpeg")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("xi-api-key", s.config.ElevenLabsAPIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ElevenLabs API error: %s - %s", resp.Status, string(bodyBytes))
	}

	// Read audio data into memory
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read audio response: %w", err)
	}

	// Generate unique filename
	audioFilename := fmt.Sprintf("tts-%s.mp3", uuid.New().String())

	// Upload directly to S3 using the new method
	s3URL, err := s.s3Client.UploadFileFromBytes(audioFilename, audioData)
	if err != nil {
		return "", fmt.Errorf("failed to upload TTS audio to S3: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"s3_url":   s3URL,
		"filename": audioFilename,
		"size":     len(audioData),
	}).Info("TTS audio uploaded to S3 successfully")

	return s3URL, nil
}

// Remove the uploadFileToS3 helper method - tidak diperlukan lagi

