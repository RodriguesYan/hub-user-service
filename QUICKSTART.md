# User Management Service - Quick Start Guide

## 🚀 Get Started in 5 Minutes

### Prerequisites
- Docker and Docker Compose installed
- Ports 8081, 50052, and 5433 available

### Step 1: Clone and Navigate
```bash
cd /Users/yanrodrigues/Documents/HubInvestmentsProject/hub-user-service
```

### Step 2: Start the Service
```bash
docker-compose up -d
```

### Step 3: Verify it's Running
```bash
# Check health
curl http://localhost:8081/health

# Expected output:
# {"healthy":true,"version":"1.0.0"}
```

### Step 4: Try It Out

#### Register a new user
```bash
curl -X POST http://localhost:8081/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!",
    "firstName": "John",
    "lastName": "Doe"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!"
  }'
```

**Save the token from the response!**

#### Get Profile (use the token from login)
```bash
curl -X GET http://localhost:8081/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 🎯 What Just Happened?

You now have a fully functional User Management microservice running with:

- ✅ **User authentication** - JWT-based login system
- ✅ **User registration** - Create new accounts with validation
- ✅ **Profile management** - Retrieve user information
- ✅ **Token validation** - Verify JWT tokens
- ✅ **Security features** - Password hashing, account locking, email validation
- ✅ **Database** - PostgreSQL with proper schema and indexes
- ✅ **Dual APIs** - Both HTTP REST and gRPC interfaces

## 🔌 Use from Your Monolith

### Add to your monolith's code:

```go
// Initialize the client (do this once at startup)
userClient, err := grpc.NewUserServiceClient("localhost:50052")
if err != nil {
    log.Fatal(err)
}
defer userClient.Close()

// Use it to validate tokens
ctx := context.Background()
valid, userID, email, err := userClient.ValidateToken(ctx, token)
if err != nil {
    return fmt.Errorf("validation failed: %w", err)
}

if !valid {
    return errors.New("invalid token")
}

// Now you have authenticated user info
fmt.Printf("User %s (%s) authenticated\n", email, userID)
```

## 📊 Service Architecture

```
┌─────────────────────────────────────────────┐
│         External Clients                    │
│    (Web Apps, Mobile Apps, Monolith)       │
└──────────────┬──────────────────────────────┘
               │
               ├─── HTTP REST (Port 8081)
               │    • POST /login
               │    • POST /register
               │    • GET /profile
               │    • POST /validate-token
               │
               └─── gRPC (Port 50052)
                    • Login()
                    • RegisterUser()
                    • GetUserProfile()
                    • ValidateToken()
                           │
                           ▼
               ┌───────────────────────┐
               │   User Service        │
               │   (hub-user-service)  │
               └───────────┬───────────┘
                           │
                           ▼
               ┌───────────────────────┐
               │   PostgreSQL          │
               │   (hub_users_db)      │
               │   Port: 5433          │
               └───────────────────────┘
```

## 🛠️ Useful Commands

```bash
# View logs
docker-compose logs -f user-service

# Stop the service
docker-compose down

# Restart the service
docker-compose restart user-service

# View database
docker-compose exec user-db psql -U postgres -d hub_users_db

# Check running containers
docker-compose ps
```

## 📚 Next Steps

1. **Read the full documentation**: Check out `README.md`
2. **Deploy to production**: See `DEPLOYMENT.md`
3. **Understand the migration**: Read `MIGRATION_SUMMARY.md`
4. **Integrate with monolith**: Update your authentication flow

## 🐛 Troubleshooting

### Port already in use?
```bash
# Change ports in docker-compose.yml
ports:
  - "8082:8080"  # Change 8081 to 8082
  - "50053:50051"  # Change 50052 to 50053
```

### Can't connect to database?
```bash
# Wait for database to be ready
docker-compose logs user-db

# Look for: "database system is ready to accept connections"
```

### Service won't start?
```bash
# Check logs
docker-compose logs user-service

# Rebuild from scratch
docker-compose down -v
docker-compose up --build
```

## ✅ Verification Checklist

- [ ] Service starts without errors
- [ ] Health endpoint returns 200 OK
- [ ] Can register a new user
- [ ] Can login with registered user
- [ ] Receive JWT token
- [ ] Can access protected profile endpoint with token
- [ ] Token validation works
- [ ] Database contains user data

## 🎓 Learning Resources

- **Clean Architecture**: See the code structure in `internal/`
- **Domain-Driven Design**: Check `internal/domain/model/`
- **Use Cases Pattern**: Look at `internal/application/usecase/`
- **gRPC**: Explore `proto/user_service.proto`
- **Dependency Injection**: See `shared/container/container.go`

## 🤝 Support

If you encounter issues:
1. Check the logs: `docker-compose logs`
2. Verify ports are available: `lsof -i :8081 -i :50052 -i :5433`
3. Review the documentation in `README.md`
4. Check the deployment guide in `DEPLOYMENT.md`

---

**You're all set!** The User Management Service is now running and ready to handle authentication for your Hub Investments platform. 🎉
