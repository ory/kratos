#!/bin/sh
set -e

echo "🚀 Starting Ory Kratos setup..."

# Construct DSN from Railway PostgreSQL variables
if [ -n "$POSTGRES_HOST" ]; then
    export DSN="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DATABASE}?sslmode=require"
    echo "✅ Using Railway PostgreSQL"
elif [ -n "$DATABASE_URL" ]; then
    export DSN="$DATABASE_URL"
    echo "✅ Using DATABASE_URL"
else
    echo "❌ ERROR: No database configuration found"
    exit 1
fi

echo "📊 Database connection configured"

# Wait for database to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
MAX_RETRIES=30
RETRY_COUNT=0
RETRY_INTERVAL=2

until kratos migrate sql status -e 2>/dev/null || [ $RETRY_COUNT -eq $MAX_RETRIES ]; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "   Attempt $RETRY_COUNT/$MAX_RETRIES - Database not ready yet..."
    sleep $RETRY_INTERVAL
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "❌ ERROR: Database did not become ready after $MAX_RETRIES attempts"
    exit 1
fi

echo "✅ Database is ready!"

# Show migration status before running migrations
echo ""
echo "📋 Current migration status:"
kratos migrate sql status -e || true

# Run migrations
echo ""
echo "🔄 Running database migrations..."
kratos migrate sql up -e --yes

# Show migration status after running migrations
echo ""
echo "📋 Migration status after update:"
kratos migrate sql status -e

echo ""
echo "🎉 Migrations completed successfully!"
echo "🚀 Starting Kratos server..."
echo ""

# Start Kratos server
exec kratos serve all -c /etc/config/kratos/kratos.yml
