package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"student-money-manager/models"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetTransactions(c *gin.Context) {
	userID := c.GetInt("user_id")

	// Get query parameters for filtering
	limit := c.DefaultQuery("limit", "5")
	offset := c.DefaultQuery("offset", "0")
	category := c.Query("category")
	transactionType := c.Query("type")

	query := `SELECT id, user_id, amount, type, category, description, date, created_at, updated_at 
			  FROM transactions WHERE user_id = $1`

	args := []interface{}{userID}
	argCount := 1

	if category != "" {
		argCount++
		query += " AND category = $" + strconv.Itoa(argCount)
		args = append(args, category)
	}

	if transactionType != "" {
		argCount++
		query += " AND type = $" + strconv.Itoa(argCount)
		args = append(args, transactionType)
	}

	query += " ORDER BY date DESC, created_at DESC"

	argCount++
	query += " LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	argCount++
	query += " OFFSET $" + strconv.Itoa(argCount)
	args = append(args, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Type, &t.Category,
			&t.Description, &t.Date, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan transaction"})
			return
		}
		transactions = append(transactions, t)
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"count":        len(transactions),
	})
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req models.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	// Create transaction
	var transaction models.Transaction
	query := `INSERT INTO transactions (user_id, amount, type, category, description, date, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			  RETURNING id, user_id, amount, type, category, description, date, created_at, updated_at`

	err = h.db.QueryRow(query, userID, req.Amount, req.Type, req.Category, req.Description, date).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Type,
		&transaction.Category, &transaction.Description, &transaction.Date,
		&transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	// Update account balance
	var balanceChange float64
	if req.Type == "income" {
		balanceChange = req.Amount
	} else {
		balanceChange = -req.Amount
	}

	_, err = h.db.Exec("UPDATE accounts SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2",
		balanceChange, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

func (h *Handler) GetTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")
	transactionID := c.Param("id")

	var transaction models.Transaction
	query := `SELECT id, user_id, amount, type, category, description, date, created_at, updated_at 
			  FROM transactions WHERE id = $1 AND user_id = $2`

	err := h.db.QueryRow(query, transactionID, userID).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Type,
		&transaction.Category, &transaction.Description, &transaction.Date,
		&transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

func (h *Handler) UpdateTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")
	transactionID := c.Param("id")

	var req models.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing transaction
	var oldTransaction models.Transaction
	query := `SELECT id, user_id, amount, type, category, description, date, created_at, updated_at 
			  FROM transactions WHERE id = $1 AND user_id = $2`

	err := h.db.QueryRow(query, transactionID, userID).Scan(
		&oldTransaction.ID, &oldTransaction.UserID, &oldTransaction.Amount, &oldTransaction.Type,
		&oldTransaction.Category, &oldTransaction.Description, &oldTransaction.Date,
		&oldTransaction.CreatedAt, &oldTransaction.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction"})
		return
	}

	// Parse new date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	// Update transaction
	var transaction models.Transaction
	updateQuery := `UPDATE transactions 
					SET amount = $1, type = $2, category = $3, description = $4, date = $5, updated_at = NOW()
					WHERE id = $6 AND user_id = $7
					RETURNING id, user_id, amount, type, category, description, date, created_at, updated_at`

	err = h.db.QueryRow(updateQuery, req.Amount, req.Type, req.Category, req.Description, date, transactionID, userID).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Type,
		&transaction.Category, &transaction.Description, &transaction.Date,
		&transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	// Update account balance
	var oldBalanceChange float64
	if oldTransaction.Type == "income" {
		oldBalanceChange = -oldTransaction.Amount
	} else {
		oldBalanceChange = oldTransaction.Amount
	}

	var newBalanceChange float64
	if req.Type == "income" {
		newBalanceChange = req.Amount
	} else {
		newBalanceChange = -req.Amount
	}

	totalBalanceChange := oldBalanceChange + newBalanceChange

	_, err = h.db.Exec("UPDATE accounts SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2",
		totalBalanceChange, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

func (h *Handler) PatchTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")
	transactionID := c.Param("id")

	var req models.TransactionPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing transaction
	var oldTransaction models.Transaction
	query := `SELECT id, user_id, amount, type, category, description, date, created_at, updated_at 
			  FROM transactions WHERE id = $1 AND user_id = $2`

	err := h.db.QueryRow(query, transactionID, userID).Scan(
		&oldTransaction.ID, &oldTransaction.UserID, &oldTransaction.Amount, &oldTransaction.Type,
		&oldTransaction.Category, &oldTransaction.Description, &oldTransaction.Date,
		&oldTransaction.CreatedAt, &oldTransaction.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction"})
		return
	}

	// Prepare update fields with existing values as defaults
	amount := oldTransaction.Amount
	transactionType := oldTransaction.Type
	category := oldTransaction.Category
	description := oldTransaction.Description
	date := oldTransaction.Date

	// Update only provided fields
	if req.Amount != nil {
		amount = *req.Amount
	}
	if req.Type != nil {
		transactionType = *req.Type
	}
	if req.Category != nil {
		category = *req.Category
	}
	if req.Description != nil {
		description = *req.Description
	}
	if req.Date != nil {
		parsedDate, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
		date = parsedDate
	}

	// Update transaction
	var transaction models.Transaction
	updateQuery := `UPDATE transactions 
					SET amount = $1, type = $2, category = $3, description = $4, date = $5, updated_at = NOW()
					WHERE id = $6 AND user_id = $7
					RETURNING id, user_id, amount, type, category, description, date, created_at, updated_at`

	err = h.db.QueryRow(updateQuery, amount, transactionType, category, description, date, transactionID, userID).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Type,
		&transaction.Category, &transaction.Description, &transaction.Date,
		&transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	// Update account balance only if amount or type changed
	if req.Amount != nil || req.Type != nil {
		var oldBalanceChange float64
		if oldTransaction.Type == "income" {
			oldBalanceChange = -oldTransaction.Amount
		} else {
			oldBalanceChange = oldTransaction.Amount
		}

		var newBalanceChange float64
		if transactionType == "income" {
			newBalanceChange = amount
		} else {
			newBalanceChange = -amount
		}

		totalBalanceChange := oldBalanceChange + newBalanceChange

		_, err = h.db.Exec("UPDATE accounts SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2",
			totalBalanceChange, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
			return
		}
	}

	c.JSON(http.StatusOK, transaction)
}

func (h *Handler) DeleteTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")
	transactionID := c.Param("id")

	// Get existing transaction
	var transaction models.Transaction
	query := `SELECT id, user_id, amount, type, category, description, date, created_at, updated_at 
			  FROM transactions WHERE id = $1 AND user_id = $2`

	err := h.db.QueryRow(query, transactionID, userID).Scan(
		&transaction.ID, &transaction.UserID, &transaction.Amount, &transaction.Type,
		&transaction.Category, &transaction.Description, &transaction.Date,
		&transaction.CreatedAt, &transaction.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction"})
		return
	}

	// Delete transaction
	_, err = h.db.Exec("DELETE FROM transactions WHERE id = $1 AND user_id = $2", transactionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	// Update account balance
	var balanceChange float64
	if transaction.Type == "income" {
		balanceChange = -transaction.Amount
	} else {
		balanceChange = transaction.Amount
	}

	_, err = h.db.Exec("UPDATE accounts SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2",
		balanceChange, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update account balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted successfully"})
}
