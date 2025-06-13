# Footy Tipping Project - Tasklist & Development Rules

## Project Overview
A football tipping application built with Go backend, React frontend, PostgreSQL database, and Docker Compose orchestration. Uses event sourcing architecture with CQRS pattern for scalable match and prediction management.

## Current Architecture

### Backend (Go)
- **Framework**: Gorilla Mux for HTTP routing
- **Architecture**: Event Sourcing with CQRS
- **Database**: PostgreSQL with read models (`matches_view`, `predictions_view`)
- **Port**: 8080
- **Main Entry**: `backend/cmd/api/main.go`

### Frontend (React + TypeScript)
- **Framework**: React 18 with TypeScript
- **UI Library**: Material-UI (MUI)
- **State Management**: TanStack Query (React Query)
- **Routing**: React Router
- **Port**: 3000

### Database Schema
- **Event Store**: `events` table for event sourcing
- **Read Models**:
  - `matches_view` (id, home_team, away_team, match_date, competition, status, home_goals, away_goals)
  - `predictions_view` (id, user_id, match_id, home_goals, away_goals, created_at, points)

## Current Features ✅

### Completed
- [x] Event sourcing architecture with domain events
- [x] Match management (create, update, list)
- [x] Prediction system with score parsing ("2-1" format)
- [x] Home page with upcoming matches section
- [x] Matches page with prediction submission
- [x] Profile and Leaderboard page stubs
- [x] Docker Compose setup with hot reload
- [x] Comprehensive Makefile for development
- [x] 20 sample matches imported from major competitions
- [x] Read model projections for performance
- [x] CORS middleware for API access
- [x] Consistent camelCase JSON serialization
- [x] MIT License

### API Endpoints
- `GET /api/matches` - List all matches
- `GET /api/matches/upcoming` - List upcoming matches (limited to 5)
- `GET /api/matches/{id}` - Get specific match
- `POST /api/matches` - Create new match
- `PUT /api/matches/{id}/score` - Update match score
- `POST /api/predictions` - Create prediction
- `GET /api/matches/{matchId}/predictions/{userId}` - Get user prediction for match

## Development Rules & Guidelines

### 1. Event Sourcing Principles
- **NEVER** modify events once stored
- **ALWAYS** create new events for state changes
- **ENSURE** read models are updated via event handlers
- **USE** domain events: `MatchCreated`, `MatchScoreUpdated`, `PredictionMade`

### 2. API Design Standards
- **JSON Format**: Use camelCase for all API responses
- **Error Handling**: Return appropriate HTTP status codes
- **Route Order**: Specific routes BEFORE parameterized routes (e.g., `/upcoming` before `/{id}`)
- **CORS**: Ensure all endpoints support CORS for frontend access

### 3. Database Operations
- **Read Models**: Use for queries, never for commands
- **Event Store**: Single source of truth for all state changes
- **Migrations**: Handle schema changes carefully
- **Transactions**: Use for atomic operations

### 4. Frontend Standards
- **TypeScript**: Strict typing for all components
- **Material-UI**: Consistent component usage
- **React Query**: For all API interactions
- **Error Handling**: User-friendly error messages
- **Loading States**: Show loading indicators for async operations

### 5. Docker & Development
- **Hot Reload**: Both frontend and backend support live reloading
- **Make Commands**: Use Makefile for all common operations
- **Environment**: Development setup via Docker Compose
- **Ports**: Backend (8080), Frontend (3000), PostgreSQL (5432)

## Pending Tasks & Future Features

### High Priority 🔴
- [ ] User authentication and authorization system
- [ ] Real user management (replace hardcoded `user123`)
- [ ] Points calculation system for predictions
- [ ] Leaderboard functionality with real data
- [ ] Match status management (LIVE, FINISHED)
- [ ] Real-time score updates

### Medium Priority 🟡
- [ ] Email notifications for match results
- [ ] Competition management and filtering
- [ ] User profile management with statistics
- [ ] Prediction history and analytics
- [ ] Mobile responsive design improvements
- [ ] Admin panel for match management

### Low Priority 🟢
- [ ] Social features (comments, sharing)
- [ ] Multiple competition support
- [ ] Advanced statistics and charts
- [ ] Export functionality for predictions
- [ ] API rate limiting
- [ ] Caching layer (Redis)

## Technical Debt & Improvements

### Code Quality
- [ ] Add comprehensive unit tests for all handlers
- [ ] Integration tests for API endpoints
- [ ] Frontend component testing
- [ ] Error boundary implementation
- [ ] Logging improvements with structured logging

### Performance
- [ ] Database indexing optimization
- [ ] API response caching
- [ ] Frontend bundle optimization
- [ ] Image optimization and CDN
- [ ] Database connection pooling

### Security
- [ ] Input validation and sanitization
- [ ] SQL injection prevention
- [ ] XSS protection
- [ ] Rate limiting implementation
- [ ] Security headers

## Development Workflow

### Making Changes
1. **Backend Changes**: Modify Go code, rebuild with `make rebuild`
2. **Frontend Changes**: Modify React code, hot reload automatic
3. **Database Changes**: Update migrations, run `make db-reset` if needed
4. **Testing**: Use `make test` for backend, manual testing for frontend

### Common Commands
```bash
make start          # Start all services
make stop           # Stop all services
make restart        # Restart all services
make rebuild        # Rebuild and start
make db-shell       # Access database
make import-fixtures # Import sample matches
make health         # Check service health
```

### File Structure Rules
- **Backend**: Follow Go project layout standards
- **Frontend**: Component-based architecture in `src/`
- **Shared**: Types and interfaces should be consistent
- **Config**: Environment-specific settings in Docker Compose

## Data Models

### Domain Objects
```go
type Match struct {
    ID          string      `json:"id"`
    HomeTeam    string      `json:"homeTeam"`
    AwayTeam    string      `json:"awayTeam"`
    Date        time.Time   `json:"date"`
    Competition string      `json:"competition"`
    Status      MatchStatus `json:"status"`
    Score       *Score      `json:"score"`
}

type Prediction struct {
    ID        string    `json:"id"`
    UserID    string    `json:"userId"`
    MatchID   string    `json:"matchId"`
    HomeGoals int       `json:"homeGoals"`
    AwayGoals int       `json:"awayGoals"`
    CreatedAt time.Time `json:"createdAt"`
    Points    int       `json:"points"`
}
```

### Event Types
- `MatchCreated`: New match added to system
- `MatchScoreUpdated`: Match score changed
- `MatchStatusChanged`: Match status updated
- `PredictionMade`: User made a prediction

## Testing Strategy

### Backend Testing
- Unit tests for domain logic
- Integration tests for API endpoints
- Event handler testing
- Repository testing with test database

### Frontend Testing
- Component unit tests with React Testing Library
- Integration tests for user flows
- E2E tests with Cypress (future)

## Deployment Considerations

### Production Readiness
- [ ] Environment variable management
- [ ] Database backup strategy
- [ ] Monitoring and alerting
- [ ] Load balancing setup
- [ ] SSL/TLS configuration
- [ ] CI/CD pipeline

### Scaling Considerations
- Event sourcing supports horizontal scaling
- Read models can be replicated
- Frontend can be served via CDN
- Database read replicas for queries

---

## Important Notes for Future Development

1. **Event Sourcing**: Never bypass the event store for state changes
2. **API Consistency**: Maintain camelCase JSON throughout
3. **Route Order**: Always place specific routes before parameterized ones
4. **Error Handling**: Provide meaningful error messages to users
5. **Testing**: Write tests for new features before implementation
6. **Documentation**: Update this file when adding new features or changing architecture

This document should be updated whenever significant changes are made to the project architecture or when new features are implemented.
