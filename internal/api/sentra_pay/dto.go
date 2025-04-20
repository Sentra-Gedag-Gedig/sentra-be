package sentrapay

import (
	"time"
)

type TopUpRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
	Bank   string  `json:"bank" validate:"required"`
}

type TopUpResponse struct {
	TransactionID   string    `json:"transaction_id"`
	ReferenceNo     string    `json:"reference_no"`
	VirtualAccount  string    `json:"virtual_account"`
	Bank            string    `json:"bank"`
	Amount          float64   `json:"amount"`
	ExpiresAt       string    `json:"expires_at"`
	PaymentGuideURL string    `json:"payment_guide_url"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type PaymentCallbackRequest struct {
	PartnerServiceId    string         `json:"partnerServiceId"`
	CustomerNo          string         `json:"customerNo"`
	VirtualAccountNo    string         `json:"virtualAccountNo"`
	VirtualAccountName  string         `json:"virtualAccountName"`
	VirtualAccountEmail string         `json:"virtualAccountEmail"`
	VirtualAccountPhone string         `json:"virtualAccountPhone"`
	TrxId               string         `json:"trxId"`
	PaymentRequestId    string         `json:"paymentRequestId"`
	PaidAmount          Amount         `json:"paidAmount"`
	TotalAmount         Amount         `json:"totalAmount"`
	TrxDateTime         string         `json:"trxDateTime"`
	AdditionalInfo      AdditionalInfo `json:"additionalInfo"`
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type AdditionalInfo struct {
	Channel         string `json:"channel"`
	SenderName      string `json:"senderName"`
	SourceAccountNo string `json:"sourceAccountNo"`
	SourceBankCode  string `json:"sourceBankCode"`
	SourceBankName  string `json:"sourceBankName"`
}

type WalletTransaction struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Amount        float64   `json:"amount"`
	Type          string    `json:"type"`
	ReferenceNo   string    `json:"reference_no"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`
	BankAccount   string    `json:"bank_account,omitempty"`
	BankName      string    `json:"bank_name,omitempty"`
	Description   string    `json:"description,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type WalletBalance struct {
	UserID      string    `json:"user_id"`
	Balance     float64   `json:"balance"`
	LastUpdated time.Time `json:"last_updated"`
}

type TransactionHistoryResponse struct {
	Transactions []WalletTransaction `json:"transactions"`
	Total        int                 `json:"total"`
}
