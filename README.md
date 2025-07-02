# Reminder App

A modern Go-based reminder application with full-text search capabilities powered by Typesense.

## Features

- ✅ Create, read, update, and delete reminders
- 🔍 Full-text search using Typesense
- 🐳 Docker containerization
- 🗄️ PostgreSQL database
- 🌐 Modern web interface with HTMX
- 📱 Responsive design

## Tech Stack

- **Backend**: Go 1.23 with Echo framework
- **Database**: PostgreSQL 15
- **Search Engine**: Typesense
- **Frontend**: HTML, CSS, HTMX
- **Containerization**: Docker & Docker Compose

## Project Structure

```
├── cmd/
│   └── server/          # Application entrypoint
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   └── search/          # Search functionality
├── web/
│   └── static/          # Static web assets
├── docker-compose.yml   # Docker services
├── Dockerfile          # Container build
└── go.mod              # Go dependencies
```

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone <repository-url>
cd goapi
```

2. Start all services:
```bash
docker-compose up -d
```

3. Access the application:
- **Web App**: http://localhost:3000
- **API**: http://localhost:3000/api
- **PgAdmin**: http://localhost:5050 (admin@example.com / admin)

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Start PostgreSQL and Typesense:
```bash
docker-compose up -d db typesense
```

3. Set environment variables:
```bash
export DB_HOST=localhost
export DB_USER=goapi
export DB_PASSWORD=password
export DB_NAME=goapi
export TYPESENSE_HOST=localhost
export TYPESENSE_API_KEY=xyz
```

4. Run the application:
```bash
go run cmd/server/main.go
```

## API Endpoints

### Reminders
- `POST /api/reminders` - Create a new reminder
- `GET /api/reminders` - Get all reminders
- `GET /api/reminders/:id` - Get a specific reminder
- `PUT /api/reminders/:id` - Update a reminder
- `DELETE /api/reminders/:id` - Delete a reminder

### Search
- `GET /api/search?query=<term>&limit=<number>` - Search reminders

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `3000` | Server port |
| `DB_HOST` | `localhost` | Database host |
| `DB_USER` | `goapi` | Database user |
| `DB_PASSWORD` | `password` | Database password |
| `DB_NAME` | `goapi` | Database name |
| `DB_PORT` | `5432` | Database port |
| `TYPESENSE_HOST` | `localhost` | Typesense host |
| `TYPESENSE_PORT` | `8108` | Typesense port |
| `TYPESENSE_API_KEY` | `xyz` | Typesense API key |

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o reminder-api ./cmd/server
```

### Code Formatting
```bash
go fmt ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.