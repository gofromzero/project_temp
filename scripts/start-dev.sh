#!/bin/bash

# Start development servers for multi-tenant admin system

echo "Starting Multi-Tenant Admin Development Environment..."

# Start backend in background
echo "Starting backend server on port 8000..."
cd backend
go run cmd/main.go &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 2

# Start frontend in background
echo "Starting frontend dev server on port 3000..."
cd ../frontend/apps/admin-web
npm run dev &
FRONTEND_PID=$!

echo "Backend PID: $BACKEND_PID"
echo "Frontend PID: $FRONTEND_PID"
echo "Backend running on: http://localhost:8000"
echo "Frontend running on: http://localhost:3000"
echo "Health check: http://localhost:8000/health"

# Wait for user input to stop servers
echo "Press any key to stop servers..."
read -n 1

# Kill background processes
kill $BACKEND_PID
kill $FRONTEND_PID

echo "Development servers stopped."