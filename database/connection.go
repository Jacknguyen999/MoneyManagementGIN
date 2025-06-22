package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func Connect() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "student_money_db")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	// Create users table
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create accounts table
	accountsTable := `
	CREATE TABLE IF NOT EXISTS accounts (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		balance DECIMAL(20,2) DEFAULT 0.00,
		savings_balance DECIMAL(20,2) DEFAULT 0.00,
		allowance_income DECIMAL(20,2) DEFAULT 0.00,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create transactions table
	transactionsTable := `
	CREATE TABLE IF NOT EXISTS transactions (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		amount DECIMAL(20,2) NOT NULL,
		type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
		category VARCHAR(100) NOT NULL,
		description TEXT,
		date TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create savings_goals table
	savingsGoalsTable := `
	CREATE TABLE IF NOT EXISTS savings_goals (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		target_amount DECIMAL(20,2) NOT NULL,
		current_amount DECIMAL(20,2) DEFAULT 0.00,
		deadline DATE,
		description TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create savings_transactions table
	savingsTransactionsTable := `
	CREATE TABLE IF NOT EXISTS savings_transactions (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
		goal_id INTEGER REFERENCES savings_goals(id) ON DELETE SET NULL,
		amount DECIMAL(20,2) NOT NULL,
		type VARCHAR(20) NOT NULL CHECK (type IN ('deposit', 'withdrawal')),
		description TEXT,
		date TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Create indexes for better performance
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category);`,
		`CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_savings_goals_user_id ON savings_goals(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_savings_goals_active ON savings_goals(is_active);`,
		`CREATE INDEX IF NOT EXISTS idx_savings_transactions_user_id ON savings_transactions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_savings_transactions_goal_id ON savings_transactions(goal_id);`,
		`CREATE INDEX IF NOT EXISTS idx_savings_transactions_date ON savings_transactions(date);`,
	}

	// Add savings_balance column to existing accounts table if it doesn't exist
	alterAccountsTable := `
	ALTER TABLE accounts 
	ADD COLUMN IF NOT EXISTS savings_balance DECIMAL(20,2) DEFAULT 0.00;`

	// Update existing columns to support larger amounts
	alterExistingColumns := []string{
		`ALTER TABLE accounts ALTER COLUMN balance TYPE DECIMAL(20,2);`,
		`ALTER TABLE accounts ALTER COLUMN savings_balance TYPE DECIMAL(20,2);`,
		`ALTER TABLE accounts ALTER COLUMN allowance_income TYPE DECIMAL(20,2);`,
		`ALTER TABLE transactions ALTER COLUMN amount TYPE DECIMAL(20,2);`,
		`ALTER TABLE savings_goals ALTER COLUMN target_amount TYPE DECIMAL(20,2);`,
		`ALTER TABLE savings_goals ALTER COLUMN current_amount TYPE DECIMAL(20,2);`,
		`ALTER TABLE savings_transactions ALTER COLUMN amount TYPE DECIMAL(20,2);`,
	}

	// Execute table creation
	if _, err := db.Exec(usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	if _, err := db.Exec(accountsTable); err != nil {
		return fmt.Errorf("failed to create accounts table: %v", err)
	}

	// Add savings_balance column if it doesn't exist
	if _, err := db.Exec(alterAccountsTable); err != nil {
		return fmt.Errorf("failed to alter accounts table: %v", err)
	}

	if _, err := db.Exec(transactionsTable); err != nil {
		return fmt.Errorf("failed to create transactions table: %v", err)
	}

	if _, err := db.Exec(savingsGoalsTable); err != nil {
		return fmt.Errorf("failed to create savings_goals table: %v", err)
	}

	if _, err := db.Exec(savingsTransactionsTable); err != nil {
		return fmt.Errorf("failed to create savings_transactions table: %v", err)
	}

	// Update existing columns to support larger amounts (ignore errors for non-existent tables)
	for _, alterCmd := range alterExistingColumns {
		db.Exec(alterCmd) // Ignore errors as tables might not exist yet
	}

	// Create indexes
	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %v", err)
		}
	}

	return nil
}
