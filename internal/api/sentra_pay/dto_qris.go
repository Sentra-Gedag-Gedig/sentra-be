package sentrapay

type QRISDecodeRequest struct {
	QRContent string `json:"qr_content" validate:"required"`
}

type QRISDecodeResponse struct {
	ReferenceNo    string                   `json:"reference_no"`
	MerchantName   string                   `json:"merchant_name"`
	Amount         float64                  `json:"amount"`
	FeeAmount      float64                  `json:"fee_amount"`
	TotalAmount    float64                  `json:"total_amount"`
	PaymentType    string                   `json:"payment_type"`
	AdditionalInfo QRISDecodeAdditionalInfo `json:"additional_info"`
}

type QRISDecodeAdditionalInfo struct {
	PointOfInitiationMethod            string `json:"point_of_initiation_method"`
	PointOfInitiationMethodDescription string `json:"point_of_initiation_method_description"`
	FeeType                            string `json:"fee_type"`
	FeeTypeDescription                 string `json:"fee_type_description"`
}

type QRISPaymentRequest struct {
	QRContent string `json:"qr_content" validate:"required"`
	AuthCode  string `json:"auth_code""`
}

type QRISPaymentResponse struct {
	TransactionID   string                    `json:"transaction_id"`
	ReferenceNo     string                    `json:"reference_no"`
	Amount          float64                   `json:"amount"`
	FeeAmount       float64                   `json:"fee_amount"`
	TotalAmount     float64                   `json:"total_amount"`
	MerchantName    string                    `json:"merchant_name"`
	Status          string                    `json:"status"`
	TransactionDate string                    `json:"transaction_date"`
	PaymentMethod   string                    `json:"payment_method"`
	AdditionalInfo  QRISPaymentAdditionalInfo `json:"additional_info"`
}

type QRISPaymentAdditionalInfo struct {
	TransactionType            string `json:"transaction_type"`
	TransactionTypeDescription string `json:"transaction_type_description"`
	Acquirer                   string `json:"acquirer"`
	AcquirerName               string `json:"acquirer_name"`
}
