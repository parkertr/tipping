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
#### Current Endpoints
- `GET /api/matches` - List all matches
- `GET /api/matches/upcoming` - List upcoming matches (limited to 5)
- `GET /api/matches/{id}` - Get specific match
- `POST /api/matches` - Create new match
- `PUT /api/matches/{id}/score` - Update match score
- `POST /api/predictions` - Create prediction
- `GET /api/matches/{matchId}/predictions/{userId}` - Get user prediction for match

#### New Authentication Endpoints (To Be Added)
- `GET /api/auth/google` - Initiate Google OAuth flow
- `GET /api/auth/google/callback` - Handle Google OAuth callback
- `GET /api/auth/me` - Get current authenticated user profile
- `PUT /api/auth/me` - Update user profile
- `POST /api/auth/refresh` - Refresh JWT token
- `POST /api/auth/logout` - Logout user (invalidate token)
- `GET /api/users/me/stats` - Get user statistics
- `GET /api/users/me/predictions` - Get user's prediction history

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

## Google Authentication Integration - Detailed Task Breakdown

### Phase 1: Backend Infrastructure 🔴
#### 1.1 User Domain & Events
- [ ] Create `User` domain model with Google profile data
  - [ ] `internal/domain/user.go` - User struct with GoogleID, Email, Name, Picture
  - [ ] User methods: `NewUser()`, `UpdateProfile()`, `IsActive()`
- [ ] Define user-related events
  - [ ] `UserRegistered` - New user signs up via Google
  - [ ] `UserProfileUpdated` - User profile information changed
  - [ ] `UserDeactivated` - User account disabled
- [ ] Add user events to `pkg/events/events.go`

#### 1.2 User Repository & Read Model
- [ ] Create users database table and read model
  - [ ] Migration: `users_view` table (id, google_id, email, name, picture_url, created_at, updated_at, is_active)
  - [ ] Add migration to `backend/migrations/`
- [ ] Implement user repository
  - [ ] `internal/infrastructure/repository/user_repository.go` interface
  - [ ] `internal/infrastructure/repository/postgres/user_repository.go` implementation
  - [ ] Methods: `Create()`, `GetByGoogleID()`, `GetByID()`, `Update()`, `List()`
- [ ] Create user event handler
  - [ ] `internal/infrastructure/eventhandlers/user_handler.go`
  - [ ] Process `UserRegistered`, `UserProfileUpdated` events

#### 1.3 Google OAuth Setup
- [ ] Google Cloud Console configuration
  - [ ] Create new Google Cloud Project or use existing
  - [ ] Enable Google+ API and Google OAuth2 API
  - [ ] Create OAuth 2.0 credentials (Client ID and Secret)
  - [ ] Configure authorized redirect URIs for development and production
- [ ] Environment configuration
  - [ ] Add Google OAuth credentials to `docker-compose.yml`
  - [ ] Environment variables: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`

#### 1.4 Authentication Middleware & JWT
- [ ] Install Go dependencies
  - [ ] `go get golang.org/x/oauth2`
  - [ ] `go get golang.org/x/oauth2/google`
  - [ ] `go get github.com/golang-jwt/jwt/v5`
- [ ] Implement JWT token management
  - [ ] `internal/auth/jwt.go` - Generate, validate, refresh JWT tokens
  - [ ] Token claims structure with user ID, email, expiration
  - [ ] JWT secret management via environment variables
- [ ] Create authentication middleware
  - [ ] `internal/infrastructure/api/middleware/auth.go`
  - [ ] Extract and validate JWT from Authorization header
  - [ ] Add user context to request for protected endpoints

#### 1.5 Authentication Handlers
- [ ] Google OAuth flow handlers
  - [ ] `internal/infrastructure/api/handlers/auth.go`
  - [ ] `GET /api/auth/google` - Redirect to Google OAuth
  - [ ] `GET /api/auth/google/callback` - Handle OAuth callback
  - [ ] `POST /api/auth/refresh` - Refresh JWT token
  - [ ] `POST /api/auth/logout` - Invalidate token (optional)
- [ ] User profile handlers
  - [ ] `GET /api/auth/me` - Get current user profile
  - [ ] `PUT /api/auth/me` - Update user profile

### Phase 2: Frontend Integration 🔴
#### 2.1 Authentication Context & State
- [ ] Install React dependencies
  - [ ] `npm install @google-cloud/local-auth google-auth-library`
  - [ ] `npm install js-cookie @types/js-cookie` (for token storage)
- [ ] Create authentication context
  - [ ] `frontend/src/contexts/AuthContext.tsx`
  - [ ] Manage user state, login/logout functions, token refresh
  - [ ] Provide authentication state to entire app
- [ ] Token management utilities
  - [ ] `frontend/src/utils/auth.ts`
  - [ ] Store/retrieve JWT tokens in httpOnly cookies or localStorage
  - [ ] Automatic token refresh logic
  - [ ] Axios interceptors for adding auth headers

#### 2.2 Login/Logout Components
- [ ] Google Sign-In button component
  - [ ] `frontend/src/components/GoogleSignIn.tsx`
  - [ ] Use Google Sign-In JavaScript library
  - [ ] Handle OAuth redirect flow
  - [ ] Error handling for failed authentication
- [ ] User profile dropdown/menu
  - [ ] `frontend/src/components/UserMenu.tsx`
  - [ ] Display user name, email, profile picture
  - [ ] Logout functionality
  - [ ] Link to profile page

#### 2.3 Protected Routes & Navigation
- [ ] Route protection wrapper
  - [ ] `frontend/src/components/ProtectedRoute.tsx`
  - [ ] Redirect unauthenticated users to login
  - [ ] Show loading state while checking authentication
- [ ] Update navigation bar
  - [ ] Show login button for unauthenticated users
  - [ ] Show user menu for authenticated users
  - [ ] Update `frontend/src/components/Navigation.tsx`

#### 2.4 Update Existing Components
- [ ] Remove hardcoded `user123` from all components
  - [ ] `frontend/src/pages/Matches.tsx` - Use real user ID from auth context
  - [ ] `frontend/src/pages/Profile.tsx` - Display real user data
  - [ ] `frontend/src/pages/Home.tsx` - Personalize welcome message
- [ ] Add authentication checks to API calls
  - [ ] Update all axios calls to include authentication headers
  - [ ] Handle 401 responses with automatic logout

### Phase 3: Database & API Updates 🟡
#### 3.1 Update Existing Models
- [ ] Modify prediction system for real users
  - [ ] Update `Prediction` domain model to use real user IDs
  - [ ] Migrate existing predictions to use proper user references
  - [ ] Update prediction handlers to validate user ownership
- [ ] Update API endpoints for user context
  - [ ] Protect prediction endpoints with authentication middleware
  - [ ] Filter predictions by authenticated user
  - [ ] Validate user permissions for prediction operations

#### 3.2 User Profile Management
- [ ] Enhanced user profile API
  - [ ] `GET /api/users/me/stats` - User statistics (predictions, points, rank)
  - [ ] `GET /api/users/me/predictions` - User's prediction history
  - [ ] `PUT /api/users/me/preferences` - User preferences/settings
- [ ] Admin user management (future)
  - [ ] `GET /api/admin/users` - List all users (admin only)
  - [ ] `PUT /api/admin/users/{id}/status` - Activate/deactivate users

### Phase 4: Security & Production Readiness 🟡
#### 4.1 Security Enhancements
- [ ] Input validation and sanitization
  - [ ] Validate all user inputs in auth handlers
  - [ ] Sanitize user profile data from Google
  - [ ] Rate limiting for authentication endpoints
- [ ] Security headers and CORS
  - [ ] Update CORS configuration for production domains
  - [ ] Add security headers middleware
  - [ ] Implement CSRF protection for state-changing operations

#### 4.2 Error Handling & Logging
- [ ] Comprehensive error handling
  - [ ] User-friendly error messages for auth failures
  - [ ] Proper HTTP status codes for different auth scenarios
  - [ ] Graceful handling of Google API failures
- [ ] Authentication logging
  - [ ] Log successful/failed login attempts
  - [ ] Log token refresh operations
  - [ ] Security event logging (suspicious activity)

#### 4.3 Testing
- [ ] Backend authentication tests
  - [ ] Unit tests for JWT token generation/validation
  - [ ] Integration tests for auth endpoints
  - [ ] Mock Google OAuth for testing
- [ ] Frontend authentication tests
  - [ ] Test authentication context state management
  - [ ] Test protected route behavior
  - [ ] Test login/logout flows

### Phase 5: Data Migration & Deployment 🟢
#### 5.1 Data Migration Strategy
- [ ] Migrate existing data
  - [ ] Create migration script for existing predictions
  - [ ] Handle orphaned predictions without valid users
  - [ ] Backup strategy before migration
- [ ] User onboarding flow
  - [ ] Welcome message for new users
  - [ ] Tutorial or guide for first-time users
  - [ ] Default user preferences setup

#### 5.2 Environment Configuration
- [ ] Production environment setup
  - [ ] Configure Google OAuth for production domain
  - [ ] Set up proper JWT secrets and rotation
  - [ ] Configure secure cookie settings
- [ ] Monitoring and alerting
  - [ ] Monitor authentication success/failure rates
  - [ ] Alert on suspicious authentication patterns
  - [ ] Track user registration and activity metrics

## Pending Tasks & Future Features

### High Priority 🔴
- [ ] **Google Authentication Integration** (See detailed breakdown below)
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

// New User model for Google Auth integration
type User struct {
    ID         string    `json:"id"`
    GoogleID   string    `json:"googleId"`
    Email      string    `json:"email"`
    Name       string    `json:"name"`
    PictureURL string    `json:"pictureUrl"`
    CreatedAt  time.Time `json:"createdAt"`
    UpdatedAt  time.Time `json:"updatedAt"`
    IsActive   bool      `json:"isActive"`
}
```

### Event Types
- `MatchCreated`: New match added to system
- `MatchScoreUpdated`: Match score changed
- `MatchStatusChanged`: Match status updated
- `PredictionMade`: User made a prediction
- `UserRegistered`: New user registered via Google OAuth
- `UserProfileUpdated`: User profile information updated
- `UserDeactivated`: User account deactivated

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

## Google Authentication Flow

### Backend OAuth Flow
1. User clicks "Sign in with Google" → Frontend redirects to `/api/auth/google`
2. Backend redirects to Google OAuth with client credentials
3. User authenticates with Google → Google redirects to `/api/auth/google/callback`
4. Backend exchanges authorization code for user profile data
5. Backend creates/updates user via event sourcing (`UserRegistered`/`UserProfileUpdated`)
6. Backend generates JWT token and returns to frontend
7. Frontend stores JWT and redirects to dashboard

### Frontend Authentication State
1. App loads → Check for existing JWT token
2. If valid token exists → Set user as authenticated
3. If no/invalid token → Show login screen
4. On successful login → Store JWT and update auth context
5. On API calls → Include JWT in Authorization header
6. On 401 responses → Clear token and redirect to login

### Token Management
- **JWT Structure**: `{ userId, email, name, exp, iat }`
- **Storage**: httpOnly cookies (preferred) or localStorage
- **Refresh**: Automatic refresh before expiration
- **Expiration**: 24 hours (configurable)

## Important Notes for Future Development

1. **Event Sourcing**: Never bypass the event store for state changes
2. **API Consistency**: Maintain camelCase JSON throughout
3. **Route Order**: Always place specific routes before parameterized ones
4. **Authentication**: All user-related operations must validate JWT tokens
5. **User Context**: Always use authenticated user ID, never trust client-provided user IDs
6. **Error Handling**: Provide meaningful error messages to users
7. **Testing**: Write tests for new features before implementation
8. **Google OAuth**: Keep client credentials secure and rotate regularly
9. **Documentation**: Update this file when adding new features or changing architecture

This document should be updated whenever significant changes are made to the project architecture or when new features are implemented.
