package doku

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/PTNUSASATUINTIARTHA-DOKU/doku-golang-library/controllers"
	"github.com/PTNUSASATUINTIARTHA-DOKU/doku-golang-library/doku"
	checkVaModels "github.com/PTNUSASATUINTIARTHA-DOKU/doku-golang-library/models/va/checkVa"
	createVa "github.com/PTNUSASATUINTIARTHA-DOKU/doku-golang-library/models/va/createVa"
	"github.com/PTNUSASATUINTIARTHA-DOKU/doku-golang-library/models/va/notification/payment"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
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
	DecodeQRIS(qrContent string) (*DecodeQRISResponse, error)
	PaymentQRIS(qrContent string, transactionAmount, feeAmount float64, authCode string) (*PaymentQRISResponse, error)
}

type dokuService struct {
	client *doku.Snap
	log    *logrus.Logger
}

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

func (d *dokuService) DecodeQRIS(qrContent string) (*DecodeQRISResponse, error) {
	partnerReferenceNo := fmt.Sprintf("QRIS%d", time.Now().Unix())

	request := DecodeQRISRequest{
		PartnerReferenceNo: partnerReferenceNo,
		QRContent:          qrContent,
		ScanTime:           time.Now().Format("2006-01-02T15:04:05-07:00"),
	}

	tokenB2B := d.client.GetTokenB2B()
	if tokenB2B.ResponseCode != "2007300" {
		return nil, fmt.Errorf("failed to get B2B token: %s", tokenB2B.ResponseMessage)
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05-07:00")
	externalId := fmt.Sprintf("%d", time.Now().UnixNano())

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	endpointUrl := "/snap-adapter/b2b/v1.0/qr/qr-mpm-decode"
	signature := generateSignature("POST", endpointUrl, tokenB2B.AccessToken, reqBody, timestamp, d.client.SecretKey)

	var url string
	if d.client.IsProduction {
		url = "https://api.doku.com"
	} else {
		url = "https://api-sandbox.doku.com"
	}
	url += endpointUrl

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PARTNER-ID", d.client.ClientId)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-EXTERNAL-ID", externalId)
	req.Header.Set("X-SIGNATURE", signature)
	req.Header.Set("Authorization", "Bearer "+tokenB2B.AccessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	d.log.WithFields(logrus.Fields{
		"response_raw": string(respBody),
	}).Debug("QRIS decode raw response")

	var response DecodeQRISResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		d.log.WithFields(logrus.Fields{
			"error":        err.Error(),
			"response_raw": string(respBody),
		}).Error("Failed to unmarshal QRIS response")
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if response.ResponseCode != "2004800" {
		return nil, fmt.Errorf("decode QRIS failed: %s", response.ResponseMessage)
	}

	return &response, nil
}

func (d *dokuService) PaymentQRIS(qrContent string, transactionAmount, feeAmount float64, authCode string) (*PaymentQRISResponse, error) {
	partnerReferenceNo := fmt.Sprintf("PAY%d", time.Now().Unix())

	request := PaymentQRISRequest{
		PartnerReferenceNo: partnerReferenceNo,
		Amount: Amount{
			Value:    json.Number(fmt.Sprintf("%.2f", transactionAmount)),
			Currency: "IDR",
		},
		FeeAmount: Amount{
			Value:    json.Number(fmt.Sprintf("%.2f", feeAmount)),
			Currency: "IDR",
		},
		AdditionalInfo: PaymentQRISAdditionalInfo{
			QRContent: qrContent,
			Origin: PaymentOrigin{
				Product:       "SDK",
				Source:        "Golang",
				SourceVersion: "1.0.0",
				System:        "sentra-pay",
				ApiFormat:     "SNAP",
			},
		},
	}

	tokenB2B := d.client.GetTokenB2B()
	if tokenB2B.ResponseCode != "2007300" {
		return nil, fmt.Errorf("failed to get B2B token: %s", tokenB2B.ResponseMessage)
	}

	tokenB2B2C, err := d.client.GetTokenB2B2C(authCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get B2B2C token: %v", err)
	}

	if tokenB2B2C.ResponseCode != "2007300" {
		return nil, fmt.Errorf("invalid B2B2C token response: %s", tokenB2B2C.ResponseMessage)
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05-07:00")
	externalId := fmt.Sprintf("%d", time.Now().UnixNano())

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	endpointUrl := "/snap-adapter/b2b2c/v1.0/qr/qr-mpm-payment"
	signature := generateSignature("POST", endpointUrl, tokenB2B.AccessToken, reqBody, timestamp, d.client.SecretKey)

	var url string
	if d.client.IsProduction {
		url = "https://api.doku.com"
	} else {
		url = "https://api-sandbox.doku.com"
	}
	url += endpointUrl

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PARTNER-ID", d.client.ClientId)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-EXTERNAL-ID", externalId)
	req.Header.Set("X-SIGNATURE", signature)
	req.Header.Set("Authorization", "Bearer "+tokenB2B.AccessToken)
	req.Header.Set("Authorization-Customer", "Bearer "+tokenB2B2C.AccessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	d.log.Debug(fmt.Sprintf("[response_raw:%s] QRIS payment raw response", string(respBody)))

	var response PaymentQRISResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if response.ResponseCode != "2005500" {
		d.log.Warn(fmt.Sprintf("[response_code:%s] [response_message:%s] Unexpected response code for QRIS payment",
			response.ResponseCode, response.ResponseMessage))
		return nil, fmt.Errorf("payment QRIS failed: %s", response.ResponseMessage)
	}

	return &response, nil
}

func generateSignature(httpMethod, endpointUrl, accessToken string, minifiedRequestBody []byte, timestamp, secretKey string) string {

	hash := sha256.New()
	hash.Write(minifiedRequestBody)
	hashedBody := hex.EncodeToString(hash.Sum(nil))

	stringToSign := httpMethod + ":" + endpointUrl + ":" + accessToken + ":" + strings.ToLower(hashedBody) + ":" + timestamp

	hmac := hmac.New(sha512.New, []byte(secretKey))
	hmac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(hmac.Sum(nil))

	return signature
}
