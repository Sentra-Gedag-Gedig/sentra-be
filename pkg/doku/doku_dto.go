package doku

import (
	"encoding/json"
	"time"
)

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

type DecodeQRISRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	QRContent          string `json:"qrContent"`
	ScanTime           string `json:"scanTime"`
}

type DecodeQRISResponse struct {
	ResponseCode       string         `json:"responseCode"`
	ResponseMessage    string         `json:"responseMessage"`
	ReferenceNo        string         `json:"referenceNo"`
	PartnerReferenceNo string         `json:"partnerReferenceNo"`
	MerchantName       string         `json:"merchantName"`
	TransactionAmount  Amount         `json:"transactionAmount"`
	FeeAmount          Amount         `json:"feeAmount"`
	AdditionalInfo     AdditionalInfo `json:"additionalInfo"`
}

type PaymentQRISRequest struct {
	PartnerReferenceNo string                    `json:"partnerReferenceNo"`
	Amount             Amount                    `json:"amount"`
	FeeAmount          Amount                    `json:"feeAmount"`
	AdditionalInfo     PaymentQRISAdditionalInfo `json:"additionalInfo"`
}

type PaymentQRISResponse struct {
	ResponseCode       string                        `json:"responseCode"`
	ResponseMessage    string                        `json:"responseMessage"`
	ReferenceNo        string                        `json:"referenceNo"`
	PartnerReferenceNo string                        `json:"partnerReferenceNo"`
	TransactionDate    string                        `json:"transactionDate"`
	Original           OriginalQRISTransaction       `json:"original"`
	AdditionalInfo     PaymentQRISResponseAdditional `json:"additionalInfo"`
}

type Amount struct {
	Value    json.Number `json:"value"`
	Currency string      `json:"currency"`
}

type AdditionalInfo struct {
	PointOfInitiationMethod            string `json:"pointOfInitiationMethod"`
	PointOfInitiationMethodDescription string `json:"pointOfInitiationMethodDescription"`
	FeeType                            string `json:"feeType"`
	FeeTypeDescription                 string `json:"feeTypeDescription"`
}

type PaymentQRISAdditionalInfo struct {
	QRContent string        `json:"qrContent"`
	Origin    PaymentOrigin `json:"origin"`
}

type PaymentOrigin struct {
	Product       string `json:"product"`
	Source        string `json:"source"`
	SourceVersion string `json:"sourceVersion"`
	System        string `json:"system"`
	ApiFormat     string `json:"apiFormat"`
}

type OriginalQRISTransaction struct {
	ReferenceNo string `json:"referenceNo"`
	Amount      Amount `json:"amount"`
	FeeAmount   Amount `json:"feeAmount"`
}

type PaymentQRISResponseAdditional struct {
	TransactionType            string `json:"transactionType"`
	TransactionTypeDescription string `json:"transactionTypeDescription"`
	Acquirer                   string `json:"acquirer"`
	AcquirerName               string `json:"acquirerName"`
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

type QRISPaymentRequest struct {
	PartnerReferenceNo string                 `json:"partnerReferenceNo"`
	Amount             float64                `json:"amount"`
	PaymentType        string                 `json:"paymentType"`
	AdditionalInfo     map[string]interface{} `json:"additionalInfo"`
}

type QRISPaymentResponse struct {
	ReferenceNo     string `json:"referenceNo"`
	PaymentURL      string `json:"paymentUrl"`
	QRISRedirectURL string `json:"qrisRedirectUrl"`
	Status          string `json:"status"`
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
}
