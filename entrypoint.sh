#!/bin/sh
#
# Entrypoint script for doit API
# This script runs migrations before starting the application
#

set -e  # Exit on error

echo "🚀 Starting doit API..."
echo "================================================"

# Function to check if database is ready
wait_for_db() {
    echo "⏳ Waiting for database to be ready..."
    
    # Install postgresql-client if not present (for pg_isready)
    if ! command -v pg_isready > /dev/null 2>&1; then
        echo "📦 Installing postgresql-client..."
        apk add --no-cache postgresql-client > /dev/null 2>&1
    fi
    
    # Wait for PostgreSQL
    until pg_isready -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} > /dev/null 2>&1; do
        echo "   Database is unavailable - sleeping"
        sleep 2
    done
    
    echo "✅ Database is ready!"
    SSL_MODE=${DB_SSL_MODE:-disable}
    DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${SSL_MODE}"
    echo "DB_URL: $DB_URL"
    echo "SSL_MODE: $SSL_MODE"
}

# Function to run migrations
run_migrations() {
    echo ""
    echo "📦 Running database migrations..."
    
    # Check if migrate binary exists
    if [ -f "./migrate" ]; then
        MIGRATE_CMD="./migrate"
    elif command -v migrate > /dev/null 2>&1; then
        MIGRATE_CMD="migrate"
    else
        echo "⚠️  Warning: migrate binary not found. Skipping migrations."
        echo "   Install migrate or add it to your Docker image."
        return 0
    fi
    # Construct database URL (default to sslmode=disable if not set)
    SSL_MODE=${DB_SSL_MODE:-disable}
    DB_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${SSL_MODE}"
    echo "DB_URL: $DB_URL"
    # Run migrations
    echo "   Migration path: /app/internal/data/migrations"
    if $MIGRATE_CMD -path=/app/internal/data/migrations -database="$DB_URL" up; then
        echo "✅ Migrations completed successfully!"
    else
        echo "❌ Migration failed!"
        exit 1
    fi
}

# Main execution
main() {
    # Wait for database
    wait_for_db
    
    # Run migrations (if SKIP_MIGRATIONS is not set)
    if [ -z "$SKIP_MIGRATIONS" ]; then
        run_migrations
    else
        echo "⏭️  Skipping migrations (SKIP_MIGRATIONS is set)"
    fi
    
    echo ""
    echo "================================================"
    echo "🎉 Starting application..."
    echo ""
    
    # Execute the main application
    exec ./app "$@"
}

# Run main function
main "$@"

