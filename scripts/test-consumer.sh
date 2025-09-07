#!/bin/bash

# Test script for demonstrating consumer functionality
set -e

echo "🚀 Testing Message Queue Consumer"
echo "================================="

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}1. Running unit tests...${NC}"
go test ./internal/transport/mq/ -v

echo -e "\n${BLUE}2. Running consumer application tests...${NC}"
go test ./cmd/consumer/ -v

echo -e "\n${BLUE}3. Running benchmarks...${NC}"
go test ./internal/transport/mq/ -bench=. -benchmem

echo -e "\n${BLUE}4. Building applications...${NC}"
go build -o bin/server cmd/server/main.go
go build -o bin/consumer cmd/consumer/main.go

echo -e "\n${GREEN}✅ All tests passed successfully!${NC}"
echo -e "\n${YELLOW}📋 Manual Testing Instructions:${NC}"
echo "================================="
echo "1. Start RabbitMQ (optional, will use mock by default):"
echo "   docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management"
echo ""
echo "2. Start the consumer (in terminal 1):"
echo "   ./bin/consumer"
echo ""
echo "3. Start the server (in terminal 2):"
echo "   ./bin/server"
echo ""
echo "4. Test API endpoints (in terminal 3):"
echo "   # Create an example"
echo "   curl -X POST http://localhost:8080/api/v1/examples \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"name\":\"Test User\",\"email\":\"test@example.com\",\"age\":25}'"
echo ""
echo "   # Update an example"
echo "   curl -X PUT http://localhost:8080/api/v1/examples/[ID] \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"name\":\"Updated User\",\"email\":\"updated@example.com\",\"age\":30}'"
echo ""
echo "   # Delete an example"
echo "   curl -X DELETE http://localhost:8080/api/v1/examples/[ID]"
echo ""
echo "5. Watch consumer logs for event processing messages"
echo ""
echo -e "${GREEN}🎉 Consumer testing complete!${NC}"
