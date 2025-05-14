package sentrapayService

import (
	sentrapay "ProjectGolang/internal/api/sentra_pay"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/log"
	"context"
	"fmt"
	"time"
)

func (s *sentraPayService) DecodeQRIS(ctx context.Context, req sentrapay.QRISDecodeRequest) (*sentrapay.QRISDecodeResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	decodeResponse, err := s.dokuService.DecodeQRIS(req.QRContent)
	if err != nil {
		s.log.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to decode QRIS")
		return nil, err
	}

	var amount, feeAmount float64

	if v, err := decodeResponse.TransactionAmount.Value.Float64(); err == nil {
		amount = v
	} else {
		s.log.WithFields(log.Fields{
			"request_id": requestID,
			"value":      decodeResponse.TransactionAmount.Value.String(),
			"error":      err.Error(),
		}).Error("Failed to parse transaction amount")
		return nil, fmt.Errorf("failed to parse transaction amount: %v", err)
	}

	if v, err := decodeResponse.FeeAmount.Value.Float64(); err == nil {
		feeAmount = v
	} else {
		s.log.WithFields(log.Fields{
			"request_id": requestID,
			"value":      decodeResponse.FeeAmount.Value.String(),
			"error":      err.Error(),
		}).Error("Failed to parse fee amount")
		return nil, fmt.Errorf("failed to parse fee amount: %v", err)
	}

	response := &sentrapay.QRISDecodeResponse{
		ReferenceNo:  decodeResponse.ReferenceNo,
		MerchantName: decodeResponse.MerchantName,
		Amount:       amount,
		FeeAmount:    feeAmount,
		TotalAmount:  amount + feeAmount,
		PaymentType:  getPaymentType(decodeResponse.AdditionalInfo.PointOfInitiationMethod),
		AdditionalInfo: sentrapay.QRISDecodeAdditionalInfo{
			PointOfInitiationMethod:            decodeResponse.AdditionalInfo.PointOfInitiationMethod,
			PointOfInitiationMethodDescription: decodeResponse.AdditionalInfo.PointOfInitiationMethodDescription,
			FeeType:                            decodeResponse.AdditionalInfo.FeeType,
			FeeTypeDescription:                 decodeResponse.AdditionalInfo.FeeTypeDescription,
		},
	}

	return response, nil
}

func (s *sentraPayService) PaymentQRIS(ctx context.Context, userID string, req sentrapay.QRISPaymentRequest) (*sentrapay.QRISPaymentResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	decodeRequest := sentrapay.QRISDecodeRequest{
		QRContent: req.QRContent,
	}
	decodeResponse, err := s.DecodeQRIS(ctx, decodeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode QRIS: %v", err)
	}

	repo, err := s.walletRepository.NewClient(true)
	if err != nil {
		s.log.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return nil, err
	}
	defer repo.Rollback()

	wallet, err := repo.Wallet.GetWallet(ctx, userID)
	if err != nil {
		s.log.WithFields(log.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to get wallet")
		return nil, err
	}

	if wallet.Balance < decodeResponse.TotalAmount {
		s.log.WithFields(log.Fields{
			"request_id":   requestID,
			"user_id":      userID,
			"balance":      wallet.Balance,
			"total_amount": decodeResponse.TotalAmount,
		}).Warn("Insufficient balance for QRIS payment")
		return nil, sentrapay.ErrInsufficientBalance
	}

	paymentResponse, err := s.dokuService.PaymentQRIS(
		req.QRContent,
		decodeResponse.Amount,
		decodeResponse.FeeAmount,
		req.AuthCode,
	)
	if err != nil {
		s.log.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to process QRIS payment")
		return nil, err
	}

	transactionID, err := s.utils.NewULIDFromTimestamp(time.Now())
	if err != nil {
		s.log.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to generate transaction ID")
		return nil, err
	}

	totalAmount := decodeResponse.Amount + decodeResponse.FeeAmount

	transaction := sentrapay.WalletTransaction{
		ID:            transactionID,
		UserID:        userID,
		Amount:        totalAmount * -1,
		Type:          "qris_payment",
		ReferenceNo:   paymentResponse.ReferenceNo,
		PaymentMethod: "qris",
		Status:        "success",
		BankAccount:   decodeResponse.MerchantName,
		BankName:      "QRIS",
		Description:   fmt.Sprintf("QRIS payment to %s", decodeResponse.MerchantName),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := repo.Wallet.CreateTransaction(ctx, transaction); err != nil {
		s.log.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create transaction record")

		return nil, err
	}

	newBalance := wallet.Balance - totalAmount
	if err := repo.Wallet.UpdateWalletBalance(ctx, userID, newBalance); err != nil {
		s.log.WithFields(log.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Error("Failed to update wallet balance")

		return nil, err
	}

	if err := repo.Commit(); err != nil {
		s.log.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction")
		return nil, err
	}

	response := &sentrapay.QRISPaymentResponse{
		TransactionID:   transactionID,
		ReferenceNo:     paymentResponse.ReferenceNo,
		Amount:          decodeResponse.Amount,
		FeeAmount:       decodeResponse.FeeAmount,
		TotalAmount:     totalAmount,
		MerchantName:    decodeResponse.MerchantName,
		Status:          "success",
		TransactionDate: paymentResponse.TransactionDate,
		PaymentMethod:   "qris",
		AdditionalInfo: sentrapay.QRISPaymentAdditionalInfo{
			TransactionType:            paymentResponse.AdditionalInfo.TransactionType,
			TransactionTypeDescription: paymentResponse.AdditionalInfo.TransactionTypeDescription,
			Acquirer:                   paymentResponse.AdditionalInfo.Acquirer,
			AcquirerName:               paymentResponse.AdditionalInfo.AcquirerName,
		},
	}

	s.log.WithFields(log.Fields{
		"request_id":     requestID,
		"user_id":        userID,
		"transaction_id": transactionID,
		"amount":         totalAmount,
		"merchant":       decodeResponse.MerchantName,
	}).Info("QRIS payment completed successfully")

	return response, nil
}

func getPaymentType(pointOfInitiation string) string {
	switch pointOfInitiation {
	case "11":
		return "STATIC"
	case "12":
		return "DYNAMIC"
	default:
		return "UNKNOWN"
	}
}
