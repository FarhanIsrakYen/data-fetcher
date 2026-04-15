# Data Fetcher API

A Go-based REST API for fetching and managing trading data, performance metrics, and execution information.

## Overview

Data Fetcher API provides endpoints for retrieving trading instrument data, performance analytics, execution records, and reliability scores. Built with Go and Gin framework, it uses TimescaleDB for time-series data storage and Redis for caching.

## Tech Stack

- **Language**: Go 1.20
- **Framework**: Gin
- **Database**: TimescaleDB (PostgreSQL)
- **Cache**: Redis
- **Message Queue**: RabbitMQ with MQTT
- **Error Tracking**: Sentry
- **Cloud Storage**: GCP

## Prerequisites

- Go 1.20 or higher
- Docker & Docker Compose
- TimescaleDB
- Redis
- RabbitMQ (optional, for message queue features)

## Installation

### 1. Clone the Repository

```bash
git clone <repository-url>
cd data-fetcher-api
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Environment Configuration

Copy the example environment file and configure it:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
APP_ENV=local

# Google Drive API
GOOGLE_CLIENT_ID=your_client_id
GOOGLE_PRIVATE_KEY="your_private_key"
GOOGLE_PRIVATE_KEY_ID=your_key_id

# Database
TIMESCALEDB_PASSWORD=your_password
TIMESCALEDB_CONNECTION_STRING=postgres://user:password@host:5432/database

# RabbitMQ
RABBITMQ_PASSWORD=your_password

# Sentry
SENTRY_DSN=your_sentry_dsn
```

### 4. Configure Parameters

Edit configuration files in `config/parameters/`:

- `all.yaml` - Base configuration
- `dev.yaml` - Development environment

Example `config/parameters/all.yaml`:

```yaml
parameters:
  app.namespace: 'Data'
  config.path: 'config/google/config.json'
  redis.address: 'localhost:6379'
  gin.server.port: 2000
  df.session.prefix: 'df:session:'
  df.product.api.uri: 'http://localhost:3000'
```

### 5. Database Setup

Run migrations to set up the database schema:

```bash
make migrate
```

Or manually:

```bash
go run src/Command/General/MigrateCommand.go
```

## Running the Application

### Development Mode

Using Make:

```bash
make up
```

Or run directly:

```bash
go run main.go
```

The API will start on `http://localhost:2000` (or the port specified in your configuration).

### Using Docker

```bash
docker-compose up -d
```

## API Endpoints

### Health Check

- **GET** `/status` - Check API health status

### Instruments

- **GET** `/guest/instruments/:instrumentName/price` - Get real-time instrument price
  - Parameters: `instrumentName` (path parameter)

### Executions

- **GET** `/guest/executions` - Get execution records
  - Query params: `limit`, `orderBy`, `order`, `page`, filters
  
- **GET** `/guest/executions/historical` - Get historical execution records
  - Query params: `limit`, `orderBy`, `order`, `page`, filters

- **GET** `/guest/templates/:id/executions` - Get executions for a specific template
  - Parameters: `id` (template ID)
  - Query params: `limit`, `orderBy`, `order`, `page`

- **GET** `/guest/templates/:id/executions/historical` - Get historical executions for a template
  - Parameters: `id` (template ID)
  - Query params: `limit`, `orderBy`, `order`, `page`

### Performances

- **GET** `/guest/performances` - Get performance metrics
  - Query params: `limit`, `orderBy`, `order`, `page`, filters

- **GET** `/guest/performances/historical` - Get historical performance metrics
  - Query params: `limit`, `orderBy`, `order`, `page`, filters

- **GET** `/guest/templates/:id/performances` - Get performance for a specific template
  - Parameters: `id` (template ID)
  - Query params: `limit`, `orderBy`, `order`, `page`

- **GET** `/guest/templates/:id/performances/historical` - Get historical performance for a template
  - Parameters: `id` (template ID)
  - Query params: `limit`, `orderBy`, `order`, `page`

### Profit Estimation

- **GET** `/guest/templates/eligible/estimation` - Get eligible strategies for profit estimation
  - Query params: template filters

- **GET** `/guest/templates/:id/profit/estimation` - Get estimated profit for a template
  - Parameters: `id` (template ID)

### Reliability Score

- **GET** `/guest/templates/:id/score` - Get reliability score for a template
  - Parameters: `id` (template ID)

### Testing

- **POST** `/tests/set-user` - Set test user (development only)

## Response Format

All endpoints return JSON responses in the following format:

### Success Response

```json
{
  "status": true,
  "message": "Success message",
  "data": { ... },
  "pagination": {
    "currentPage": 1,
    "totalPages": 10,
    "totalItems": 100,
    "itemsPerPage": 10
  }
}
```

### Error Response

```json
{
  "status": false,
  "message": "Error message",
  "data": []
}
```

## Query Parameters

Common query parameters across endpoints:

- `limit` - Number of records per page (default: 10)
- `page` - Page number (default: 1)
- `orderBy` - Field to order by (default: "time")
- `order` - Sort order: "ASC" or "DESC" (default: "DESC")
- Additional filters based on endpoint

## Database Tables

The API uses the following TimescaleDB hypertables:

- `df_data_data` - Raw data storage
- `df_data_executions` - Execution records
- `df_data_historical_executions` - Historical execution records
- `df_data_performances` - Performance metrics
- `df_data_historical_performances` - Historical performance metrics
- `df_data_instrument_logs` - Instrument price logs
- `df_data_reliables` - Reliability scores

## Development

### Available Make Commands

```bash
make up          # Start the application
make down        # Stop Docker containers
make migrate     # Run database migrations
make package     # Install Go dependencies
make cache       # Update cache
make ssh         # SSH into Docker container
make test        # Run tests
make lint        # Run linter
```

### Project Structure

```
data-fetcher-api/
├── config/
│   ├── packages/      # Database, cache, sentry configs
│   ├── parameters/    # YAML configuration files
│   └── routes/        # API route definitions
├── src/
│   ├── Api/          # External API integrations
│   ├── Command/      # CLI commands & cronjobs
│   ├── Controllers/  # HTTP request handlers
│   ├── Helper/       # Utility functions
│   ├── Lib/          # Libraries (Redis, MQTT)
│   ├── Migrations/   # Database migrations
│   ├── Model/        # Data models
│   ├── Mq/          # Message queue handlers
│   └── Repository/   # Data access layer
├── main.go          # Application entry point
├── go.mod           # Go module definition
└── Makefile         # Build automation
```

## CORS Configuration

The API is configured for local development with the following allowed origins:

- `http://127.0.0.1`
- `http://localhost`
- `http://127.0.0.1:3000`
- `http://localhost:3000`
- `http://127.0.0.1:8080`
- `http://localhost:8080`

To modify CORS settings, edit `config/routes/routes.go`.

## Troubleshooting

### Port Already in Use

If port 2000 is already in use, change the port in `config/parameters/all.yaml`:

```yaml
gin.server.port: 3000
```

### Database Connection Issues

1. Ensure TimescaleDB is running
2. Check connection string in `.env`
3. Verify database credentials
4. Run migrations: `make migrate`

### Redis Connection Issues

1. Ensure Redis is running on the configured address
2. Check `redis.address` in `config/parameters/all.yaml`
3. Test connection: `redis-cli ping`

## License

[Your License Here]

## Contributing

[Contributing Guidelines Here]

## Support

For issues and questions, please open an issue in the repository.
