# Footy Tipping App

A modern footy tipping application built with React and Golang, featuring event sourcing for robust data management.

## Features

- User authentication and management
- Match predictions and tipping
- Real-time updates
- Historical performance tracking
- Leaderboards and statistics
- **Match fixtures import system** - Import upcoming matches from JSON files

## Tech Stack

### Frontend
- React with TypeScript
- Material-UI for modern UI components
- React Query for data fetching
- React Router for navigation

### Backend
- Golang
- Event sourcing architecture
- PostgreSQL for event store
- RESTful API
- JWT authentication

## Getting Started

### Prerequisites
- Node.js 18+
- Go 1.21+
- Docker and Docker Compose
- PostgreSQL 15+

### Development Setup

1. Clone the repository
2. Start the development environment:
   ```bash
   docker-compose up -d
   ```
3. Install frontend dependencies:
   ```bash
   cd frontend
   npm install
   ```
4. Start the frontend development server:
   ```bash
   npm run dev
   ```
5. The backend will be available at `http://localhost:8080`
6. The frontend will be available at `http://localhost:3000`

### Import Sample Fixtures

To populate the database with sample upcoming matches:

```bash
# View available fixtures (dry run)
cd backend
./scripts/import-fixtures.sh --dry-run

# Import fixtures to database
./scripts/import-fixtures.sh
```

## Project Structure

```
.
├── frontend/           # React frontend application
├── backend/           # Golang backend application
│   ├── cmd/          # Application entry points
│   ├── internal/     # Private application code
│   ├── pkg/          # Public library code
│   ├── fixtures/     # Match fixtures and import tools
│   ├── scripts/      # Utility scripts
│   └── migrations/   # Database migrations
└── docker-compose.yml # Docker services configuration
```

## Match Fixtures System

The application includes a comprehensive system for importing match fixtures from JSON files.

### Quick Start with Fixtures

1. **View available fixtures**:
   ```bash
   cd backend
   ./scripts/import-fixtures.sh --dry-run
   ```

2. **Import fixtures to database**:
   ```bash
   ./scripts/import-fixtures.sh
   ```

### Fixture File Format

Fixtures are defined in JSON format:

```json
[
  {
    "id": "match_001",
    "homeTeam": "Manchester United",
    "awayTeam": "Liverpool",
    "date": "2025-06-20T15:00:00Z",
    "competition": "Premier League"
  }
]
```

### Available Competitions

The default fixtures include matches from:
- **Premier League** - English top division
- **La Liga** - Spanish top division
- **Bundesliga** - German top division
- **Serie A** - Italian top division
- **Ligue 1** - French top division
- **International competitions** - Nations League, Copa America, etc.

### Import Script Options

```bash
# Basic usage
./scripts/import-fixtures.sh

# Custom fixtures file
./scripts/import-fixtures.sh -f custom-fixtures.json

# Custom database
./scripts/import-fixtures.sh -d "postgres://postgres:postgres@host:port/db?sslmode=disable"

# Dry run mode
./scripts/import-fixtures.sh --dry-run

# Help
./scripts/import-fixtures.sh --help
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | Database connection URL | `postgres://postgres:postgres@localhost:5432/footy_tipping?sslmode=disable` |
| `FIXTURES_FILE` | Path to fixtures file | `fixtures/matches.json` |
| `DRY_RUN` | Enable dry run mode | `false` |

### Creating Custom Fixtures

1. **Create a new JSON file**:
   ```bash
   cp backend/fixtures/matches.json backend/fixtures/my-fixtures.json
   ```

2. **Edit the fixture data** with your matches

3. **Import the fixtures**:
   ```bash
   ./scripts/import-fixtures.sh -f fixtures/my-fixtures.json
   ```

### How It Works

The fixtures system integrates with the event sourcing architecture:

1. **Read fixtures** - Parse JSON fixture files
2. **Check existing** - Query event store for existing matches
3. **Create events** - Generate `MatchCreated` events for new matches
4. **Save events** - Store events in the event store
5. **Event processing** - Application handlers update read models

Each fixture creates a `MatchCreated` event that's processed by the application to update the `matches_view` table, making matches available through the API and frontend.

### Troubleshooting Fixtures

**Database connection failed**
- Ensure PostgreSQL is running via `docker-compose up -d`
- Check DATABASE_URL matches docker-compose configuration

**Fixtures file not found**
- Ensure you're in the `backend` directory
- Verify the fixtures file path is correct

**Invalid JSON format**
- Validate JSON syntax
- Check date format is ISO 8601 (UTC): `YYYY-MM-DDTHH:MM:SSZ`

**Match already exists**
- Normal behavior - duplicates are automatically skipped
- Use different match IDs for similar matches

## API Endpoints

### Matches
- `GET /api/matches` - List all matches
- `GET /api/matches/{id}` - Get specific match
- `POST /api/matches` - Create new match
- `PUT /api/matches/{id}/score` - Update match score

### Predictions
- `POST /api/predictions` - Create prediction
- `GET /api/users/{userId}/predictions` - Get user predictions
- `GET /api/matches/{matchId}/predictions` - Get match predictions

## Development

### Running Tests

```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd frontend
npm test
```

### Database Migrations

Migrations are automatically applied when starting with docker-compose. For manual migration:

```bash
# Apply migrations
cd backend
go run ./cmd/migrate up

# Create new migration
go run ./cmd/migrate create migration_name
```

## License

MIT

## Google OAuth Setup

To enable Google authentication, you need to:

1. Create a Google Cloud Project:
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Create a new project or select an existing one
   - Enable the Google+ API and Google OAuth2 API
   - Create OAuth 2.0 credentials (Client ID and Secret)
   - Configure authorized redirect URIs:
     - Development: `http://localhost:8080/api/auth/google/callback`
     - Production: `https://your-domain.com/api/auth/google/callback`

2. Set up environment variables:
   ```bash
   # Google OAuth Configuration
   GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-client-secret
   GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback

   # JWT Configuration
   JWT_SECRET=your-secret-key
   ```

3. Start the application:
   ```bash
   make up
   ```
