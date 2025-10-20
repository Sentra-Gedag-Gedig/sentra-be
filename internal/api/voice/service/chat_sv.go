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

	
	session, err := s.getOrCreateSession(ctx, repo, userID)
	if err != nil {
		return nil, err
	}

	
	if session.PendingConfirmation {
		response, err := s.handleConfirmationState(ctx, repo, userID, transcript, session)
		if err != nil {
			return nil, err
		}
		
		
		audioURL, err := s.generateAudioResponse(response.Text)
		if err == nil {
			response.AudioURL = audioURL
		}
		
		response.Transcript = transcript
		
		
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

	
	availablePages, err := s.getAvailablePagesForGPT(ctx, repo)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to get available pages, continuing without navigation")
		availablePages = []chatGPT.PageInfo{}
	}

	
	multiIntent, err := s.chatGPT.ProcessMultiIntent(ctx, transcript, availablePages)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Multi-intent processing failed")
		return nil, err
	}

	
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

	
	response, err := s.processMultiIntents(ctx, repo, userID, multiIntent, session, transcript)
	if err != nil {
		return nil, err
	}

	
	audioURL, err := s.generateAudioResponse(response.Text)
	if err == nil {
		response.AudioURL = audioURL
	}

	response.Transcript = transcript

	
	if err := s.saveVoiceCommand(ctx, repo, userID, audioFilename, transcript, response); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to save voice command")
	}

	
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

	for _, intent := range multiIntent.Intents {
		s.log.WithFields(logrus.Fields{
			"request_id":  requestID,
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

		case "delete_transaction":
			response, err := s.handleDeleteTransactionIntent(ctx, userID, intent, session)
			if err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err.Error(),
				}).Error("Failed to handle delete transaction intent")
				responses = append(responses, "Maaf, gagal menghapus transaksi.")
				continue
			}

			responses = append(responses, response.Text)
			if response.Success {
				successCount++
				if finalAction == "" {
					finalAction = "delete_transaction"
				}
			} else {
				needsConfirmation = response.SessionState != nil && response.SessionState.PendingConfirmation
				if needsConfirmation {
					pendingContext = response.SessionState.Context
				}
			}

		
		case "logout":
			response, err := s.handleLogoutIntent(ctx, intent, session)
			if err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err.Error(),
				}).Error("Failed to handle logout intent")
				responses = append(responses, "Maaf, gagal memproses logout.")
				continue
			}

			responses = append(responses, response.Text)
			needsConfirmation = true 
			pendingContext = response.SessionState.Context
			finalAction = "confirm_logout"

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


func (s *voiceService) handleNavigationIntent(
	ctx context.Context,
	intent chatGPT.Intent,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	
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

	
	if intent.Confidence < 0.7 {
		
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

func (s *voiceService) handleTransactionIntent(
	ctx context.Context,
	userID string,
	intent chatGPT.Intent,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	
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

	
	if amount == 0 || txType == "" {
		return &voice.VoiceResponse{
			Text:    "Maaf, informasi transaksi tidak lengkap. Bisa ulangi dengan menyebutkan nominal dan jenisnya?",
			Action:  "clarify",
			Success: false,
		}, nil
	}

	
	if category == "" || !entity.IsValidCategory(txType, category) {
		category = "lainnya"
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"original_category": intent.Data["category"],
			"fallback_to": "lainnya",
		}).Info("Category not recognized, using fallback 'lainnya'")
	}

	
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

	case "awaiting_logout_confirmation":
		return s.handleLogoutConfirmation(ctx, transcript, session)

	default:
		session.PendingConfirmation = false
		session.Context = make(map[string]interface{})

		return &voice.VoiceResponse{
			Text:    "Maaf, saya lupa apa yang sedang kita bicarakan. Silakan ulangi perintah Anda.",
			Action:  "retry",
			Success: false,
		}, nil
	}
}

func (s *voiceService) handleLogoutConfirmation(
	ctx context.Context,
	transcript string,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	
	isConfirmed := s.isConfirmation(transcript)

	if !isConfirmed {
		
		session.PendingConfirmation = false
		session.Context = make(map[string]interface{})

		return &voice.VoiceResponse{
			Text:    "Baik, logout dibatalkan. Ada yang bisa saya bantu?",
			Action:  "cancelled",
			Success: false,
		}, nil
	}

	
	logoutURL, _ := session.Context["logout_url"].(string)

	
	session.PendingConfirmation = false
	session.Context = make(map[string]interface{})

	return &voice.VoiceResponse{
		Text:       "Baik, Anda akan keluar dari aplikasi.",
		Action:     "logout",
		Target:     logoutURL,
		Success:    true,
		Confidence: 1.0,
		Metadata: map[string]interface{}{
			"logout_url": logoutURL,
			"message":    "Frontend will handle logout locally",
		},
	}, nil
}


func (s *voiceService) handleCategoryConfirmation(
	ctx context.Context,
	userID string,
	transcript string,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	
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
		description = "Transaksi" 
	}

	
	category := s.normalizeCategoryName(transcript)

	
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

	
	txData := &nlp.TransactionData{
		Type:        txType,
		Amount:      amount,
		Description: description,
		Category:    category,
		Confidence:  0.95,
	}

	
	response, err := s.createBudgetTransaction(ctx, userID, txData)
	if err != nil {
		return nil, err
	}

	
	session.PendingConfirmation = false
	session.Context = make(map[string]interface{})

	return response, nil
}


func (s *voiceService) handleNavigationConfirmation(
	ctx context.Context,
	transcript string,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	
	isConfirmed := s.isConfirmation(transcript)

	if !isConfirmed {
		
		session.PendingConfirmation = false
		session.PendingPageID = ""
		session.Context = make(map[string]interface{})

		return &voice.VoiceResponse{
			Text:    "Baik, dibatalkan. Ada yang bisa saya bantu?",
			Action:  "cancelled",
			Success: false,
		}, nil
	}

	
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


func (s *voiceService) getAvailablePagesForGPT(
	ctx context.Context,
	repo voiceRepository.Client,  
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


func (s *voiceService) saveVoiceCommand(
	ctx context.Context,
	repo voiceRepository.Client,  
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
		"sehari-hari":   "sehari-hari",
		"daily":         "sehari-hari",
		"transportasi":  "transportasi",
		"transport":     "transportasi",
		"sosial":        "sosial",
		"social":        "sosial",
		"perumahan":     "perumahan",
		"housing":       "perumahan",
		"hadiah":        "hadiah",
		"gift":          "hadiah",
		"komunikasi":    "komunikasi",
		"communication": "komunikasi",
		"pakaian":       "pakaian",
		"clothing":      "pakaian",
		"hiburan":       "hiburan",
		"entertainment": "hiburan",
		"tampilan":      "tampilan",
		"appearance":    "tampilan",
		"kesehatan":     "kesehatan",
		"health":        "kesehatan",
		"pajak":         "pajak",
		"tax":           "pajak",
		"pendidikan":    "pendidikan",
		"education":     "pendidikan",
		"investasi":     "investasi",
		"investment":    "investasi",
		"peliharaan":    "peliharaan",
		"pet":           "peliharaan",
		"liburan":       "liburan",
		"vacation":      "liburan",
		
		
		"gaji":          "gaji",
		"salary":        "gaji",
		"bonus":         "bonus",
		"part time":     "part time",
		"parttime":      "part time",
		
		
		"lainnya":       "lainnya",
		"other":         "lainnya",
		"lain":          "lainnya",
		"others":        "lainnya",
	}

	normalized := strings.ToLower(strings.TrimSpace(category))
	if mapped, exists := categoryMap[normalized]; exists {
		return mapped
	}

	
	return "lainnya"
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
	
	
	if len(history) > 10 {
		history = history[len(history)-10:]
	}
	
	session.Context["conversation_history"] = history
	session.LastActivity = time.Now()
} 

func (s *voiceService) handleDeleteTransactionIntent(
	ctx context.Context,
	userID string,
	intent chatGPT.Intent,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	amount, _ := intent.Data["amount"].(float64)
	description, _ := intent.Data["description"].(string)

	s.log.WithFields(logrus.Fields{
		"request_id":  requestID,
		"amount":      amount,
		"description": description,
	}).Info("Processing simplified delete transaction intent")

	// Validasi minimal harus ada nominal
	if amount == 0 {
		return &voice.VoiceResponse{
			Text:    "Maaf, mohon sebutkan nominal transaksi yang ingin dihapus.",
			Action:  "clarify",
			Success: false,
		}, nil
	}

	// Ambil semua transaksi user
	allTransactions, err := s.budgetService.GetTransactionsByUserID(ctx, userID)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get user transactions")
		
		return &voice.VoiceResponse{
			Text:    "Maaf, terjadi kesalahan saat mengambil data transaksi.",
			Action:  "error",
			Success: false,
		}, err
	}

	// Filter transaksi berdasarkan nominal dan deskripsi (jika ada)
	matchedTransactions := s.filterTransactionsByNominalAndDesc(allTransactions, amount, description)

	if len(matchedTransactions) == 0 {
		responseText := fmt.Sprintf("Tidak ada transaksi dengan nominal Rp%.0f", amount)
		if description != "" {
			responseText += fmt.Sprintf(" dan deskripsi yang mengandung '%s'", description)
		}
		responseText += " yang ditemukan."

		return &voice.VoiceResponse{
			Text:    responseText,
			Action:  "not_found",
			Success: false,
		}, nil
	}

	// LANGSUNG HAPUS SEMUA transaksi yang cocok
	deletedCount := 0
	failedCount := 0
	var failedIDs []string
	
	for _, tx := range matchedTransactions {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"tx_id":      tx.ID,
			"nominal":    tx.Nominal,
			"desc":       tx.Description,
		}).Info("Deleting matched transaction")

		if err := s.budgetService.DeleteTransaction(ctx, tx.ID, userID); err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"tx_id":      tx.ID,
				"error":      err.Error(),
			}).Error("Failed to delete transaction")
			failedCount++
			failedIDs = append(failedIDs, tx.ID)
		} else {
			deletedCount++
		}
	}

	// Buat response berdasarkan hasil
	var responseText string
	if deletedCount > 0 {
		if deletedCount == 1 {
			responseText = fmt.Sprintf("Berhasil menghapus 1 transaksi dengan nominal Rp%.0f", amount)
		} else {
			responseText = fmt.Sprintf("Berhasil menghapus %d transaksi dengan nominal Rp%.0f", deletedCount, amount)
		}
		
		if description != "" {
			responseText += fmt.Sprintf(" yang mengandung '%s'", description)
		}
		responseText += "."
		
		if failedCount > 0 {
			responseText += fmt.Sprintf(" Namun, %d transaksi gagal dihapus.", failedCount)
		}
	} else {
		responseText = "Maaf, tidak ada transaksi yang berhasil dihapus. Silakan coba lagi."
	}

	return &voice.VoiceResponse{
		Text:    responseText,
		Action:  "delete_complete",
		Success: deletedCount > 0,
		Metadata: map[string]interface{}{
			"deleted_count": deletedCount,
			"failed_count":  failedCount,
			"total_found":   len(matchedTransactions),
			"failed_ids":    failedIDs,
		},
	}, nil
}

func (s *voiceService) handleLogoutIntent(
	ctx context.Context,
	intent chatGPT.Intent,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	s.log.WithFields(logrus.Fields{
		"request_id": requestID,
		"action":     intent.Action,
		"confidence": intent.Confidence,
	}).Info("Processing logout intent")

	
	url, _ := intent.Data["url"].(string)
	displayName, _ := intent.Data["display_name"].(string)

	
	session.PendingConfirmation = true
	session.Context = map[string]interface{}{
		"step":            "awaiting_logout_confirmation",
		"logout_url":      url,
		"logout_action":   true,
	}

	responseText := "Apakah Anda yakin ingin keluar dari aplikasi?"

	return &voice.VoiceResponse{
		Text:       responseText,
		Action:     "confirm_logout",
		Success:    false,
		Confidence: intent.Confidence,
		Metadata: map[string]interface{}{
			"logout_url":      url,
			"display_name":    displayName,
		},
		SessionState: &voice.SessionState{
			PendingConfirmation: true,
			Context:             session.Context,
		},
	}, nil
}

func (s *voiceService) filterTransactionsByNominalAndDesc(
	transactions []entity.BudgetTransaction,
	amount float64,
	description string,
) []entity.BudgetTransaction {
	var matched []entity.BudgetTransaction

	for _, tx := range transactions {
		// Filter berdasarkan nominal (harus sama persis)
		if tx.Nominal != amount {
			continue
		}

		// Jika deskripsi disebutkan, filter juga berdasarkan deskripsi
		// Menggunakan case-insensitive substring matching
		if description != "" {
			txDescLower := strings.ToLower(tx.Description)
			searchDescLower := strings.ToLower(strings.TrimSpace(description))
			
			if !strings.Contains(txDescLower, searchDescLower) {
				continue
			}
		}

		// Transaksi cocok dengan kriteria
		matched = append(matched, tx)
	}

	return matched
}