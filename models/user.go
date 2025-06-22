package models

import (
	"time"
)

type User struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Account struct {
	ID              int       `json:"id" db:"id"`
	UserID          int       `json:"user_id" db:"user_id"`
	Balance         float64   `json:"balance" db:"balance"`
	SavingsBalance  float64   `json:"savings_balance" db:"savings_balance"`
	AllowanceIncome float64   `json:"allowance_income" db:"allowance_income"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type Transaction struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	Amount      float64   `json:"amount" db:"amount"`
	Type        string    `json:"type" db:"type"` // "income" or "expense"
	Category    string    `json:"category" db:"category"`
	Description string    `json:"description" db:"description"`
	Date        time.Time `json:"date" db:"date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Student-specific expense categories
var StudentCategories = map[string][]string{
	"expense": {
		"Food & Dining",
		"Transportation",
		"Books & Supplies",
		"Entertainment",
		"Clothing",
		"Health & Fitness",
		"Technology",
		"Miscellaneous",
	},
	"income": {
		"Allowance",
		"Part-time Job",
		"Scholarship",
		"Gift Money",
		"Other Income",
	},
}

// Request/Response structs
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TransactionRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type" binding:"required,oneof=income expense"`
	Category    string  `json:"category" binding:"required"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required"`
}

type TransactionPatchRequest struct {
	Amount      *float64 `json:"amount,omitempty" binding:"omitempty,gt=0"`
	Type        *string  `json:"type,omitempty" binding:"omitempty,oneof=income expense"`
	Category    *string  `json:"category,omitempty"`
	Description *string  `json:"description,omitempty"`
	Date        *string  `json:"date,omitempty"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type Summary struct {
	TotalIncome      float64 `json:"total_income"`
	TotalExpense     float64 `json:"total_expense"`
	CurrentBalance   float64 `json:"current_balance"`
	SavingsBalance   float64 `json:"savings_balance"`
	TransactionCount int     `json:"transaction_count"`
}

type CategoryAnalytics struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Count    int     `json:"count"`
	Type     string  `json:"type"`
}

type SavingsGoal struct {
	ID            int        `json:"id" db:"id"`
	UserID        int        `json:"user_id" db:"user_id"`
	Name          string     `json:"name" db:"name"`
	TargetAmount  float64    `json:"target_amount" db:"target_amount"`
	CurrentAmount float64    `json:"current_amount" db:"current_amount"`
	Deadline      *time.Time `json:"deadline,omitempty" db:"deadline"`
	Description   string     `json:"description" db:"description"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type SavingsTransaction struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	GoalID      *int      `json:"goal_id,omitempty" db:"goal_id"`
	Amount      float64   `json:"amount" db:"amount"`
	Type        string    `json:"type" db:"type"` // "deposit" or "withdrawal"
	Description string    `json:"description" db:"description"`
	Date        time.Time `json:"date" db:"date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type SavingsGoalRequest struct {
	Name         string  `json:"name" binding:"required"`
	TargetAmount float64 `json:"target_amount" binding:"required,gt=0"`
	Deadline     *string `json:"deadline,omitempty"`
	Description  string  `json:"description"`
}

type SavingsTransactionRequest struct {
	GoalID      *int    `json:"goal_id,omitempty"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type" binding:"required,oneof=deposit withdrawal"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required"`
}

type SavingsTransferRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Type        string  `json:"type" binding:"required,oneof=to_savings from_savings"`
	Description string  `json:"description"`
	GoalID      *int    `json:"goal_id,omitempty"`
}
