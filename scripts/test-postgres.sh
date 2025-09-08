#!/bin/bash

# Test script for demonstrating PostgreSQL functionality
set -e

echo "🐘 Testing PostgreSQL Integration"
echo "================================="

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if PostgreSQL is running
check_postgres() {
    echo -e "${BLUE}Checking PostgreSQL availability...${NC}"
    if command -v psql &> /dev/null; then
        if pg_isready -h localhost -p 5432 &> /dev/null; then
            echo -e "${GREEN}✅ PostgreSQL is running${NC}"
            return 0
        else
            echo -e "${RED}❌ PostgreSQL is not running${NC}"
            return 1
        fi
    else
        echo -e "${RED}❌ PostgreSQL client not found${NC}"
        return 1
    fi
}

# Start PostgreSQL with Docker if not running
start_postgres_docker() {
    echo -e "${BLUE}Starting PostgreSQL with Docker...${NC}"
    docker run -d --name postgres-test \
        -e POSTGRES_DB=example_db \
        -e POSTGRES_USER=postgres \
        -e POSTGRES_PASSWORD=password \
        -p 5432:5432 postgres:15 || true
    
    echo "Waiting for PostgreSQL to start..."
    sleep 5
    
    # Wait for PostgreSQL to be ready
    for i in {1..30}; do
        if docker exec postgres-test pg_isready -U postgres &> /dev/null; then
            echo -e "${GREEN}✅ PostgreSQL is ready${NC}"
            return 0
        fi
        sleep 1
    done
    
    echo -e "${RED}❌ PostgreSQL failed to start${NC}"
    return 1
}

# Run tests
run_tests() {
    echo -e "${BLUE}1. Running unit tests...${NC}"
    go test ./internal/repository/ -v
    
    echo -e "\n${BLUE}2. Running database package tests...${NC}"
    go test ./pkg/database/ -v
    
    echo -e "\n${BLUE}3. Running integration tests with PostgreSQL...${NC}"
    export TEST_DATABASE_URL="postgres://postgres:password@localhost:5432/example_db?sslmode=disable"
    export DB_HOST=localhost
    export DB_PORT=5432
    export DB_NAME=example_db
    export DB_USER=postgres
    export DB_PASSWORD=password
    
    go test ./internal/repository/ -v -run TestPostgreSQLIntegration
    go test ./pkg/database/ -v -run TestPostgreSQLIntegration
}

# Build and test applications
test_applications() {
    echo -e "${BLUE}4. Building applications...${NC}"
    go build -o bin/server-postgres cmd/server/main.go
    go build -o bin/consumer-postgres cmd/consumer/main.go
    
    echo -e "${BLUE}5. Testing PostgreSQL configuration...${NC}"
    
    # Set PostgreSQL environment
    export DB_TYPE=postgres
    export DB_HOST=localhost
    export DB_PORT=5432
    export DB_NAME=example_db
    export DB_USERNAME=postgres
    export DB_PASSWORD=password
    export MQ_ENABLE_MOCK=true
    export LOG_LEVEL=info
    
    # Test server startup (background)
    echo "Starting server with PostgreSQL..."
    timeout 10s ./bin/server-postgres &
    SERVER_PID=$!
    sleep 3
    
    # Test if server started successfully
    if kill -0 $SERVER_PID 2>/dev/null; then
        echo -e "${GREEN}✅ Server started successfully with PostgreSQL${NC}"
        kill $SERVER_PID 2>/dev/null || true
    else
        echo -e "${RED}❌ Server failed to start${NC}"
    fi
    
    # Test consumer startup (background)
    echo "Starting consumer with PostgreSQL..."
    timeout 10s ./bin/consumer-postgres &
    CONSUMER_PID=$!
    sleep 3
    
    # Test if consumer started successfully
    if kill -0 $CONSUMER_PID 2>/dev/null; then
        echo -e "${GREEN}✅ Consumer started successfully with PostgreSQL${NC}"
        kill $CONSUMER_PID 2>/dev/null || true
    else
        echo -e "${RED}❌ Consumer failed to start${NC}"
    fi
}

# Cleanup
cleanup() {
    echo -e "${BLUE}Cleaning up...${NC}"
    
    # Stop and remove test containers
    docker stop postgres-test 2>/dev/null || true
    docker rm postgres-test 2>/dev/null || true
    
    # Remove test binaries
    rm -f bin/server-postgres bin/consumer-postgres
    
    echo -e "${GREEN}✅ Cleanup complete${NC}"
}

# Main execution
main() {
    # Check if PostgreSQL is available
    if ! check_postgres; then
        echo -e "${YELLOW}PostgreSQL not available, starting with Docker...${NC}"
        if ! start_postgres_docker; then
            echo -e "${RED}Failed to start PostgreSQL. Exiting.${NC}"
            exit 1
        fi
    fi
    
    # Run tests
    if run_tests; then
        echo -e "${GREEN}✅ All PostgreSQL tests passed!${NC}"
    else
        echo -e "${RED}❌ Some tests failed${NC}"
        cleanup
        exit 1
    fi
    
    # Test applications
    if test_applications; then
        echo -e "${GREEN}✅ Application tests passed!${NC}"
    else
        echo -e "${RED}❌ Application tests failed${NC}"
        cleanup
        exit 1
    fi
    
    # Success message
    echo -e "\n${GREEN}🎉 PostgreSQL Integration Test Complete!${NC}"
    echo -e "\n${YELLOW}📋 PostgreSQL Features Tested:${NC}"
    echo "================================="
    echo "✅ Database connection and health checks"
    echo "✅ GORM ORM integration"
    echo "✅ Repository pattern implementation"
    echo "✅ CRUD operations (Create, Read, Update, Delete)"
    echo "✅ Database migrations"
    echo "✅ Connection pooling"
    echo "✅ Error handling and fallback to in-memory"
    echo "✅ Transaction support"
    echo "✅ Advanced queries (search, age filtering, statistics)"
    echo "✅ Server and consumer application integration"
    echo ""
    echo -e "${YELLOW}📊 Performance Benchmarks:${NC}"
    go test ./internal/repository/ -bench=BenchmarkPostgreSQLRepository -benchmem
    
    # Cleanup
    cleanup
}

# Handle script interruption
trap cleanup EXIT INT TERM

# Run main function
main "$@"
