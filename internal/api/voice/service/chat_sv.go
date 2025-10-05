package voiceService

import (
	"ProjectGolang/internal/api/voice"
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

	// Get conversation history from session
	session, err := s.getOrCreateSession(ctx, repo, userID)
	if err != nil {
		return nil, err
	}

	conversationHistory := s.getConversationHistory(session)

	// Process with ChatGPT
	gptResponse, err := s.chatGPT.ProcessConversation(ctx, transcript, conversationHistory)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("ChatGPT processing failed")
		return nil, err
	}

	// Check if GPT detected transaction intent
	txIntent, _ := s.chatGPT.ProcessTransactionIntent(ctx, transcript)
	
	var response *voice.VoiceResponse
	
	if txIntent != nil && txIntent.IsTransaction && txIntent.Confidence > 0.7 {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"is_transaction": txIntent.IsTransaction,
			"type":         txIntent.Type,
			"amount":       txIntent.Amount,
			"confidence":   txIntent.Confidence,
		}).Info("Transaction detected by GPT")

		// Process as transaction
		response, err = s.handleGPTTransaction(ctx, userID, txIntent, session)
		if err != nil {
			return nil, err
		}
	} else {
		// Regular conversation response
		response = &voice.VoiceResponse{
			Text:       gptResponse,
			Action:     "chat",
			Success:    true,
			Confidence: 0.9,
		}
	}

	// ✅ CRITICAL FIX: Generate TTS audio for BOTH transaction and chat responses
	s.log.WithFields(logrus.Fields{
		"request_id": requestID,
		"text":       response.Text,
	}).Debug("Generating TTS audio response")

	audioURL, err := s.generateAudioResponse(response.Text)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to generate TTS audio, continuing without audio")
	} else {
		response.AudioURL = audioURL
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"audio_url":  audioURL,
		}).Info("TTS audio generated and uploaded to S3")
	}

	// Update conversation history in session
	s.updateConversationHistory(session, transcript, gptResponse)

	// Save to database
	commandID, _ := s.utils.NewULIDFromTimestamp(time.Now())
	now := time.Now()
	
	voiceCommand := entity.VoiceCommand{
		ID:         commandID,
		UserID:     userID,
		AudioFile:  audioFilename,
		Transcript: transcript,
		Command:    "chat",
		Response:   response.Text, // Save actual response text
		AudioURL:   response.AudioURL, // Save TTS audio URL
		Confidence: response.Confidence,
		Metadata: map[string]interface{}{
			"mode":           "gpt_chat",
			"session_id":     session.ID,
			"is_transaction": txIntent != nil && txIntent.IsTransaction,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := repo.VoiceCommands.CreateVoiceCommand(ctx, voiceCommand); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to save chat command")
	}

	if err := repo.Sessions.UpdateSession(ctx, *session); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to update session")
	}

	if err := repo.Commit(); err != nil {
		return nil, voice.ErrVoiceCommandFailed
	}

	return response, nil
}

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

	// If category is suggested with high confidence, create directly
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

		// ✅ Response sudah di-return dengan text, TIDAK perlu generate audio disini
		// Audio akan di-generate di ProcessChatCommand
		return response, nil
	}

	// Otherwise, ask for category
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
	// Map GPT category names to entity category constants
	categoryMap := map[string]string{
		// Expense categories
		"makanan":      "makanan",
		"food":         "makanan",
		"transportasi": "transportasi",
		"transport":    "transportasi",
		"belanja":      "belanja",
		"shopping":     "belanja",
		"kesehatan":    "kesehatan",
		"health":       "kesehatan",
		"hiburan":      "hiburan",
		"entertainment": "hiburan",
		"tagihan":      "tagihan",
		"bills":        "tagihan",
		
		// Income categories
		"gaji":       "gaji",
		"salary":     "gaji",
		"bonus":      "bonus",
		"investasi":  "investasi",
		"investment": "investasi",
		"bisnis":     "bisnis",
		"business":   "bisnis",
		"freelance":  "freelance",
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
	
	// Add user message
	history = append(history, chatGPT.ConversationMessage{
		Role:    "user",
		Content: userMsg,
	})
	
	// Add assistant message
	history = append(history, chatGPT.ConversationMessage{
		Role:    "assistant",
		Content: assistantMsg,
	})
	
	// Keep only last 10 messages
	if len(history) > 10 {
		history = history[len(history)-10:]
	}
	
	// Save to session context
	session.Context["conversation_history"] = history
	session.LastActivity = time.Now()
}