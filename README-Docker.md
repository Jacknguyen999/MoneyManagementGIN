# Docker Setup for Money Manager Backend

This directory contains Docker configuration for the Money Manager backend API.

## Files Overview

- `Dockerfile` - Multi-stage build for Go application
- `docker-compose.yml` - Orchestrates backend, database, and Redis services
- `.dockerignore` - Excludes unnecessary files from Docker context
- `.env.example` - Environment variables template
- `scripts/wait-for-db.sh` - Database readiness script
- `Makefile` - Cross-platform commands for Docker operations

## Quick Start

1. **Setup Environment**
   ```bash
   cd backend
   cp .env.example .env  # Linux/Mac
   copy .env.example .env  # Windows
   ```
   Edit `.env` with your settings.

2. **Start Services**
   ```bash
   make up
   ```

3. **Stop Services**
   ```bash
   make down
   ```

4. **View All Commands**
   ```bash
   make help
   ```

## Available Commands

### Essential Commands
```bash
make up           # Start all services
make down         # Stop all services
make restart      # Restart all services
make dev          # Start in development mode (with logs)
```

### Development Commands
```bash
make logs         # View logs (all services)
make logs-api     # View backend API logs only
make health       # Check service health
make build        # Build the backend image
```

### Database Commands
```bash
make db-shell     # Access database shell
make db-backup    # Create database backup
```

### Utility Commands
```bash
make shell        # Access backend container shell
make clean        # Remove containers and volumes
make test         # Run tests
```

## Services

### Backend API
- **Port**: 8080
- **Health Check**: http://localhost:8080/health
- **Environment**: Configurable via .env file

### PostgreSQL Database
- **Port**: 5432
- **Default DB**: money_manager
- **Default User**: admin
- **Default Password**: password123

### Redis (Optional)
- **Port**: 6379
- **Purpose**: Session storage, caching

## Environment Variables

Key variables in `.env`:

```env
# Database
DB_HOST=db                    # Use 'db' for Docker, 'localhost' for local
DB_PORT=5432
DB_USER=admin
DB_PASSWORD=password123
DB_NAME=money_manager

# Application
PORT=8080
JWT_SECRET=your-secret-key
GIN_MODE=debug               # 'release' for production

# CORS
CORS_ORIGIN=http://localhost:3000
```

## Development Workflow

1. **Make code changes** in your editor
2. **Rebuild container**:
   ```bash
   docker-compose up --build backend
   ```
3. **View logs** to debug:
   ```bash
   docker-compose logs -f backend
   ```

## Database Access

### Connect to PostgreSQL
```bash
docker-compose exec db psql -U admin -d money_manager
```

### Database Backup
```bash
docker-compose exec db pg_dump -U admin money_manager > backup.sql
```

### Database Restore
```bash
docker-compose exec -T db psql -U admin money_manager < backup.sql
```

## Troubleshooting

### Container won't start
1. Check Docker is running
2. Verify .env file exists
3. Check port conflicts: `netstat -an | findstr :8080`

### Database connection issues
1. Wait for database to be ready (health check)
2. Verify database credentials in .env
3. Check logs: `docker-compose logs db`

### Build failures
1. Clear Docker cache: `docker system prune`
2. Rebuild from scratch: `docker-compose build --no-cache`

## Production Deployment

For production:

1. **Update .env**:
   ```env
   GIN_MODE=release
   JWT_SECRET=strong-random-secret
   DB_PASSWORD=strong-password
   ```

2. **Use production compose**:
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

3. **Enable SSL/TLS** with reverse proxy (nginx/traefik)

## Next Steps

- Set up CI/CD pipeline
- Add monitoring (Prometheus/Grafana)
- Implement backup automation
- Configure log aggregation 