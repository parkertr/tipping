#!/bin/bash

# Import fixtures script for the footy tipping application
# Usage: ./scripts/import-fixtures.sh [options]

set -e

# Default values
DATABASE_URL=${DATABASE_URL:-"postgres://postgres:postgres@localhost:5432/footy_tipping?sslmode=disable"}
FIXTURES_FILE=${FIXTURES_FILE:-"fixtures/matches.json"}
DRY_RUN=${DRY_RUN:-false}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Import match fixtures into the footy tipping database.

OPTIONS:
    -d, --database URL      Database connection URL (default: from DATABASE_URL env var)
    -f, --fixtures FILE     Path to fixtures JSON file (default: fixtures/matches.json)
    --dry-run              Show what would be imported without actually importing
    -h, --help             Show this help message

EXAMPLES:
    # Import fixtures using default settings
    $0

    # Dry run to see what would be imported
    $0 --dry-run

    # Import from custom fixtures file
    $0 -f custom-fixtures.json

    # Import to custom database
    $0 -d "postgres://user:pass@localhost:5432/mydb?sslmode=disable"

ENVIRONMENT VARIABLES:
    DATABASE_URL           Database connection URL
    FIXTURES_FILE          Path to fixtures JSON file
    DRY_RUN               Set to 'true' for dry run mode

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--database)
            DATABASE_URL="$2"
            shift 2
            ;;
        -f|--fixtures)
            FIXTURES_FILE="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Check if we're in the backend directory
if [[ ! -f "go.mod" ]]; then
    print_error "This script must be run from the backend directory"
    exit 1
fi

# Check if fixtures file exists
if [[ ! -f "$FIXTURES_FILE" ]]; then
    print_error "Fixtures file not found: $FIXTURES_FILE"
    exit 1
fi

# Print configuration
print_info "Configuration:"
echo "  Database URL: $DATABASE_URL"
echo "  Fixtures file: $FIXTURES_FILE"
echo "  Dry run: $DRY_RUN"
echo

# Build the import tool if it doesn't exist or is outdated
IMPORT_TOOL="./bin/import-fixtures"
if [[ ! -f "$IMPORT_TOOL" ]] || [[ "cmd/import-fixtures/main.go" -nt "$IMPORT_TOOL" ]]; then
    print_info "Building import tool..."
    mkdir -p bin
    go build -o "$IMPORT_TOOL" ./cmd/import-fixtures
    print_success "Import tool built successfully"
fi

# Prepare arguments
ARGS="-db=$DATABASE_URL -fixtures=$FIXTURES_FILE"
if [[ "$DRY_RUN" == "true" ]]; then
    ARGS="$ARGS -dry-run"
fi

# Run the import tool
print_info "Starting fixtures import..."
if $IMPORT_TOOL $ARGS; then
    print_success "Fixtures import completed successfully!"
else
    print_error "Fixtures import failed!"
    exit 1
fi
