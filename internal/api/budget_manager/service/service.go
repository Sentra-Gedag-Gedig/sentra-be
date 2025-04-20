package budgetService

import (
	"ProjectGolang/internal/api/budget_manager"
	budgetRepository "ProjectGolang/internal/api/budget_manager/repository"
	"ProjectGolang/internal/entity"
	"ProjectGolang/pkg/s3"
	"ProjectGolang/pkg/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"mime/multipart"
)

type IBudgetService interface {
	CreateTransaction(ctx context.Context, req budget_manager.CreateTransactionRequest, audioFile *multipart.FileHeader) error
	GetTransactionByID(ctx context.Context, id string) (entity.BudgetTransaction, error)
	GetTransactionsByUserID(ctx context.Context, userID string) ([]entity.BudgetTransaction, error)
	GetTransactionsByPeriod(ctx context.Context, userID string, period string) ([]entity.BudgetTransaction, error)
	UpdateTransaction(ctx context.Context, req budget_manager.UpdateTransactionRequest, audioFile *multipart.FileHeader) error
	DeleteTransaction(ctx context.Context, id string, userID string) error
	GetTransactionsByTypeAndCategory(ctx context.Context, userID string, transactionType string, category string) ([]entity.BudgetTransaction, error)
}

type budgetService struct {
	log              *logrus.Logger
	budgetRepository budgetRepository.Repository
	s3               s3.ItfS3
	utils            utils.IUtils
}

func NewBudgetService(log *logrus.Logger, br budgetRepository.Repository, s3 s3.ItfS3, utils utils.IUtils) IBudgetService {
	return &budgetService{
		log:              log,
		budgetRepository: br,
		s3:               s3,
		utils:            utils,
	}
}
