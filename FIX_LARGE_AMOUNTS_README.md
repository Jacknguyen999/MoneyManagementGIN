# Fix for Large Amount Bug - Savings Goals

## Problem

Users getting `500 Internal Server Error` when creating savings goals with large amounts (e.g.,
2,000,000,000).

## Root Cause

Database columns were using `DECIMAL(10,2)` which only supports amounts up to 99,999,999.99 (8
digits + 2 decimals).

## Solution

Updated all amount columns to `DECIMAL(20,2)` which supports amounts up to
999,999,999,999,999,999.99 (18 digits + 2 decimals).

## What Was Fixed

### Database Schema Changes:

- `accounts.balance`: DECIMAL(10,2) → DECIMAL(20,2)
- `accounts.savings_balance`: DECIMAL(10,2) → DECIMAL(20,2)
- `accounts.allowance_income`: DECIMAL(10,2) → DECIMAL(20,2)
- `transactions.amount`: DECIMAL(10,2) → DECIMAL(20,2)
- `savings_goals.target_amount`: DECIMAL(10,2) → DECIMAL(20,2)
- `savings_goals.current_amount`: DECIMAL(10,2) → DECIMAL(20,2)
- `savings_transactions.amount`: DECIMAL(10,2) → DECIMAL(20,2)

### New Deployment Instructions:

1. For new deployments: Just run `make up` - the migration will happen automatically
2. For existing deployments: Run `make db-migrate-large-amounts` to update column types

### Testing the Fix:

```bash
# 1. Apply the fix to existing database
make db-migrate-large-amounts

# 2. Test with large amount
curl -X POST http://localhost:8080/api/savings/goals \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Laptop",
    "target_amount": 2000000000,
    "deadline": "2025-06-27",
    "description": ""
  }'
```

## Supported Amount Range

- **Before**: $0.01 to $99,999,999.99
- **After**: $0.01 to $999,999,999,999,999,999.99

No frontend validation was added - users can enter any large number as requested.

## Files Modified:

- `backend/database/connection.go` - Updated schema and added migration
- `backend/scripts/migrate_large_amounts.sql` - Migration script
- `backend/Makefile` - Added migration command
- `backend/docker-compose.yml` - Added scripts volume mount
