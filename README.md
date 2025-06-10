# Footy Tipping App

A modern footy tipping application built with React and Golang, featuring event sourcing for robust data management.

## Features

- User authentication and management
- Match predictions and tipping
- Real-time updates
- Historical performance tracking
- Leaderboards and statistics

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

## Project Structure

```
.
├── frontend/           # React frontend application
├── backend/           # Golang backend application
│   ├── cmd/          # Application entry points
│   ├── internal/     # Private application code
│   ├── pkg/          # Public library code
│   └── events/       # Event definitions and handlers
└── docker/           # Docker configuration files
```

## License

MIT
