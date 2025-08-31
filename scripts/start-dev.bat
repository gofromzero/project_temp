@echo off

echo Starting Multi-Tenant Admin Development Environment...

REM Start Docker services first
echo Starting Docker services (MySQL and Redis)...
cd backend
docker-compose up -d
cd ..

REM Wait for services to be ready
echo Waiting for database services to start...
timeout /t 10 >nul

REM Start backend
echo Starting backend server on port 8000...
start "Backend Server" cmd /k "cd backend && go run cmd/main.go"

REM Wait a moment for backend to start
timeout /t 3 >nul

REM Start frontend
echo Starting frontend dev server on port 3000...
start "Frontend Server" cmd /k "cd frontend/apps/admin-web && npm run dev"

echo.
echo Development environment started!
echo Backend: http://localhost:8000
echo Frontend: http://localhost:3000
echo Health check: http://localhost:8000/health
echo.
echo Press any key to stop all services...
pause >nul

REM Stop Docker services
echo Stopping Docker services...
cd backend
docker-compose down
cd ..

echo Development environment stopped.