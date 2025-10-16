# Cross-Service Authentication Guide

## Overview

This document explains how JWT token authentication works across the **hub-user-service** (microservice) and the **HubInvestments** (monolith), ensuring seamless token validation between both services.

---

## 🎯 Authentication Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    AUTHENTICATION FLOW                          │
└─────────────────────────────────────────────────────────────────┘

1. User Login (via Microservice)
   ┌──────────┐                    ┌──────────────────┐
   │  Client  │  ──gRPC Login──>  │  hub-user-service│
   └──────────┘                    └──────────────────┘
                                            │
                                            │ Generate JWT Token
                                            │ (HS256 + MY_JWT_SECRET)
                                            ↓
                                   ┌─────────────────┐
                                   │  JWT Token      │
                                   │  eyJhbGciOi...  │
                                   └─────────────────┘
                                            │
                                            ↓
   ┌──────────┐  <── Token Response ───    │
   │  Client  │                             │
   └──────────┘                             │
        │                                   │
        │ Store token                       │
        └───────────────────────────────────┘

2. API Request (to Monolith)
   ┌──────────┐                    ┌──────────────────┐
   │  Client  │  ──HTTP Request─>  │  Monolith        │
   │          │  Authorization:    │  (HubInvestments)│
   │          │  Bearer <token>    │                  │
   └──────────┘                    └──────────────────┘
                                            │
                                            │ Validate JWT Token
                                            │ (HS256 + MY_JWT_SECRET)
                                            ↓
                                   ┌─────────────────┐
                                   │  Token Valid?   │
                                   │  Yes: Allow     │
                                   │  No: 401        │
                                   └─────────────────┘
```

---

## 🔐 Token Synchronization Mechanism

### Key Principle: **Shared Secret**

Both services use the **same JWT secret** to ensure tokens created by one service can be validated by the other.

| Component | Microservice | Monolith | Must Match |
|-----------|--------------|----------|------------|
| **JWT Secret** | `MY_JWT_SECRET` env var | `MY_JWT_SECRET` env var | ✅ YES |
| **Algorithm** | HS256 | HS256 | ✅ YES |
| **Expiration** | 10 minutes | 10 minutes | ✅ YES |
| **Claims** | `username`, `userId`, `exp` | `username`, `userId`, `exp` | ✅ YES |
| **Token Format** | `Bearer <token>` | `Bearer <token>` | ✅ YES |

### How It Works

1. **Token Creation (Microservice)**
   ```go
   // hub-user-service/internal/auth/token/token_service.go
   secret := os.Getenv("MY_JWT_SECRET") // e.g., "my-super-secret-key"
   
   token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
       "username": "user@example.com",
       "userId":   "user123",
       "exp":      time.Now().Add(10 * time.Minute).Unix(),
   })
   
   tokenString, _ := token.SignedString([]byte(secret))
   // Result: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ..."
   ```

2. **Token Validation (Monolith)**
   ```go
   // HubInvestments/internal/auth/token/token_service.go
   secret := os.Getenv("MY_JWT_SECRET") // SAME SECRET: "my-super-secret-key"
   
   token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
       return []byte(secret), nil
   })
   
   // If secrets match → token is valid
   // If secrets differ → signature verification fails
   ```

### Why This Works

- **HMAC-SHA256 (HS256)**: Symmetric encryption algorithm
- **Same Secret** = Same signature = Cross-validation works
- **No Database Sync Needed**: Stateless JWT validation
- **No Session Storage**: Each service validates independently

---

## ⚙️ Configuration Requirements

### Environment Variables (CRITICAL)

Both services **MUST** have identical JWT configuration:

```bash
# Microservice (hub-user-service)
export MY_JWT_SECRET="your-production-secret-key-min-32-chars"

# Monolith (HubInvestments)
export MY_JWT_SECRET="your-production-secret-key-min-32-chars"
```

⚠️ **CRITICAL**: If the secrets don't match, tokens created by one service will be **rejected** by the other!

### Configuration Files

**Microservice: `hub-user-service/config.env`**
```env
MY_JWT_SECRET=your-production-secret-key-min-32-chars
ENVIRONMENT=production
HTTP_PORT=8080
GRPC_PORT=50051
```

**Monolith: `HubInvestments/.env`** (or existing config)
```env
MY_JWT_SECRET=your-production-secret-key-min-32-chars
```

---

## 🧪 Verification & Testing

### Integration Tests

The project includes comprehensive integration tests to verify cross-service authentication:

```bash
cd hub-user-service
go test ./test/integration/... -v
```

**Test Coverage:**
1. ✅ Microservice creates token → Monolith validates it
2. ✅ Token expiration respected by both services
3. ✅ Secret mismatch detection
4. ✅ Configuration verification
5. ✅ Real-world authentication flow
6. ✅ Environment variable synchronization

### Manual Verification Steps

#### Step 1: Start Microservice
```bash
cd hub-user-service
export MY_JWT_SECRET="test-secret-for-verification"
go run cmd/server/main.go
```

#### Step 2: Login via gRPC (get token)
```bash
# Using grpcurl or your gRPC client
grpcurl -plaintext -d '{
  "email": "test@example.com",
  "password": "password123"
}' localhost:50051 AuthService/Login

# Response will include JWT token:
# {
#   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#   "userInfo": { "userId": "user123", "email": "test@example.com" }
# }
```

#### Step 3: Verify Token in Monolith
```bash
# Use the token from Step 2 in a request to the monolith
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
     http://localhost:8080/api/protected-endpoint

# Expected: 200 OK (token validated successfully)
```

#### Step 4: Wait for Token Expiration
```bash
# Wait 10+ minutes, then retry Step 3
# Expected: 401 Unauthorized (token expired)
```

---

## 🔒 Security Considerations

### JWT Secret Management

**Production Requirements:**
- ✅ Use a strong, random secret (minimum 32 characters)
- ✅ Never commit secrets to version control
- ✅ Use environment variables or secret management systems (AWS Secrets Manager, HashiCorp Vault)
- ✅ Rotate secrets periodically
- ✅ Use different secrets for dev/staging/production

**Example: Generating a Strong Secret**
```bash
# Generate a secure random secret (64 characters)
openssl rand -base64 48

# Example output:
# kJ8N2mP9qR3tV5xY7zB1cD4fG6hK8mN0pQ2sU4wX6yA9bC1eF3gH5jL7nM9pR1t
```

### Token Expiration

- **Current**: 10 minutes
- **Rationale**: Balance between security (shorter = better) and UX (longer = less re-auth)
- **Recommendation**: Consider implementing **refresh tokens** for longer sessions

### Token Validation

Both services validate:
1. ✅ Signature (HMAC-SHA256 with shared secret)
2. ✅ Expiration time (`exp` claim)
3. ✅ Token structure (required claims: `username`, `userId`, `exp`)

---

## 🚨 Troubleshooting

### Problem: "Invalid token signature"

**Cause**: Microservice and monolith are using **different JWT secrets**

**Solution**:
1. Check `MY_JWT_SECRET` in both services:
   ```bash
   # Microservice
   echo $MY_JWT_SECRET
   
   # Monolith
   echo $MY_JWT_SECRET
   ```
2. Ensure both values are **identical**
3. Restart both services after updating

### Problem: "Token has expired"

**Cause**: Token was issued more than 10 minutes ago

**Solution**:
1. This is **expected behavior** for security
2. Client should request a new token (re-login)
3. Consider implementing refresh tokens for better UX

### Problem: "Token is missing"

**Cause**: Authorization header not sent or malformed

**Solution**:
1. Verify header format: `Authorization: Bearer <token>`
2. Ensure "Bearer " prefix is included (with space)
3. Check for typos in header name

### Problem: "Monolith doesn't accept microservice tokens"

**Cause**: Multiple possible issues

**Checklist**:
- [ ] Both services use the same `MY_JWT_SECRET`
- [ ] Both services use HS256 algorithm
- [ ] Token format is `Bearer <token>` (with space after Bearer)
- [ ] Token hasn't expired (check system clocks)
- [ ] Monolith's TokenService hasn't been modified

**Debug**:
```bash
# Decode JWT token to inspect claims (without validating)
echo "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." | \
  cut -d. -f2 | \
  base64 -d | \
  jq
```

---

## 📋 Deployment Checklist

Before deploying to production:

### Pre-Deployment
- [ ] Generate a strong production JWT secret (min 32 chars)
- [ ] Store secret in secure location (e.g., AWS Secrets Manager)
- [ ] Document secret rotation procedure
- [ ] Set up monitoring for authentication failures

### Microservice Deployment
- [ ] Set `MY_JWT_SECRET` environment variable
- [ ] Verify `ENVIRONMENT=production`
- [ ] Test gRPC Login endpoint
- [ ] Test gRPC ValidateToken endpoint
- [ ] Verify logs show correct secret length (masked)

### Monolith Configuration
- [ ] Verify monolith uses the **SAME** `MY_JWT_SECRET`
- [ ] Test that monolith validates microservice tokens
- [ ] Test token expiration behavior
- [ ] Verify existing authentication still works

### Post-Deployment Verification
- [ ] Login via microservice
- [ ] Use token to access monolith endpoints
- [ ] Verify token expiration after 10 minutes
- [ ] Monitor authentication error rates
- [ ] Check logs for JWT validation errors

### Rollback Plan
- [ ] Document how to switch back to monolith-only auth
- [ ] Keep monolith's internal auth logic intact (initially)
- [ ] Test rollback procedure in staging

---

## 🎯 Success Criteria

Your cross-service authentication is working correctly when:

1. ✅ User can login via microservice gRPC endpoint
2. ✅ Received JWT token is valid for 10 minutes
3. ✅ Monolith accepts and validates tokens from microservice
4. ✅ Both services reject expired tokens (after 10 minutes)
5. ✅ Both services reject tokens with wrong signature (wrong secret)
6. ✅ Token contains correct claims: `username`, `userId`, `exp`
7. ✅ All integration tests pass (`go test ./test/integration/... -v`)

---

## 📚 Additional Resources

- **JWT Specification**: https://datatracker.ietf.org/doc/html/rfc7519
- **HS256 Algorithm**: https://datatracker.ietf.org/doc/html/rfc7518#section-3.2
- **JWT Best Practices**: https://tools.ietf.org/html/rfc8725
- **Integration Tests**: `test/integration/cross_service_auth_test.go`
- **Token Service Implementation**: `internal/auth/token/token_service.go`

---

## 🆘 Support

If you encounter issues with cross-service authentication:

1. Run integration tests: `go test ./test/integration/... -v`
2. Check logs for JWT validation errors
3. Verify `MY_JWT_SECRET` matches in both services
4. Review this documentation's troubleshooting section
5. Contact the platform team for assistance

---

**Last Updated**: Phase 10.1, Step 3.5 (Cross-Service Authentication Verification)

