package doku

import (
	"fmt"
	"github.com/PTNUSASATUINTIARTHA-DOKU/doku-golang-library/controllers"
	"github.com/PTNUSASATUINTIARTHA-DOKU/doku-golang-library/doku"
	checkVaModels "github.com/PTNUSASATUINTIARTHA-DOKU/doku-golang-library/models/va/checkVa"
	createVa "github.com/PTNUSASATUINTIARTHA-DOKU/doku-golang-library/models/va/createVa"
	"github.com/PTNUSASATUINTIARTHA-DOKU/doku-golang-library/models/va/notification/payment"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
	"time"
)

type IDokuService interface {
	Init() error
	CreateVirtualAccount(req CreateVaRequest) (*CreateVaResponse, error)
	ValidateCallback(token string, notification payment.PaymentNotificationRequestBodyDTO) (payment.PaymentNotificationResponseBodyDTO, error)
	CheckVAStatus(vaNumber string, customerNo string, partnerServiceId string, trxId string) (bool, error)
}

type dokuService struct {
	client *doku.Snap
	log    *logrus.Logger
}

type CreateVaRequest struct {
	UserID          string
	Name            string
	Email           string
	Phone           string
	Amount          float64
	TrxId           string
	Bank            string
	ExpiredDuration time.Duration
	ReusableStatus  bool
}

type CreateVaResponse struct {
	VirtualAccountNo  string
	Bank              string
	Amount            float64
	TransactionID     string
	ExpiryDate        string
	VirtualAccountURL string
}

const (
	BankBCA      = "VIRTUAL_ACCOUNT_BCA"
	BankMANDIRI  = "VIRTUAL_ACCOUNT_BANK_MANDIRI"
	BankBRI      = "VIRTUAL_ACCOUNT_BRI"
	BankBNI      = "VIRTUAL_ACCOUNT_BNI"
	BankDANAMON  = "VIRTUAL_ACCOUNT_BANK_DANAMON"
	BankPERMATA  = "VIRTUAL_ACCOUNT_BANK_PERMATA"
	BankMAYBANK  = "VIRTUAL_ACCOUNT_MAYBANK"
	BankBTN      = "VIRTUAL_ACCOUNT_BTN"
	BankBSI      = "VIRTUAL_ACCOUNT_BSI"
	BankCIMB     = "VIRTUAL_ACCOUNT_BANK_CIMB"
	BankSINARMAS = "VIRTUAL_ACCOUNT_SINARMAS"
	BankDOKU     = "VIRTUAL_ACCOUNT_DOKU"
)

func NewDokuService(log *logrus.Logger) IDokuService {
	return &dokuService{
		log: log,
	}
}

func (d *dokuService) Init() error {
	d.log.WithFields(logrus.Fields{
		"client_id":     os.Getenv("DOKU_CLIENT_ID"),
		"is_production": os.Getenv("DOKU_IS_PRODUCTION"),
	}).Info("Initializing Doku client")

	var privateKey string

	privateKeyPEM, err := os.ReadFile("private.key")
	if err != nil {
		return fmt.Errorf("failed to read private key file: %v", err)
	}
	privateKey = strings.TrimSpace(string(privateKeyPEM))

	if !strings.Contains(privateKey, "-----BEGIN") {
		return fmt.Errorf("invalid private key format")
	}

	d.client = &doku.Snap{
		PrivateKey: privateKey,
		PublicKey:  os.Getenv("DOKU_PUBLIC_KEY"),
		ClientId:   os.Getenv("DOKU_CLIENT_ID"),
		SecretKey:  os.Getenv("DOKU_SECRET_KEY"),
		IsProduction: func() bool {
			isProd, _ := strconv.ParseBool(os.Getenv("DOKU_IS_PRODUCTION"))
			return isProd
		}(),
	}

	doku.TokenController = &controllers.TokenController{}
	doku.VaController = &controllers.VaController{}
	doku.NotificationController = &controllers.NotificationController{}
	doku.DirectDebitController = &controllers.DirectDebitController{}

	d.log.Info("About to get token from DOKU API")

	response := d.client.GetTokenB2B()

	if response.AccessToken != "" {
		tokenPreview := response.AccessToken[:10] + "..."
		d.log.WithFields(logrus.Fields{
			"token_preview": tokenPreview,
		}).Info("Doku access token preview")
	}

	if response.ResponseCode != "2007300" {
		return fmt.Errorf("failed to initialize Doku client: %s", response.ResponseMessage)
	}

	return nil
}

func (d *dokuService) CreateVirtualAccount(req CreateVaRequest) (*CreateVaResponse, error) {
	amountStr := fmt.Sprintf("%.2f", req.Amount)

	customerNo := "3"

	partnerServiceId := "   84923"

	virtualAccountNo := partnerServiceId + customerNo

	loc, _ := time.LoadLocation("Asia/Jakarta")
	expiredTime := time.Now().In(loc).Add(req.ExpiredDuration)
	expiredDate := expiredTime.Format("2006-01-02T15:04:05") + "+07:00"
	fmt.Println("ExpiredDate:", expiredDate)

	createVaRequest := createVa.CreateVaRequestDto{
		PartnerServiceId:    partnerServiceId,
		CustomerNo:          customerNo,
		VirtualAccountNo:    virtualAccountNo,
		VirtualAccountName:  req.Name,
		VirtualAccountEmail: req.Email,
		VirtualAccountPhone: req.Phone,
		TrxId:               req.TrxId,
		TotalAmount: createVa.TotalAmount{
			Value:    amountStr,
			Currency: "IDR",
		},
		AdditionalInfo: createVa.AdditionalInfo{
			Channel: req.Bank,
			VirtualAccountConfig: createVa.VirtualAccountConfig{
				ReusableStatus: req.ReusableStatus,
			},
		},
		VirtualAccountTrxType: "C",
		ExpiredDate:           expiredDate,
	}

	response, err := d.client.CreateVa(createVaRequest)
	if err != nil {
		d.log.WithError(err).Error("Failed to create virtual account")
		return nil, err
	}

	if response.ResponseCode != "2002500" && response.ResponseCode != "2002700" {
		d.log.WithFields(logrus.Fields{
			"response_code":    response.ResponseCode,
			"response_message": response.ResponseMessage,
		}).Error("Failed to create virtual account")
		return nil, fmt.Errorf("failed to create virtual account: %s", response.ResponseMessage)
	}

	if response.VirtualAccountData == nil {
		return nil, fmt.Errorf("virtual account data is nil")
	}

	return &CreateVaResponse{
		VirtualAccountNo:  response.VirtualAccountData.VirtualAccountNo,
		Bank:              req.Bank,
		Amount:            req.Amount,
		TransactionID:     req.TrxId,
		ExpiryDate:        createVaRequest.ExpiredDate,
		VirtualAccountURL: response.VirtualAccountData.AdditionalInfo.HowToPayPage,
	}, nil
}

func (d *dokuService) ValidateCallback(token string, notification payment.PaymentNotificationRequestBodyDTO) (payment.PaymentNotificationResponseBodyDTO, error) {
	response, err := d.client.ValidateTokenAndGenerateNotificationResponse(token, notification)
	if err != nil {
		d.log.WithError(err).Error("Failed to validate callback")
		return payment.PaymentNotificationResponseBodyDTO{}, err
	}

	return response, nil
}

func (d *dokuService) CheckVAStatus(vaNumber string, customerNo string, partnerServiceId string, trxId string) (bool, error) {
	checkStatusRequest := checkVaModels.CheckStatusVARequestDto{
		PartnerServiceId: partnerServiceId,
		CustomerNo:       customerNo,
		VirtualAccountNo: vaNumber,
	}

	response, err := d.client.CheckStatusVa(checkStatusRequest)
	if err != nil {
		d.log.WithError(err).Error("Failed to check VA status")
		return false, err
	}

	if (response.ResponseCode == "2002600" || response.ResponseCode == "2002400") && response.VirtualAccountData != nil {
		if response.VirtualAccountData.PaidAmount.Value != "0.00" {
			return true, nil
		}
	}

	return false, nil
}
