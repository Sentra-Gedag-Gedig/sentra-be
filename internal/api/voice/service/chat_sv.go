package voiceService

import (
	"ProjectGolang/internal/api/voice"
	voiceRepository "ProjectGolang/internal/api/voice/repository"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/nlp"
	chatGPT "ProjectGolang/pkg/openai"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func (s *voiceService) ProcessChatCommand(
	ctx context.Context,
	userID string,
	req voice.ProcessVoiceRequest,
) (*voice.VoiceResponse, error) {
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

	// Validate and transcribe audio
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

	s.log.WithFields(logrus.Fields{
		"request_id": requestID,
		"transcript": transcript,
	}).Info("Chat transcript received")

	// Get or create session
	session, err := s.getOrCreateSession(ctx, repo, userID)
	if err != nil {
		return nil, err
	}

	// Check if user is in confirmation state
	if session.PendingConfirmation {
		response, err := s.handleConfirmationState(ctx, repo, userID, transcript, session)
		if err != nil {
			return nil, err
		}
		
		// Generate audio for response
		audioURL, err := s.generateAudioResponse(response.Text)
		if err == nil {
			response.AudioURL = audioURL
		}
		
		response.Transcript = transcript
		
		// Save command to database
		if err := s.saveVoiceCommand(ctx, repo, userID, audioFilename, transcript, response); err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("Failed to save voice command")
		}
		
		if err := repo.Commit(); err != nil {
			return nil, err
		}
		
		return response, nil
	}

	// Get available pages for navigation
	availablePages, err := s.getAvailablePagesForGPT(ctx, repo)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to get available pages, continuing without navigation")
		availablePages = []chatGPT.PageInfo{}
	}

	// Process with multi-intent detection
	multiIntent, err := s.chatGPT.ProcessMultiIntent(ctx, transcript, availablePages)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Multi-intent processing failed")
		return nil, err
	}

	// Handle clarification needed
	if multiIntent.NeedsClarification {
		response := &voice.VoiceResponse{
			Text:       multiIntent.ClarificationQuestion,
			Transcript: transcript,
			Action:     "clarify",
			Success:    false,
			Confidence: multiIntent.Confidence,
		}
		
		audioURL, err := s.generateAudioResponse(response.Text)
		if err == nil {
			response.AudioURL = audioURL
		}
		
		if err := s.saveVoiceCommand(ctx, repo, userID, audioFilename, transcript, response); err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("Failed to save voice command")
		}
		
		if err := repo.Commit(); err != nil {
			return nil, err
		}
		
		return response, nil
	}

	// Process intents
	response, err := s.processMultiIntents(ctx, repo, userID, multiIntent, session, transcript)
	if err != nil {
		return nil, err
	}

	// Generate audio response
	audioURL, err := s.generateAudioResponse(response.Text)
	if err == nil {
		response.AudioURL = audioURL
	}

	response.Transcript = transcript

	// Save voice command
	if err := s.saveVoiceCommand(ctx, repo, userID, audioFilename, transcript, response); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to save voice command")
	}

	// Update session
	s.updateConversationHistory(session, transcript, response.Text)
	if err := repo.Sessions.UpdateSession(ctx, *session); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to update session")
	}

	if err := repo.Commit(); err != nil {
		return nil, err
	}

	return response, nil
}

// Process multiple intents in order
func (s *voiceService) processMultiIntents(
	ctx context.Context,
	repo voiceRepository.Client,  
	userID string,
	multiIntent *chatGPT.MultiIntentResult,
	session *entity.VoiceSession,
	transcript string,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	if len(multiIntent.Intents) == 0 {
		// No intent detected, treat as general query
		conversationHistory := s.getConversationHistory(session)
		gptResponse, err := s.chatGPT.ProcessConversation(ctx, transcript, conversationHistory)
		if err != nil {
			return nil, err
		}

		return &voice.VoiceResponse{
			Text:       gptResponse,
			Action:     "chat",
			Success:    true,
			Confidence: multiIntent.Confidence,
		}, nil
	}

	var responses []string
	var finalAction string
	var finalTarget string
	var needsConfirmation bool
	var pendingContext map[string]interface{}
	successCount := 0

	// Process each intent in order
	for _, intent := range multiIntent.Intents {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"intent_type": intent.Type,
			"action":      intent.Action,
			"order":       intent.Order,
			"confidence":  intent.Confidence,
		}).Info("Processing intent")

		switch intent.Type {
		case "navigation":
			response, err := s.handleNavigationIntent(ctx, intent, session)
			if err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err.Error(),
				}).Error("Failed to handle navigation intent")
				responses = append(responses, "Maaf, gagal membuka halaman yang diminta.")
				continue
			}

			if response.Success {
				responses = append(responses, response.Text)
				finalAction = "navigate"
				finalTarget = response.Target
				successCount++
			} else {
				responses = append(responses, response.Text)
				needsConfirmation = response.SessionState != nil && response.SessionState.PendingConfirmation
				if needsConfirmation {
					pendingContext = response.SessionState.Context
				}
			}

		case "transaction":
			response, err := s.handleTransactionIntent(ctx, userID, intent, session)
			if err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err.Error(),
				}).Error("Failed to handle transaction intent")
				responses = append(responses, "Maaf, gagal mencatat transaksi.")
				continue
			}

			responses = append(responses, response.Text)
			if response.Success {
				successCount++
				if finalAction == "" {
					finalAction = "transaction"
				}
			} else {
				needsConfirmation = response.SessionState != nil && response.SessionState.PendingConfirmation
				if needsConfirmation {
					pendingContext = response.SessionState.Context
				}
			}

		case "query":
			conversationHistory := s.getConversationHistory(session)
			gptResponse, err := s.chatGPT.ProcessConversation(ctx, transcript, conversationHistory)
			if err != nil {
				responses = append(responses, "Maaf, saya tidak dapat memproses permintaan Anda.")
				continue
			}
			responses = append(responses, gptResponse)
			if finalAction == "" {
				finalAction = "chat"
			}
			successCount++
		}
	}

	// Build final response
	finalText := strings.Join(responses, " ")
	
	if finalAction == "" {
		finalAction = "chat"
	}

	finalResponse := &voice.VoiceResponse{
		Text:       finalText,
		Action:     finalAction,
		Target:     finalTarget,
		Success:    successCount > 0,
		Confidence: multiIntent.Confidence,
		Metadata: map[string]interface{}{
			"intents_processed": len(multiIntent.Intents),
			"success_count":     successCount,
		},
	}

	// Set session state if confirmation needed
	if needsConfirmation {
		session.PendingConfirmation = true
		session.Context = pendingContext
		finalResponse.SessionState = &voice.SessionState{
			PendingConfirmation: true,
			Context:             pendingContext,
		}
	} else {
		session.PendingConfirmation = false
		session.Context = make(map[string]interface{})
	}

	return finalResponse, nil
}

// Handle navigation intent
func (s *voiceService) handleNavigationIntent(
	ctx context.Context,
	intent chatGPT.Intent,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	// Extract navigation data
	pageID, _ := intent.Data["page_id"].(string)
	url, _ := intent.Data["url"].(string)
	displayName, _ := intent.Data["display_name"].(string)

	if pageID == "" || url == "" {
		return &voice.VoiceResponse{
			Text:    "Maaf, halaman yang Anda maksud tidak jelas. Bisa sebutkan lagi?",
			Action:  "clarify",
			Success: false,
		}, nil
	}

	s.log.WithFields(logrus.Fields{
		"request_id":   requestID,
		"page_id":      pageID,
		"url":          url,
		"display_name": displayName,
		"confidence":   intent.Confidence,
	}).Info("Navigating to page")

	// Check confidence
	if intent.Confidence < 0.7 {
		// Ask for confirmation
		session.PendingConfirmation = true
		session.PendingPageID = pageID
		session.Context = map[string]interface{}{
			"step":         "awaiting_navigation_confirmation",
			"pending_page": map[string]interface{}{
				"page_id":      pageID,
				"url":          url,
				"display_name": displayName,
			},
		}

		responseText := fmt.Sprintf("Apakah Anda ingin membuka halaman %s?", displayName)
		return &voice.VoiceResponse{
			Text:       responseText,
			Action:     "confirm_navigation",
			Success:    false,
			Confidence: intent.Confidence,
			SessionState: &voice.SessionState{
				PendingConfirmation: true,
				PendingPageID:       pageID,
				Context:             session.Context,
			},
		}, nil
	}

	// High confidence, navigate directly
	responseText := fmt.Sprintf("Baik, membuka halaman %s.", displayName)
	return &voice.VoiceResponse{
		Text:       responseText,
		Action:     "navigate",
		Target:     url,
		Success:    true,
		Confidence: intent.Confidence,
		Metadata: map[string]interface{}{
			"page_id":      pageID,
			"display_name": displayName,
		},
	}, nil
}

// Handle transaction intent
func (s *voiceService) handleTransactionIntent(
	ctx context.Context,
	userID string,
	intent chatGPT.Intent,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	// Extract transaction data
	amount, _ := intent.Data["amount"].(float64)
	description, _ := intent.Data["description"].(string)
	category, _ := intent.Data["category"].(string)
	txType, _ := intent.Data["type"].(string)

	s.log.WithFields(logrus.Fields{
		"request_id":  requestID,
		"amount":      amount,
		"description": description,
		"category":    category,
		"type":        txType,
		"confidence":  intent.Confidence,
	}).Info("Processing transaction intent")

	// Validate required fields
	if amount == 0 || txType == "" {
		return &voice.VoiceResponse{
			Text:    "Maaf, informasi transaksi tidak lengkap. Bisa ulangi dengan menyebutkan nominal dan jenisnya?",
			Action:  "clarify",
			Success: false,
		}, nil
	}

	// If category not provided or low confidence, ask for category
	if category == "" || intent.Confidence < 0.8 {
		session.PendingConfirmation = true
		session.Context = map[string]interface{}{
			"step":               "awaiting_category",
			"transaction_type":   txType,
			"transaction_amount": amount,
			"transaction_desc":   description,
		}

		categories := s.getAvailableCategories(txType)
		categoriesText := strings.Join(categories, ", ")

		typeText := "pengeluaran"
		if txType == "income" {
			typeText = "pemasukan"
		}

		responseText := fmt.Sprintf(
			"Saya catat %s Rp%.0f untuk %s. Kategori apa? Pilihan: %s",
			typeText,
			amount,
			description,
			categoriesText,
		)

		return &voice.VoiceResponse{
			Text:    responseText,
			Action:  "ask_category",
			Success: false,
			SessionState: &voice.SessionState{
				PendingConfirmation: true,
				Context:             session.Context,
			},
		}, nil
	}

	// Create transaction directly
	txData := &nlp.TransactionData{
		Type:        txType,
		Amount:      amount,
		Description: description,
		Category:    category,
		Confidence:  intent.Confidence,
	}

	response, err := s.createBudgetTransaction(ctx, userID, txData)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Handle confirmation state (when user is confirming something)
// Line ~245 di chat_sv.go
// Handle confirmation state (when user is confirming something)
func (s *voiceService) handleConfirmationState(
	ctx context.Context,
	repo voiceRepository.Client,
	userID string,
	transcript string,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	step, _ := session.Context["step"].(string)

	s.log.WithFields(logrus.Fields{
		"request_id": requestID,
		"step":       step,
		"transcript": transcript,
	}).Info("Handling confirmation state")

	switch step {
	case "awaiting_category":
		return s.handleCategoryConfirmation(ctx, userID, transcript, session)

	case "awaiting_navigation_confirmation":
		return s.handleNavigationConfirmation(ctx, transcript, session)

	default:
		// Unknown step, clear pending state
		session.PendingConfirmation = false
		session.Context = make(map[string]interface{})

		return &voice.VoiceResponse{
			Text:    "Maaf, saya lupa apa yang sedang kita bicarakan. Silakan ulangi perintah Anda.",
			Action:  "retry",
			Success: false,
		}, nil
	}
}

// Handle category confirmation for transaction
func (s *voiceService) handleCategoryConfirmation(
	ctx context.Context,
	userID string,
	transcript string,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	// Extract pending transaction data
	txType, ok := session.Context["transaction_type"].(string)
	if !ok {
		return &voice.VoiceResponse{
			Text:    "Maaf, terjadi kesalahan. Data transaksi tidak ditemukan.",
			Action:  "error",
			Success: false,
		}, nil
	}

	amount, ok := session.Context["transaction_amount"].(float64)
	if !ok {
		return &voice.VoiceResponse{
			Text:    "Maaf, terjadi kesalahan. Nominal transaksi tidak valid.",
			Action:  "error",
			Success: false,
		}, nil
	}

	description, _ := session.Context["transaction_desc"].(string)
	if description == "" {
		description = "Transaksi" // default description
	}

	// Normalize category from user response
	category := s.normalizeCategoryName(transcript)

	// Validate category
	validCategories := s.getAvailableCategories(txType)
	isValid := false
	for _, validCat := range validCategories {
		if category == validCat {
			isValid = true
			break
		}
	}

	if !isValid {
		categoriesText := strings.Join(validCategories, ", ")
		responseText := fmt.Sprintf(
			"Kategori '%s' tidak valid. Pilih dari: %s",
			transcript,
			categoriesText,
		)

		return &voice.VoiceResponse{
			Text:    responseText,
			Action:  "ask_category",
			Success: false,
			SessionState: &voice.SessionState{
				PendingConfirmation: true,
				Context:             session.Context,
			},
		}, nil
	}

	// Create transaction data
	txData := &nlp.TransactionData{
		Type:        txType,
		Amount:      amount,
		Description: description,
		Category:    category,
		Confidence:  0.95,
	}

	// Create transaction
	response, err := s.createBudgetTransaction(ctx, userID, txData)
	if err != nil {
		return nil, err
	}

	// Clear pending state
	session.PendingConfirmation = false
	session.Context = make(map[string]interface{})

	return response, nil
}

// Handle navigation confirmation
func (s *voiceService) handleNavigationConfirmation(
	ctx context.Context,
	transcript string,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	// Check if user confirmed
	isConfirmed := s.isConfirmation(transcript)

	if !isConfirmed {
		// User declined
		session.PendingConfirmation = false
		session.PendingPageID = ""
		session.Context = make(map[string]interface{})

		return &voice.VoiceResponse{
			Text:    "Baik, dibatalkan. Ada yang bisa saya bantu?",
			Action:  "cancelled",
			Success: false,
		}, nil
	}

	// Get pending page data
	pendingPage, ok := session.Context["pending_page"].(map[string]interface{})
	if !ok {
		session.PendingConfirmation = false
		session.Context = make(map[string]interface{})

		return &voice.VoiceResponse{
			Text:    "Maaf, terjadi kesalahan. Silakan ulangi perintah Anda.",
			Action:  "error",
			Success: false,
		}, nil
	}

	url, _ := pendingPage["url"].(string)
	displayName, _ := pendingPage["display_name"].(string)

	// Clear pending state
	session.PendingConfirmation = false
	session.PendingPageID = ""
	session.Context = make(map[string]interface{})

	responseText := fmt.Sprintf("Baik, membuka halaman %s.", displayName)
	return &voice.VoiceResponse{
		Text:       responseText,
		Action:     "navigate",
		Target:     url,
		Success:    true,
		Confidence: 1.0,
		Metadata: map[string]interface{}{
			"display_name": displayName,
		},
	}, nil
}

// Get available pages for GPT context
func (s *voiceService) getAvailablePagesForGPT(
	ctx context.Context,
	repo voiceRepository.Client,  // ✅ Gunakan Client type langsung
) ([]chatGPT.PageInfo, error) {
	mappings, err := repo.PageMappings.GetAllPageMappings(ctx)
	if err != nil {
		return nil, err
	}

	var pages []chatGPT.PageInfo
	for _, mapping := range mappings {
		if !mapping.IsActive {
			continue
		}

		pages = append(pages, chatGPT.PageInfo{
			PageID:      mapping.PageID,
			URL:         mapping.URL,
			DisplayName: mapping.DisplayName,
			Keywords:    mapping.Keywords,
			Category:    mapping.Category,
			Description: mapping.Description,
		})
	}

	return pages, nil
}

// Line ~307 di chat_sv.go
func (s *voiceService) saveVoiceCommand(
	ctx context.Context,
	repo voiceRepository.Client,  // ✅ Ganti dari interface{}
	userID string,
	audioFile string,
	transcript string,
	response *voice.VoiceResponse,
) error {
	commandID, err := s.utils.NewULIDFromTimestamp(time.Now())
	if err != nil {
		return err
	}

	now := time.Now()
	voiceCommand := entity.VoiceCommand{
		ID:         commandID,
		UserID:     userID,
		AudioFile:  audioFile,
		Transcript: transcript,
		Command:    response.Action,
		Response:   response.Text,
		AudioURL:   response.AudioURL,
		Confidence: response.Confidence,
		Metadata:   response.Metadata,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return repo.VoiceCommands.CreateVoiceCommand(ctx, voiceCommand)
}

// Existing helper methods from previous chat_sv.go
func (s *voiceService) handleGPTTransaction(
	ctx context.Context,
	userID string,
	txIntent *chatGPT.TransactionIntent,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	s.log.WithFields(logrus.Fields{
		"request_id": requestID,
		"tx_type":    txIntent.Type,
		"amount":     txIntent.Amount,
		"category":   txIntent.SuggestedCategory,
	}).Info("Processing GPT transaction")

	if txIntent.SuggestedCategory != "" && txIntent.Confidence > 0.8 {
		txData := &nlp.TransactionData{
			Type:        txIntent.Type,
			Amount:      txIntent.Amount,
			Description: txIntent.Description,
			Category:    txIntent.SuggestedCategory,
			Confidence:  txIntent.Confidence,
		}
		
		response, err := s.createBudgetTransaction(ctx, userID, txData)
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	session.Context = map[string]interface{}{
		"step":               "awaiting_category",
		"transaction_type":   txIntent.Type,
		"transaction_amount": txIntent.Amount,
		"transaction_desc":   txIntent.Description,
	}
	session.PendingConfirmation = true

	categories := s.getAvailableCategories(txIntent.Type)
	categoriesText := strings.Join(categories, ", ")

	responseText := fmt.Sprintf(
		"Saya catat %s Rp%.0f untuk %s. Kategori apa? Pilihan: %s",
		txIntent.Type,
		txIntent.Amount,
		txIntent.Description,
		categoriesText,
	)

	return &voice.VoiceResponse{
		Text:    responseText,
		Action:  "ask_category",
		Success: false,
		SessionState: &voice.SessionState{
			PendingConfirmation: true,
			Context:             session.Context,
		},
	}, nil
}

func (s *voiceService) normalizeCategoryName(category string) string {
	categoryMap := map[string]string{
		"makanan":       "makanan",
		"food":          "makanan",
		"transportasi":  "transportasi",
		"transport":     "transportasi",
		"belanja":       "belanja",
		"shopping":      "belanja",
		"kesehatan":     "kesehatan",
		"health":        "kesehatan",
		"hiburan":       "hiburan",
		"entertainment": "hiburan",
		"tagihan":       "tagihan",
		"bills":         "tagihan",
		"gaji":          "gaji",
		"salary":        "gaji",
		"bonus":         "bonus",
		"investasi":     "investasi",
		"investment":    "investasi",
		"bisnis":        "bisnis",
		"business":      "bisnis",
		"freelance":     "freelance",
	}

	normalized := strings.ToLower(strings.TrimSpace(category))
	if mapped, exists := categoryMap[normalized]; exists {
		return mapped
	}

	return normalized
}

func (s *voiceService) getTypeText(txType string) string {
	if txType == "income" {
		return "Pemasukan"
	}
	return "Pengeluaran"
}

func (s *voiceService) getConversationHistory(session *entity.VoiceSession) []chatGPT.ConversationMessage {
	history := []chatGPT.ConversationMessage{}
	
	if session.Context["conversation_history"] != nil {
		if hist, ok := session.Context["conversation_history"].([]interface{}); ok {
			for _, msg := range hist {
				if m, ok := msg.(map[string]interface{}); ok {
					history = append(history, chatGPT.ConversationMessage{
						Role:    m["role"].(string),
						Content: m["content"].(string),
					})
				}
			}
		}
	}
	
	return history
}
func (s *voiceService) updateConversationHistory(session *entity.VoiceSession, userMsg, assistantMsg string) {
	history := s.getConversationHistory(session)
	
	history = append(history, chatGPT.ConversationMessage{
		Role:    "user",
		Content: userMsg,
	})
	
	history = append(history, chatGPT.ConversationMessage{
		Role:    "assistant",
		Content: assistantMsg,
	})
	
	// Keep only last 10 messages (5 minutes session = enough for short conversation)
	if len(history) > 10 {
		history = history[len(history)-10:]
	}
	
	session.Context["conversation_history"] = history
	session.LastActivity = time.Now()
} // ✅ ADD THIS CLOSING BRACE
