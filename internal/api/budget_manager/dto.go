package budget_manager

type CreateTransactionRequest struct {
	UserID      string  `json:"user_id" validate:"required"`
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description"`
	Nominal     float64 `json:"nominal" validate:"required,gt=0"`
	Type        string  `json:"type" validate:"required,oneof=income expense"`
	Category    string  `json:"category" validate:"required"`
}

type UpdateTransactionRequest struct {
	ID          string  `json:"id" validate:"required"`
	UserID      string  `json:"user_id" validate:"required"`
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description"`
	Nominal     float64 `json:"nominal" validate:"required,gt=0"`
	Type        string  `json:"type" validate:"required,oneof=income expense"`
	Category    string  `json:"category" validate:"required"`
	DeleteAudio bool    `json:"delete_audio"`
}

type TransactionResponse struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Nominal     float64 `json:"nominal"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	AudioLink   string  `json:"audio_link,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	TotalIncome  float64               `json:"total_income"`
	TotalExpense float64               `json:"total_expense"`
	Balance      float64               `json:"balance"`
}
