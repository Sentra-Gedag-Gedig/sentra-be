package entity

import (
	"ProjectGolang/internal/api/budget_manager"
	"time"
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

type IncomeCategory string

const (
	IncomeCategorySalary     IncomeCategory = "gaji"
	IncomeCategoryBonus      IncomeCategory = "bonus"
	IncomeCategoryInvestment IncomeCategory = "investasi"
	IncomeCategoryPartTime   IncomeCategory = "part time"
)

type ExpenseCategory string

const (
	ExpenseCategoryFood           ExpenseCategory = "makanan"
	ExpenseCategoryDaily          ExpenseCategory = "sehari-hari"
	ExpenseCategoryTransportation ExpenseCategory = "transportasi"
	ExpenseCategorySocial         ExpenseCategory = "sosial"
	ExpenseCategoryHousing        ExpenseCategory = "perumahan"
	ExpenseCategoryGift           ExpenseCategory = "hadiah"
	ExpenseCategoryCommunication  ExpenseCategory = "komunikasi"
	ExpenseCategoryClothing       ExpenseCategory = "pakaian"
	ExpenseCategoryEntertainment  ExpenseCategory = "hiburan"
	ExpenseCategoryAppearance     ExpenseCategory = "tampilan"
	ExpenseCategoryHealth         ExpenseCategory = "kesehatan"
	ExpenseCategoryTax            ExpenseCategory = "pajak"
	ExpenseCategoryEducation      ExpenseCategory = "pendidikan"
	ExpenseCategoryInvestment     ExpenseCategory = "investasi"
	ExpenseCategoryPet            ExpenseCategory = "peliharaan"
	ExpenseCategoryVacation       ExpenseCategory = "liburan"
)

func IsValidIncomeCategory(category string) bool {
	switch IncomeCategory(category) {
	case IncomeCategorySalary, IncomeCategoryBonus, IncomeCategoryInvestment, IncomeCategoryPartTime:
		return true
	default:
		return false
	}
}

func IsValidExpenseCategory(category string) bool {
	switch ExpenseCategory(category) {
	case ExpenseCategoryFood, ExpenseCategoryDaily, ExpenseCategoryTransportation, ExpenseCategorySocial,
		ExpenseCategoryHousing, ExpenseCategoryGift, ExpenseCategoryCommunication, ExpenseCategoryClothing,
		ExpenseCategoryEntertainment, ExpenseCategoryAppearance, ExpenseCategoryHealth, ExpenseCategoryTax,
		ExpenseCategoryEducation, ExpenseCategoryInvestment, ExpenseCategoryPet, ExpenseCategoryVacation:
		return true
	default:
		return false
	}
}

func IsValidCategory(transactionType, category string) bool {
	switch TransactionType(transactionType) {
	case TransactionTypeIncome:
		return IsValidIncomeCategory(category)
	case TransactionTypeExpense:
		return IsValidExpenseCategory(category)
	default:
		return false
	}
}

type BudgetTransaction struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Nominal     float64   `json:"nominal"`
	Type        string    `json:"type"`
	Category    string    `json:"category"`
	AudioLink   string    `json:"audio_link"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (t *BudgetTransaction) Validate() error {
	if t.Type != string(TransactionTypeIncome) && t.Type != string(TransactionTypeExpense) {
		return budget_manager.ErrInvalidTransactionType
	}

	if !IsValidCategory(t.Type, t.Category) {
		return budget_manager.ErrInvalidCategory
	}

	if t.Nominal <= 0 {
		return budget_manager.ErrInvalidAmount
	}

	return nil
}
