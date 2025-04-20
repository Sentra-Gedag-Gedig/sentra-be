package sentrapayService

import (
	sentrapay "ProjectGolang/internal/api/sentra_pay"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/doku"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"os"
	"strconv"
	"strings"
	"time"
)

func (s *sentraPayService) CreateTopUpTransaction(ctx context.Context, userID string, req sentrapay.TopUpRequest) (*sentrapay.TopUpResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	if !isValidBank(req.Bank) {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"bank":       req.Bank,
		}).Warn("Invalid bank selection")
		return nil, sentrapay.ErrInvalidBank
	}

	if req.Amount <= 0 {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"amount":     req.Amount,
		}).Warn("Invalid amount")
		return nil, sentrapay.ErrInvalidAmount
	}

	repo, err := s.walletRepository.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create new client")
		return nil, err
	}
	defer repo.Rollback()

	_, err = repo.Wallet.GetWallet(ctx, userID)
	if err != nil {
		if errors.Is(err, sentrapay.ErrWalletNotFound) {
			if err = repo.Wallet.CreateWallet(ctx, userID); err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"user_id":    userID,
					"error":      err.Error(),
				}).Error("Failed to create wallet")
				return nil, err
			}
		} else {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"user_id":    userID,
				"error":      err.Error(),
			}).Error("Failed to get wallet")
			return nil, err
		}
	}

	transactionID, err := s.utils.NewULIDFromTimestamp(time.Now())
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to generate ULID")
		return nil, err
	}

	refNo := fmt.Sprintf("TOP%s%d", userID, time.Now().Unix())

	authRepo, err := s.authRepo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create auth repository client")
		return nil, err
	}
	user, err := authRepo.Users.GetByID(ctx, userID)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get user info")
		return nil, err
	}

	dokuReq := doku.CreateVaRequest{
		UserID:          userID,
		Name:            user.Name,
		Email:           user.Email,
		Phone:           user.PhoneNumber,
		Amount:          req.Amount,
		TrxId:           refNo,
		Bank:            req.Bank,
		ExpiredDuration: 24 * time.Hour,
		ReusableStatus:  false,
	}

	dokuRes, err := s.dokuService.CreateVirtualAccount(dokuReq)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create virtual account")
		return nil, sentrapay.ErrCreateVirtualAccount
	}

	transaction := sentrapay.WalletTransaction{
		ID:            transactionID,
		UserID:        userID,
		Amount:        req.Amount,
		Type:          "topup",
		ReferenceNo:   refNo,
		PaymentMethod: "virtual_account",
		Status:        "pending",
		BankAccount:   dokuRes.VirtualAccountNo,
		BankName:      getBankName(req.Bank),
		Description:   fmt.Sprintf("Top up via %s", getBankName(req.Bank)),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := repo.Wallet.CreateTransaction(ctx, transaction); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create transaction")
		return nil, sentrapay.ErrCreateTransaction
	}

	if err := repo.Commit(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction")
		return nil, err
	}

	response := &sentrapay.TopUpResponse{
		TransactionID:   transactionID,
		ReferenceNo:     refNo,
		VirtualAccount:  dokuRes.VirtualAccountNo,
		Bank:            getBankName(req.Bank),
		Amount:          req.Amount,
		ExpiresAt:       dokuRes.ExpiryDate,
		PaymentGuideURL: dokuRes.VirtualAccountURL,
		Status:          "pending",
		CreatedAt:       transaction.CreatedAt,
	}

	return response, nil
}

func (s *sentraPayService) ProcessPaymentCallback(ctx context.Context, req sentrapay.PaymentCallbackRequest, channelID, xExternalID, xTimestamp, xPartnerID string) error {
	requestID := contextPkg.GetRequestID(ctx)

	req.PartnerServiceId = strings.TrimSpace(req.PartnerServiceId)
	req.VirtualAccountNo = strings.TrimSpace(req.VirtualAccountNo)
	req.CustomerNo = strings.TrimSpace(req.CustomerNo)

	if req.TrxId == "" || req.VirtualAccountNo == "" {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"reference_no": req.TrxId,
		}).Error("Missing required fields in payment callback")
		return sentrapay.ErrInvalidCallback
	}

	expectedPartnerID := os.Getenv("DOKU_CLIENT_ID")
	if expectedPartnerID != "" && xPartnerID != expectedPartnerID {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"expected":   expectedPartnerID,
			"received":   xPartnerID,
		}).Error("Invalid partner ID in payment callback")
		return sentrapay.ErrInvalidCallback
	}

	if xTimestamp != "" {
		parsedTime, err := time.Parse(time.RFC3339, xTimestamp)
		if err == nil {
			timeWindow := 5 * time.Minute
			if time.Since(parsedTime) > timeWindow {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"timestamp":  xTimestamp,
					"window":     timeWindow,
				}).Warn("Payment callback timestamp is older than expected window")
			}
		}
	}

	s.log.WithFields(logrus.Fields{
		"request_id":       requestID,
		"reference_no":     req.TrxId,
		"virtual_acc_no":   req.VirtualAccountNo,
		"virtual_acc_name": req.VirtualAccountName,
		"amount":           req.PaidAmount.Value,
		"payment_channel":  req.AdditionalInfo.Channel,
		"transaction_time": req.TrxDateTime,
	}).Info("Processing payment callback")

	repo, err := s.walletRepository.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create database client")
		return err
	}
	defer repo.Rollback()

	transaction, err := repo.Wallet.GetTransactionByReferenceNo(ctx, req.TrxId)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"reference_no": req.TrxId,
			"error":        err.Error(),
		}).Error("Failed to get transaction")
		return err
	}

	if transaction.Status == "success" {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"reference_no": req.TrxId,
		}).Info("Transaction already processed successfully")
		return nil
	}

	if transaction.Status != "pending" && transaction.Status != "processing" {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"reference_no": req.TrxId,
			"status":       transaction.Status,
		}).Warn("Transaction in invalid state for payment callback")
		return sentrapay.ErrInvalidTransactionState
	}

	paidAmount, err := strconv.ParseFloat(strings.TrimSpace(req.PaidAmount.Value), 64)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"amount":     req.PaidAmount.Value,
			"error":      err.Error(),
		}).Error("Failed to parse amount")
		return err
	}

	const amountTolerance = 0.01
	if paidAmount < transaction.Amount-amountTolerance || paidAmount > transaction.Amount+amountTolerance {
		s.log.WithFields(logrus.Fields{
			"request_id":      requestID,
			"expected_amount": transaction.Amount,
			"received_amount": paidAmount,
			"difference":      paidAmount - transaction.Amount,
		}).Warn("Amount mismatch in payment callback")

		return sentrapay.ErrInvalidAmount
	}

	if err := repo.Wallet.UpdateTransactionStatus(ctx, req.TrxId, "success"); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"reference_no": req.TrxId,
			"error":        err.Error(),
		}).Error("Failed to update transaction status")
		return err
	}

	wallet, err := repo.Wallet.GetWallet(ctx, transaction.UserID)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    transaction.UserID,
			"error":      err.Error(),
		}).Error("Failed to get wallet")
		return err
	}

	newBalance := wallet.Balance + paidAmount
	if err := repo.Wallet.UpdateWalletBalance(ctx, transaction.UserID, newBalance); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id":  requestID,
			"user_id":     transaction.UserID,
			"old_balance": wallet.Balance,
			"new_balance": newBalance,
			"error":       err.Error(),
		}).Error("Failed to update wallet balance")
		return err
	}

	if err := repo.Commit(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction")
		return err
	}

	s.log.WithFields(logrus.Fields{
		"request_id":   requestID,
		"reference_no": req.TrxId,
		"user_id":      transaction.UserID,
		"old_balance":  wallet.Balance,
		"new_balance":  newBalance,
		"amount":       paidAmount,
	}).Info("Payment processed successfully")

	return nil
}

func (s *sentraPayService) GetWalletBalance(ctx context.Context, userID string) (*sentrapay.WalletBalance, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.walletRepository.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create new client")
		return nil, err
	}

	wallet, err := repo.Wallet.GetWallet(ctx, userID)
	if err != nil {
		if errors.Is(err, sentrapay.ErrWalletNotFound) {
			if err = repo.Wallet.CreateWallet(ctx, userID); err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"user_id":    userID,
					"error":      err.Error(),
				}).Error("Failed to create wallet")
				return nil, err
			}
			return &sentrapay.WalletBalance{
				UserID:      userID,
				Balance:     0,
				LastUpdated: time.Now(),
			}, nil
		}

		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get wallet")

		return nil, err
	}

	return &wallet, nil
}

func (s *sentraPayService) GetTransactionHistory(ctx context.Context, userID string, page, limit int) (*sentrapay.TransactionHistoryResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.walletRepository.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create new client")
		return nil, err
	}

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	transactions, total, err := repo.Wallet.GetTransactionsByUserID(ctx, userID, limit, offset)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get transactions")
		return nil, err
	}

	return &sentrapay.TransactionHistoryResponse{
		Transactions: transactions,
		Total:        total,
	}, nil
}

func (s *sentraPayService) CheckTransactionStatus(ctx context.Context, referenceNo string) (string, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.walletRepository.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create new client")
		return "", err
	}

	transaction, err := repo.Wallet.GetTransactionByReferenceNo(ctx, referenceNo)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"reference_no": referenceNo,
			"error":        err.Error(),
		}).Error("Failed to get transaction")
		return "", err
	}

	if transaction.Status == "success" {
		return "success", nil
	}

	partnerServiceId := "  820901"
	customerNo := fmt.Sprintf("%020s", transaction.UserID)

	isPaid, err := s.dokuService.CheckVAStatus(transaction.BankAccount, customerNo, partnerServiceId, referenceNo)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"reference_no": referenceNo,
			"error":        err.Error(),
		}).Error("Failed to check VA status")
		return transaction.Status, nil
	}

	if isPaid && transaction.Status == "pending" {
		repoTx, err := s.walletRepository.NewClient(true)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to create new client")
			return "", err
		}
		defer repoTx.Rollback()

		if err := repoTx.Wallet.UpdateTransactionStatus(ctx, referenceNo, "success"); err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id":   requestID,
				"reference_no": referenceNo,
				"error":        err.Error(),
			}).Error("Failed to update transaction status")
			return transaction.Status, nil
		}

		wallet, err := repoTx.Wallet.GetWallet(ctx, transaction.UserID)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"user_id":    transaction.UserID,
				"error":      err.Error(),
			}).Error("Failed to get wallet")
			return transaction.Status, nil
		}

		newBalance := wallet.Balance + transaction.Amount
		if err := repoTx.Wallet.UpdateWalletBalance(ctx, transaction.UserID, newBalance); err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"user_id":    transaction.UserID,
				"error":      err.Error(),
			}).Error("Failed to update wallet balance")
			return transaction.Status, nil
		}

		if err := repoTx.Commit(); err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to commit transaction")
			return transaction.Status, nil
		}

		return "success", nil
	}

	return transaction.Status, nil
}

func isValidBank(bank string) bool {
	validBanks := map[string]bool{
		doku.BankBCA:      true,
		doku.BankMANDIRI:  true,
		doku.BankBRI:      true,
		doku.BankBNI:      true,
		doku.BankDANAMON:  true,
		doku.BankPERMATA:  true,
		doku.BankMAYBANK:  true,
		doku.BankBTN:      true,
		doku.BankBSI:      true,
		doku.BankCIMB:     true,
		doku.BankSINARMAS: true,
		doku.BankDOKU:     true,
	}

	return validBanks[bank]
}

func getBankName(bank string) string {
	bankNames := map[string]string{
		doku.BankBCA:      "BCA",
		doku.BankMANDIRI:  "MANDIRI",
		doku.BankBRI:      "BRI",
		doku.BankBNI:      "BNI",
		doku.BankDANAMON:  "DANAMON",
		doku.BankPERMATA:  "PERMATA",
		doku.BankMAYBANK:  "MAYBANK",
		doku.BankBTN:      "BTN",
		doku.BankBSI:      "BSI",
		doku.BankCIMB:     "CIMB",
		doku.BankSINARMAS: "SINARMAS",
		doku.BankDOKU:     "DOKU",
	}

	name, exists := bankNames[bank]
	if !exists {
		return "UNKNOWN"
	}

	return name
}
