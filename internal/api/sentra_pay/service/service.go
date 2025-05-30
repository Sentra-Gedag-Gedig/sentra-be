package sentrapayService

import (
	authRepository "ProjectGolang/internal/api/auth/repository"
	sentrapay "ProjectGolang/internal/api/sentra_pay"
	sentrapayRepository "ProjectGolang/internal/api/sentra_pay/repository"
	"ProjectGolang/pkg/bcrypt"
	"ProjectGolang/pkg/doku"
	"ProjectGolang/pkg/utils"
	"context"
	"github.com/sirupsen/logrus"
)

type ISentraPayService interface {
	CreateTopUpTransaction(ctx context.Context, userID string, req sentrapay.TopUpRequest) (*sentrapay.TopUpResponse, error)
	ProcessPaymentCallback(ctx context.Context, req sentrapay.PaymentCallbackRequest, channelID, xExternalID, xTimestamp, xPartnerID string) error
	GetWalletBalance(ctx context.Context, userID string) (*sentrapay.WalletBalance, error)
	GetTransactionHistory(ctx context.Context, userID string, page, limit int) (*sentrapay.TransactionHistoryResponse, error)
	CheckTransactionStatus(ctx context.Context, referenceNo string) (string, error)

	DecodeQRIS(ctx context.Context, req sentrapay.QRISDecodeRequest) (*sentrapay.QRISDecodeResponse, error)
	PaymentQRIS(ctx context.Context, userID string, req sentrapay.QRISPaymentRequest) (*sentrapay.QRISPaymentResponse, error)
}

type sentraPayService struct {
	log              *logrus.Logger
	walletRepository sentrapayRepository.Repository
	dokuService      doku.IDokuService
	authRepo         authRepository.Repository
	utils            utils.IUtils
	bcryptUtils      bcrypt.IBcrypt
}

func NewSentraPayService(
	log *logrus.Logger,
	wr sentrapayRepository.Repository,
	ds doku.IDokuService,
	ar authRepository.Repository,
	utils utils.IUtils,
	bcryptUtils bcrypt.IBcrypt,
) ISentraPayService {
	return &sentraPayService{
		log:              log,
		walletRepository: wr,
		dokuService:      ds,
		authRepo:         ar,
		utils:            utils,
		bcryptUtils:      bcryptUtils,
	}
}
