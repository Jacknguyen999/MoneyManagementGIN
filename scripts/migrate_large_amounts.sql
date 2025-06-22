-- Migration script to support larger amounts (up to 18 digits)
-- Run this on existing databases to fix the savings goal amount limit

-- Update accounts table
ALTER TABLE accounts ALTER COLUMN balance TYPE DECIMAL(20,2);
ALTER TABLE accounts ALTER COLUMN savings_balance TYPE DECIMAL(20,2);
ALTER TABLE accounts ALTER COLUMN allowance_income TYPE DECIMAL(20,2);

-- Update transactions table
ALTER TABLE transactions ALTER COLUMN amount TYPE DECIMAL(20,2);

-- Update savings_goals table
ALTER TABLE savings_goals ALTER COLUMN target_amount TYPE DECIMAL(20,2);
ALTER TABLE savings_goals ALTER COLUMN current_amount TYPE DECIMAL(20,2);

-- Update savings_transactions table  
ALTER TABLE savings_transactions ALTER COLUMN amount TYPE DECIMAL(20,2);

-- Verify the changes
\d+ accounts;
\d+ transactions;
\d+ savings_goals;
\d+ savings_transactions; 