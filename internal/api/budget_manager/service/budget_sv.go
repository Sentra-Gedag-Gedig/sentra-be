package budgetService

import (
	"ProjectGolang/internal/api/budget_manager"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

func (s *budgetService) CreateTransaction(ctx context.Context, req budget_manager.CreateTransactionRequest, audioFile *multipart.FileHeader) error {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.budgetRepository.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create new client")
		return err
	}
	var audioLink string

	if !entity.IsValidCategory(req.Type, req.Category) {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"type":       req.Type,
			"category":   req.Category,
		}).Warn("Invalid transaction category for type")
		return budget_manager.ErrInvalidCategory
	}

	var fileName string
	if audioFile != nil {
		if !isAudioFile(audioFile.Filename) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"filename":   audioFile.Filename,
			}).Warn("Invalid audio file type")
			return errors.New("invalid audio file type")
		}

		uploadedFileURL, err := s.s3.UploadFile(audioFile)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to upload audio file")
			return err
		}
		audioLink = uploadedFileURL
	}

	ULID, err := s.utils.NewULIDFromTimestamp(time.Now())
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to generate ULID")
		return err
	}

	transaction := entity.BudgetTransaction{
		ID:          ULID,
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		Nominal:     req.Nominal,
		Type:        req.Type,
		Category:    req.Category,
		AudioLink:   audioLink,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := transaction.Validate(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Invalid transaction data")
		return err
	}

	if err := repo.Budget.CreateTransaction(ctx, transaction); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create transaction")

		if audioLink != "" {
			if deleteErr := s.s3.DeleteFile(fileName); deleteErr != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      deleteErr.Error(),
				}).Error("Failed to delete audio file after transaction creation failure")
			}
		}

		return budget_manager.ErrCreateTransaction
	}

	return nil
}

func (s *budgetService) GetTransactionByID(ctx context.Context, id string) (entity.BudgetTransaction, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.budgetRepository.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create new client")
		return entity.BudgetTransaction{}, err
	}

	transaction, err := repo.Budget.GetTransactionByID(ctx, id)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         id,
			"error":      err.Error(),
		}).Error("Failed to get transaction by ID")
		return entity.BudgetTransaction{}, err
	}

	audiolink, err := s.s3.PresignUrl(transaction.AudioLink)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to presign audio link")
		return entity.BudgetTransaction{}, err
	}

	transaction.AudioLink = audiolink
	return transaction, nil
}

func (s *budgetService) GetTransactionsByUserID(ctx context.Context, userID string) ([]entity.BudgetTransaction, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.budgetRepository.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create new client")
		return nil, err
	}

	transactions, err := repo.Budget.GetTransactionsByUserID(ctx, userID)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get transactions by user ID")
		return nil, err
	}

	for i, transaction := range transactions {
		if transaction.AudioLink != "" {
			audiolink, err := s.s3.PresignUrl(transaction.AudioLink)
			if err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err.Error(),
				}).Error("Failed to presign audio link")
				return nil, err
			}
			transactions[i].AudioLink = audiolink
			fmt.Println("Audio link presigned:", transactions[i].AudioLink)
		}
	}

	return transactions, nil
}

func (s *budgetService) GetTransactionsByPeriod(ctx context.Context, userID string, period string) ([]entity.BudgetTransaction, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.budgetRepository.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create new client")
		return nil, err
	}
	if period != "all" && period != "week" && period != "month" {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"period":     period,
		}).Warn("Invalid period")
		period = "all"
	}

	transactions, err := repo.Budget.GetTransactionsByPeriod(ctx, userID, period)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"period":     period,
			"error":      err.Error(),
		}).Error("Failed to get transactions by period")
		return nil, err
	}

	return transactions, nil
}

func (s *budgetService) UpdateTransaction(ctx context.Context, req budget_manager.UpdateTransactionRequest, audioFile *multipart.FileHeader) error {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.budgetRepository.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create new client")
		return err
	}

	if !entity.IsValidCategory(req.Type, req.Category) {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"type":       req.Type,
			"category":   req.Category,
		}).Warn("Invalid transaction category for type")
		return budget_manager.ErrInvalidCategory
	}

	existingTransaction, err := repo.Budget.GetTransactionByID(ctx, req.ID)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         req.ID,
			"error":      err.Error(),
		}).Error("Failed to get existing transaction")
		return err
	}

	if existingTransaction.UserID != req.UserID {
		s.log.WithFields(logrus.Fields{
			"request_id":          requestID,
			"transaction_user_id": existingTransaction.UserID,
			"request_user_id":     req.UserID,
		}).Warn("Transaction does not belong to user")
		return errors.New("transaction does not belong to user")
	}

	audioLink := existingTransaction.AudioLink

	if req.DeleteAudio && audioLink != "" {
		parts := strings.Split(audioLink, "/")
		fileName := parts[len(parts)-1]

		if err := s.s3.DeleteFile(fileName); err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"fileName":   fileName,
				"error":      err.Error(),
			}).Error("Failed to delete audio file")
		}

		audioLink = ""
	}

	if audioFile != nil {
		if !isAudioFile(audioFile.Filename) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"filename":   audioFile.Filename,
			}).Warn("Invalid audio file type")
			return errors.New("invalid audio file type")
		}

		if audioLink != "" {
			parts := strings.Split(audioLink, "/")
			fileName := parts[len(parts)-1]

			if err := s.s3.DeleteFile(fileName); err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"fileName":   fileName,
					"error":      err.Error(),
				}).Error("Failed to delete existing audio file")
			}
		}

		uploadedFileURL, err := s.s3.UploadFile(audioFile)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to upload audio file")
			return err
		}
		audioLink = uploadedFileURL
	}

	transaction := entity.BudgetTransaction{
		ID:          req.ID,
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		Nominal:     req.Nominal,
		Type:        req.Type,
		Category:    req.Category,
		AudioLink:   audioLink,
		UpdatedAt:   time.Now(),
	}

	if err := transaction.Validate(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Invalid transaction data")
		return err
	}

	if err := repo.Budget.UpdateTransaction(ctx, transaction); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to update transaction")
		return budget_manager.ErrUpdateTransaction
	}

	return nil
}

func (s *budgetService) DeleteTransaction(ctx context.Context, id string, userID string) error {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.budgetRepository.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create new client")
		return err
	}

	existingTransaction, err := repo.Budget.GetTransactionByID(ctx, id)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         id,
			"error":      err.Error(),
		}).Error("Failed to get existing transaction")
		return err
	}

	if existingTransaction.UserID != userID {
		s.log.WithFields(logrus.Fields{
			"request_id":          requestID,
			"transaction_user_id": existingTransaction.UserID,
			"request_user_id":     userID,
		}).Warn("Transaction does not belong to user")
		return errors.New("transaction does not belong to user")
	}

	if existingTransaction.AudioLink != "" {
		parts := strings.Split(existingTransaction.AudioLink, "/")
		fileName := parts[len(parts)-1]

		fmt.Println("Deleting file:", fileName)
		if err := s.s3.DeleteFile(fileName); err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"fileName":   fileName,
				"error":      err.Error(),
			}).Error("Failed to delete audio file")
		}
	}

	if err := repo.Budget.DeleteTransaction(ctx, id); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to delete transaction")
		return budget_manager.ErrDeleteTransaction
	}

	return nil
}

func (s *budgetService) GetTransactionsByTypeAndCategory(ctx context.Context, userID string, transactionType string, category string) ([]entity.BudgetTransaction, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.budgetRepository.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create new client")
		return nil, err
	}

	if transactionType != string(entity.TransactionTypeIncome) && transactionType != string(entity.TransactionTypeExpense) {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"type":       transactionType,
		}).Warn("Invalid transaction type")
		return nil, budget_manager.ErrInvalidTransactionType
	}

	if !entity.IsValidCategory(transactionType, category) {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"type":       transactionType,
			"category":   category,
		}).Warn("Invalid transaction category for type")
		return nil, budget_manager.ErrInvalidCategory
	}

	transactions, err := repo.Budget.GetTransactionsByTypeAndCategory(ctx, userID, transactionType, category)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"type":       transactionType,
			"category":   category,
			"error":      err.Error(),
		}).Error("Failed to get transactions by type and category")
		return nil, err
	}

	return transactions, nil
}

func isAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".mp3":  true,
		".wav":  true,
		".ogg":  true,
		".m4a":  true,
		".flac": true,
	}

	return validExtensions[ext]
}
