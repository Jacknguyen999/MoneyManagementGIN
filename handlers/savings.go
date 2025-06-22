package handlers

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"student-money-manager/models"
	"time"

	"github.com/gin-gonic/gin"
)

// Savings Goals

func (h *Handler) GetSavingsGoals(c *gin.Context) {
	userID := c.GetInt("user_id")

	query := `SELECT g.id, g.user_id, g.name, g.target_amount, 
			  COALESCE(SUM(CASE WHEN st.type = 'deposit' THEN st.amount WHEN st.type = 'withdrawal' THEN -st.amount ELSE 0 END), 0) as current_amount,
			  g.deadline, g.description, g.is_active, g.created_at, g.updated_at 
			  FROM savings_goals g
			  LEFT JOIN savings_transactions st ON g.id = st.goal_id AND st.goal_id IS NOT NULL
			  WHERE g.user_id = $1 AND g.is_active = true 
			  GROUP BY g.id, g.user_id, g.name, g.target_amount, g.deadline, g.description, g.is_active, g.created_at, g.updated_at
			  ORDER BY g.created_at DESC`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch savings goals"})
		return
	}
	defer rows.Close()

	var goals []models.SavingsGoal
	for rows.Next() {
		var goal models.SavingsGoal
		var deadline sql.NullTime

		err := rows.Scan(&goal.ID, &goal.UserID, &goal.Name, &goal.TargetAmount, &goal.CurrentAmount,
			&deadline, &goal.Description, &goal.IsActive, &goal.CreatedAt, &goal.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan savings goal"})
			return
		}

		if deadline.Valid {
			goal.Deadline = &deadline.Time
		}

		goals = append(goals, goal)
	}

	c.JSON(http.StatusOK, gin.H{"goals": goals})
}

func (h *Handler) CreateSavingsGoal(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req models.SavingsGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var goal models.SavingsGoal
	var deadline sql.NullTime

	if req.Deadline != nil && *req.Deadline != "" {
		parsedDeadline, err := time.Parse("2006-01-02", *req.Deadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deadline format. Use YYYY-MM-DD"})
			return
		}
		deadline.Time = parsedDeadline
		deadline.Valid = true
	}

	query := `INSERT INTO savings_goals (user_id, name, target_amount, current_amount, deadline, description, is_active, created_at, updated_at)
			  VALUES ($1, $2, $3, 0, $4, $5, true, NOW(), NOW())
			  RETURNING id, user_id, name, target_amount, current_amount, deadline, description, is_active, created_at, updated_at`

	err := h.db.QueryRow(query, userID, req.Name, req.TargetAmount, deadline, req.Description).Scan(
		&goal.ID, &goal.UserID, &goal.Name, &goal.TargetAmount, &goal.CurrentAmount,
		&deadline, &goal.Description, &goal.IsActive, &goal.CreatedAt, &goal.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create savings goal"})
		return
	}

	if deadline.Valid {
		goal.Deadline = &deadline.Time
	}

	c.JSON(http.StatusCreated, goal)
}

func (h *Handler) UpdateSavingsGoal(c *gin.Context) {
	userID := c.GetInt("user_id")
	goalID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	var req models.SavingsGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var deadline sql.NullTime
	if req.Deadline != nil && *req.Deadline != "" {
		parsedDeadline, err := time.Parse("2006-01-02", *req.Deadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deadline format. Use YYYY-MM-DD"})
			return
		}
		deadline.Time = parsedDeadline
		deadline.Valid = true
	}

	var goal models.SavingsGoal
	query := `UPDATE savings_goals 
			  SET name = $1, target_amount = $2, deadline = $3, description = $4, updated_at = NOW()
			  WHERE id = $5 AND user_id = $6
			  RETURNING id, user_id, name, target_amount, current_amount, deadline, description, is_active, created_at, updated_at`

	err = h.db.QueryRow(query, req.Name, req.TargetAmount, deadline, req.Description, goalID, userID).Scan(
		&goal.ID, &goal.UserID, &goal.Name, &goal.TargetAmount, &goal.CurrentAmount,
		&deadline, &goal.Description, &goal.IsActive, &goal.CreatedAt, &goal.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Savings goal not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update savings goal"})
		return
	}

	if deadline.Valid {
		goal.Deadline = &deadline.Time
	}

	c.JSON(http.StatusOK, goal)
}

func (h *Handler) DeleteSavingsGoal(c *gin.Context) {
	userID := c.GetInt("user_id")
	goalID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid goal ID"})
		return
	}

	query := `UPDATE savings_goals SET is_active = false, updated_at = NOW() 
			  WHERE id = $1 AND user_id = $2`

	result, err := h.db.Exec(query, goalID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete savings goal"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Savings goal not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Savings goal deleted successfully"})
}

// Savings Transactions

func (h *Handler) GetSavingsTransactions(c *gin.Context) {
	userID := c.GetInt("user_id")

	query := `SELECT id, user_id, goal_id, amount, type, description, date, created_at, updated_at
			  FROM savings_transactions WHERE user_id = $1 ORDER BY date DESC, created_at DESC`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch savings transactions"})
		return
	}
	defer rows.Close()

	var transactions []models.SavingsTransaction
	for rows.Next() {
		var transaction models.SavingsTransaction
		var goalID sql.NullInt32

		err := rows.Scan(&transaction.ID, &transaction.UserID, &goalID, &transaction.Amount,
			&transaction.Type, &transaction.Description, &transaction.Date,
			&transaction.CreatedAt, &transaction.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan savings transaction"})
			return
		}

		if goalID.Valid {
			goalIDInt := int(goalID.Int32)
			transaction.GoalID = &goalIDInt
		}

		transactions = append(transactions, transaction)
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

// Transfer between current balance and savings

func (h *Handler) TransferToSavings(c *gin.Context) {
	userID := c.GetInt("user_id")

	// Log raw JSON body first
	rawBody, _ := c.GetRawData()
	fmt.Printf("Raw JSON body: %s\n", string(rawBody))

	// Reset request body for binding
	c.Request.Body = ioutil.NopCloser(strings.NewReader(string(rawBody)))

	var req models.SavingsTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Debug logging
	fmt.Printf("Received transfer request: Amount=%f, Type=%s, GoalID=%v\n", req.Amount, req.Type, req.GoalID)

	// Additional debug for goal ID
	if req.GoalID != nil {
		fmt.Printf("Goal ID is provided: %d\n", *req.GoalID)
	} else {
		fmt.Printf("No Goal ID provided - this withdrawal will not affect specific goals\n")
	}

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// Get current account balances
	var currentBalance, savingsBalance float64
	query := `SELECT balance, savings_balance FROM accounts WHERE user_id = $1`
	err = tx.QueryRow(query, userID).Scan(&currentBalance, &savingsBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get account balance"})
		return
	}

	var newCurrentBalance, newSavingsBalance float64
	var transactionType string
	var description string

	if req.Type == "to_savings" {
		// Transfer from current balance to savings
		if currentBalance < req.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient current balance"})
			return
		}
		newCurrentBalance = currentBalance - req.Amount
		newSavingsBalance = savingsBalance + req.Amount
		transactionType = "deposit"
		if req.Description == "" {
			description = "Transfer from current balance"
		} else {
			description = req.Description
		}
	} else { // from_savings
		// Transfer from savings to current balance
		if savingsBalance < req.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient savings balance"})
			return
		}

		// If no specific goal is selected, we need to handle withdrawals from allocated goals
		if req.GoalID == nil {
			// Get total amount allocated to active goals
			var totalGoalAmount float64
			goalQuery := `SELECT COALESCE(SUM(CASE WHEN st.type = 'deposit' THEN st.amount WHEN st.type = 'withdrawal' THEN -st.amount ELSE 0 END), 0) as total_goal_amount
						  FROM savings_goals g
						  LEFT JOIN savings_transactions st ON g.id = st.goal_id
						  WHERE g.user_id = $1 AND g.is_active = true`
			err = tx.QueryRow(goalQuery, userID).Scan(&totalGoalAmount)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate goal amounts"})
				return
			}

			// Check if withdrawal amount exceeds unallocated savings
			unallocatedSavings := savingsBalance - totalGoalAmount
			if req.Amount > unallocatedSavings && totalGoalAmount > 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Cannot withdraw $%.2f from general savings. Only $%.2f available in unallocated savings. Please select a specific goal to withdraw from or reduce the amount.", req.Amount, unallocatedSavings),
				})
				return
			}
		} else {
			// Validate goal-specific withdrawal
			var goalCurrentAmount float64
			goalQuery := `SELECT COALESCE(SUM(CASE WHEN st.type = 'deposit' THEN st.amount WHEN st.type = 'withdrawal' THEN -st.amount ELSE 0 END), 0) as current_amount
						  FROM savings_goals g
						  LEFT JOIN savings_transactions st ON g.id = st.goal_id
						  WHERE g.id = $1 AND g.user_id = $2 AND g.is_active = true`
			err = tx.QueryRow(goalQuery, *req.GoalID, userID).Scan(&goalCurrentAmount)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid savings goal or goal not found"})
				return
			}

			if req.Amount > goalCurrentAmount {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Cannot withdraw $%.2f from this goal. Only $%.2f available in this goal.", req.Amount, goalCurrentAmount),
				})
				return
			}
		}

		newCurrentBalance = currentBalance + req.Amount
		newSavingsBalance = savingsBalance - req.Amount
		transactionType = "withdrawal"
		if req.Description == "" {
			description = "Transfer to current balance"
		} else {
			description = req.Description
		}
	}

	// Update account balances
	updateQuery := `UPDATE accounts SET balance = $1, savings_balance = $2, updated_at = NOW() WHERE user_id = $3`
	_, err = tx.Exec(updateQuery, newCurrentBalance, newSavingsBalance, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balances"})
		return
	}

	// Create savings transaction record
	var savingsTransaction models.SavingsTransaction
	savingsQuery := `INSERT INTO savings_transactions (user_id, goal_id, amount, type, description, date, created_at, updated_at)
					 VALUES ($1, $2, $3, $4, $5, CURRENT_DATE, NOW(), NOW())
					 RETURNING id, user_id, goal_id, amount, type, description, date, created_at, updated_at`

	var goalID sql.NullInt32
	if req.GoalID != nil {
		goalID.Int32 = int32(*req.GoalID)
		goalID.Valid = true
		fmt.Printf("Setting goalID to: %d (valid: %t)\n", goalID.Int32, goalID.Valid)
	} else {
		fmt.Printf("No goalID provided, setting to NULL\n")
	}

	var scannedGoalID sql.NullInt32
	err = tx.QueryRow(savingsQuery, userID, goalID, req.Amount, transactionType, description).Scan(
		&savingsTransaction.ID, &savingsTransaction.UserID, &scannedGoalID,
		&savingsTransaction.Amount, &savingsTransaction.Type, &savingsTransaction.Description,
		&savingsTransaction.Date, &savingsTransaction.CreatedAt, &savingsTransaction.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create savings transaction"})
		return
	}

	// Convert scanned goal_id back to model field
	if scannedGoalID.Valid {
		goalIDInt := int(scannedGoalID.Int32)
		savingsTransaction.GoalID = &goalIDInt
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Transfer completed successfully",
		"savings_transaction": savingsTransaction,
		"new_current_balance": newCurrentBalance,
		"new_savings_balance": newSavingsBalance,
	})
}
