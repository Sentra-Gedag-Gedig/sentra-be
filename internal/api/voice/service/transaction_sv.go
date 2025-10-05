package voiceService

import (
	"ProjectGolang/internal/api/budget_manager"
	"ProjectGolang/internal/api/voice"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/nlp"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func (s *voiceService) ProcessTransactionCommand(
	ctx context.Context,
	userID string,
	transcript string,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	extractor := nlp.NewNumberExtractor()
	txData, err := extractor.ExtractTransaction(transcript)
	
	if err != nil || txData == nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"transcript": transcript,
		}).Warn("Failed to extract transaction data")
		
		return &voice.VoiceResponse{
			Text:    "Maaf, saya tidak dapat memahami detail transaksi. Mohon sebutkan nominal dan keterangan dengan jelas.",
			Action:  "clarify",
			Success: false,
		}, nil
	}

	// Check if category is missing
	if txData.Category == "" {
		// Store transaction data in session context
		session.Context = map[string]interface{}{
			"step":              "awaiting_category",
			"transaction_type":  txData.Type,
			"transaction_amount": txData.Amount,
			"transaction_desc":  txData.Description,
		}
		session.PendingConfirmation = true
		session.LastActivity = time.Now()

		// Get available categories
		categories := s.getAvailableCategories(txData.Type)
		categoriesText := strings.Join(categories, ", ")

		responseText := fmt.Sprintf(
			"Kategori apa? Pilihan %s: %s",
			txData.Type,
			categoriesText,
		)

		return &voice.VoiceResponse{
			Text:       responseText,
			Action:     "ask_category",
			Success:    false,
			Confidence: txData.Confidence,
			SessionState: &voice.SessionState{
				PendingConfirmation: true,
				Context:             session.Context,
			},
		}, nil
	}

	// Create transaction
	return s.createBudgetTransaction(ctx, userID, txData)
}

func (s *voiceService) ProcessCategorySelection(
	ctx context.Context,
	userID string,
	categoryInput string,
	session *entity.VoiceSession,
) (*voice.VoiceResponse, error) {
	//requestID := contextPkg.GetRequestID(ctx)

	// Validate session context
	if session.Context["step"] != "awaiting_category" {
		return &voice.VoiceResponse{
			Text:    "Maaf, tidak ada transaksi yang sedang menunggu kategori.",
			Action:  "error",
			Success: false,
		}, nil
	}

	// Extract transaction data from session
	txType := session.Context["transaction_type"].(string)
	amount := session.Context["transaction_amount"].(float64)
	description := session.Context["transaction_desc"].(string)

	// Clean and validate category
	category := strings.ToLower(strings.TrimSpace(categoryInput))
	
	// Validate category
	if !entity.IsValidCategory(txType, category) {
		categories := s.getAvailableCategories(txType)
		categoriesText := strings.Join(categories, ", ")

		responseText := fmt.Sprintf(
			"Kategori '%s' tidak tersedia. Pilihan yang benar adalah: %s. Silakan ulangi dengan salah satu kategori tersebut.",
			categoryInput,
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
		Confidence:  0.9,
	}

	// Clear session context
	session.PendingConfirmation = false
	session.Context = make(map[string]interface{})

	// Create transaction
	return s.createBudgetTransaction(ctx, userID, txData)
}

func (s *voiceService) createBudgetTransaction(
	ctx context.Context,
	userID string,
	txData *nlp.TransactionData,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	// Create transaction request
	req := budget_manager.CreateTransactionRequest{
		UserID:      userID,
		Title:       txData.Description,
		Description: txData.Description,
		Nominal:     txData.Amount,
		Type:        txData.Type,
		Category:    txData.Category,
	}

	// Call budget service
	if err := s.budgetService.CreateTransaction(ctx, req, nil); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create budget transaction")

		return &voice.VoiceResponse{
			Text:    "Maaf, terjadi kesalahan saat menyimpan transaksi.",
			Action:  "error",
			Success: false,
		}, err
	}

	// Format response
	typeText := "Pemasukan"
	if txData.Type == "expense" {
		typeText = "Pengeluaran"
	}

	responseText := fmt.Sprintf(
		"Dicatat: %s Rp%.0f untuk %s, kategori %s.",
		typeText,
		txData.Amount,
		txData.Description,
		txData.Category,
	)

	return &voice.VoiceResponse{
		Text:       responseText,
		Action:     "transaction_created",
		Success:    true,
		Confidence: txData.Confidence,
		Metadata: map[string]interface{}{
			"transaction_type":   txData.Type,
			"transaction_amount": txData.Amount,
			"transaction_category": txData.Category,
		},
	}, nil
}

func (s *voiceService) ProcessDeleteTransaction(
	ctx context.Context,
	userID string,
	transcript string,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	extractor := nlp.NewNumberExtractor()
	
	// Extract amount and description
	amount, _ := extractor.ExtractAmount(transcript)
	description := extractor.ExtractDescription(transcript, amount)

	if amount == 0 {
		return &voice.VoiceResponse{
			Text:    "Mohon sebutkan nominal transaksi yang ingin dihapus.",
			Action:  "clarify",
			Success: false,
		}, nil
	}

	// Search for matching transaction
	transactions, err := s.budgetService.GetTransactionsByUserID(ctx, userID)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get transactions")
		
		return &voice.VoiceResponse{
			Text:    "Maaf, terjadi kesalahan saat mencari transaksi.",
			Action:  "error",
			Success: false,
		}, err
	}

	// Find matching transaction
	var matchedTx *entity.BudgetTransaction
	for _, tx := range transactions {
		if tx.Nominal == amount {
			if description == "" || strings.Contains(strings.ToLower(tx.Description), strings.ToLower(description)) {
				matchedTx = &tx
				break
			}
		}
	}

	if matchedTx == nil {
		return &voice.VoiceResponse{
			Text:    fmt.Sprintf("Transaksi Rp%.0f %s tidak ditemukan. Silakan coba dengan nominal dan keterangan lain.", amount, description),
			Action:  "not_found",
			Success: false,
		}, nil
	}

	// Ask for confirmation
	responseText := fmt.Sprintf(
		"Apakah Anda yakin ingin hapus transaksi Rp%.0f %s?",
		matchedTx.Nominal,
		matchedTx.Description,
	)

	return &voice.VoiceResponse{
		Text:    responseText,
		Action:  "confirm_delete",
		Success: false,
		Metadata: map[string]interface{}{
			"transaction_id":          matchedTx.ID,
			"pending_delete_confirmation": true,
		},
	}, nil
}

func (s *voiceService) ProcessMonthlySummary(
	ctx context.Context,
	userID string,
) (*voice.VoiceResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	// Get current month transactions
	transactions, err := s.budgetService.GetTransactionsByPeriod(ctx, userID, "month")
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get monthly transactions")
		
		return &voice.VoiceResponse{
			Text:    "Maaf, terjadi kesalahan saat mengambil data transaksi.",
			Action:  "error",
			Success: false,
		}, err
	}

	var totalIncome, totalExpense float64
	for _, tx := range transactions {
		if tx.Type == "income" {
			totalIncome += tx.Nominal
		} else {
			totalExpense += tx.Nominal
		}
	}

	balance := totalIncome - totalExpense

	responseText := fmt.Sprintf(
		"Saldo saat ini Rp%.0f. Total pemasukan Rp%.0f. Total pengeluaran Rp%.0f.",
		balance,
		totalIncome,
		totalExpense,
	)

	return &voice.VoiceResponse{
		Text:    responseText,
		Action:  "summary",
		Success: true,
		Metadata: map[string]interface{}{
			"balance":       balance,
			"total_income":  totalIncome,
			"total_expense": totalExpense,
		},
	}, nil
}

func (s *voiceService) getAvailableCategories(transactionType string) []string {
	if transactionType == "income" {
		return []string{"Gaji", "Bonus", "Investasi", "Bisnis", "Freelance", "Lainnya"}
	}
	return []string{"Makanan", "Transportasi", "Belanja", "Kesehatan", "Hiburan", "Tagihan", "Lainnya"}
}

func (s *voiceService) detectCommandType(transcript string) string {
	transcript = strings.ToLower(transcript)

	// Transaction keywords - MORE FLEXIBLE
	transactionKeywords := []string{
		"tambah", "catat", "buat transaksi", "pencatatan",
		"beli", "bayar", "belanja", "buat", // ADD THESE
		"dapat", "terima", "pendapatan", "masuk", // ADD THESE
	}
	
	transactionDeleteKeywords := []string{"hapus", "delete", "buang"}
	summaryKeywords := []string{"ringkas", "summary", "total", "saldo"}
	navigationKeywords := []string{"buka", "pergi", "lihat", "tampilkan"}

	// Check for transaction
	for _, keyword := range transactionKeywords {
		if strings.Contains(transcript, keyword) {
			// Check if it has amount indicators
			hasAmount := strings.Contains(transcript, "ribu") || 
				strings.Contains(transcript, "juta") || 
				regexp.MustCompile(`\d+`).MatchString(transcript)
			
			if hasAmount {
				return "transaction_create"
			}
		}
	}

	for _, keyword := range transactionDeleteKeywords {
		if strings.Contains(transcript, keyword) && strings.Contains(transcript, "transaksi") {
			return "transaction_delete"
		}
	}

	for _, keyword := range summaryKeywords {
		if strings.Contains(transcript, keyword) {
			return "monthly_summary"
		}
	}

	for _, keyword := range navigationKeywords {
		if strings.Contains(transcript, keyword) {
			return "navigation"
		}
	}

	return "unknown"
}