package handlers

import (
	"database/sql"
	"net/http"
	"student-money-manager/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetAccount(c *gin.Context) {
	userID := c.GetInt("user_id")

	var account models.Account
	query := `SELECT id, user_id, balance, savings_balance, allowance_income, created_at, updated_at 
			  FROM accounts WHERE user_id = $1`

	err := h.db.QueryRow(query, userID).Scan(
		&account.ID, &account.UserID, &account.Balance, &account.SavingsBalance, &account.AllowanceIncome,
		&account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch account"})
		return
	}

	c.JSON(http.StatusOK, account)
}

func (h *Handler) UpdateAccount(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req struct {
		AllowanceIncome float64 `json:"allowance_income" binding:"required,gte=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var account models.Account
	query := `UPDATE accounts 
			  SET allowance_income = $1, updated_at = NOW()
			  WHERE user_id = $2
			  RETURNING id, user_id, balance, allowance_income, created_at, updated_at`

	err := h.db.QueryRow(query, req.AllowanceIncome, userID).Scan(
		&account.ID, &account.UserID, &account.Balance, &account.AllowanceIncome,
		&account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account"})
		return
	}

	c.JSON(http.StatusOK, account)
}

func (h *Handler) ProcessAutoAllowance(c *gin.Context) {
	userID := c.GetInt("user_id")

	// Get user's account info
	var account models.Account
	query := `SELECT id, user_id, balance, allowance_income FROM accounts WHERE user_id = $1`
	err := h.db.QueryRow(query, userID).Scan(&account.ID, &account.UserID, &account.Balance, &account.AllowanceIncome)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch account"})
		return
	}

	if account.AllowanceIncome <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No allowance amount set"})
		return
	}

	// Check if allowance for current month already exists
	checkQuery := `SELECT COUNT(*) FROM transactions 
				   WHERE user_id = $1 
				   AND type = 'income' 
				   AND category = 'Allowance' 
				   AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())`

	var count int
	err = h.db.QueryRow(checkQuery, userID).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing allowance"})
		return
	}

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Allowance already processed for this month"})
		return
	}

	// Create allowance transaction
	transactionQuery := `INSERT INTO transactions (user_id, amount, type, category, description, date, created_at, updated_at)
						 VALUES ($1, $2, 'income', 'Allowance', 'Monthly allowance - auto-added', CURRENT_DATE, NOW(), NOW())
						 RETURNING id, user_id, amount, type, category, description, date, created_at, updated_at`

	var transaction models.Transaction
	err = h.db.QueryRow(transactionQuery, userID, account.AllowanceIncome).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Type,
		&transaction.Category, &transaction.Description, &transaction.Date,
		&transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create allowance transaction"})
		return
	}

	// Update account balance
	updateBalanceQuery := `UPDATE accounts SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2`
	_, err = h.db.Exec(updateBalanceQuery, account.AllowanceIncome, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Monthly allowance added successfully",
		"transaction": transaction,
		"amount":      account.AllowanceIncome,
	})
}

func (h *Handler) GetSummary(c *gin.Context) {
	userID := c.GetInt("user_id")

	var summary models.Summary

	// Get account balance and savings balance
	err := h.db.QueryRow("SELECT balance, savings_balance FROM accounts WHERE user_id = $1", userID).Scan(&summary.CurrentBalance, &summary.SavingsBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch account balance"})
		return
	}

	// Get income and expense totals
	query := `SELECT 
				COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as total_income,
				COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as total_expense,
				COUNT(*) as transaction_count
			  FROM transactions WHERE user_id = $1`

	err = h.db.QueryRow(query, userID).Scan(&summary.TotalIncome, &summary.TotalExpense, &summary.TransactionCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *Handler) GetCategoryAnalytics(c *gin.Context) {
	userID := c.GetInt("user_id")

	query := `SELECT category, type, SUM(amount) as total_amount, COUNT(*) as count
			  FROM transactions 
			  WHERE user_id = $1 
			  GROUP BY category, type 
			  ORDER BY total_amount DESC`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics"})
		return
	}
	defer rows.Close()

	var analytics []models.CategoryAnalytics
	for rows.Next() {
		var ca models.CategoryAnalytics
		err := rows.Scan(&ca.Category, &ca.Type, &ca.Amount, &ca.Count)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan analytics"})
			return
		}
		analytics = append(analytics, ca)
	}

	c.JSON(http.StatusOK, gin.H{
		"analytics":  analytics,
		"categories": models.StudentCategories,
	})
}
