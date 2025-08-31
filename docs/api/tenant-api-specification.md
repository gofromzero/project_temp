# Tenant Management API Specification

## Overview

The Tenant Management API provides comprehensive functionality for managing tenants in a multi-tenant system. This RESTful API supports creating, reading, updating, and deleting tenant information with full data isolation and security controls.

**Base URL:** `/v1/tenants`

**Authentication:** Bearer JWT Token (required for all endpoints)

## API Endpoints

### 1. List Tenants

Retrieve a paginated list of tenants with optional filtering.

**Endpoint:** `GET /tenants`

**Query Parameters:**
- `page` (integer, optional): Page number (default: 1, min: 1)
- `limit` (integer, optional): Items per page (default: 10, min: 1, max: 100)
- `name` (string, optional): Filter by tenant name (partial match)
- `code` (string, optional): Filter by tenant code (partial match)
- `status` (string, optional): Filter by status (`active`, `suspended`, `disabled`)

**Example Request:**
```bash
curl -X GET "https://api.example.com/v1/tenants?page=1&limit=20&status=active" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Accept: application/json"
```

**Response Format:**
```json
{
  "success": true,
  "data": [
    {
      "id": "tenant-uuid-1",
      "name": "Tenant One",
      "code": "TENANT1",
      "status": "active",
      "config": {
        "maxUsers": 100,
        "features": ["basic", "analytics"],
        "theme": "default",
        "domain": "tenant1.example.com"
      },
      "adminUserId": "admin-uuid-1",
      "createdAt": "2025-08-31T10:00:00Z",
      "updatedAt": "2025-08-31T10:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50,
    "pages": 3
  },
  "requestId": "req-20250831-100000",
  "timestamp": "2025-08-31T10:00:00Z"
}
```

### 2. Get Tenant Details

Retrieve complete information for a specific tenant.

**Endpoint:** `GET /tenants/{tenantId}`

**Path Parameters:**
- `tenantId` (string, required): Unique tenant identifier

**Example Request:**
```bash
curl -X GET "https://api.example.com/v1/tenants/tenant-uuid-1" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Accept: application/json"
```

**Response Format:**
```json
{
  "success": true,
  "data": {
    "id": "tenant-uuid-1",
    "name": "Tenant One",
    "code": "TENANT1",
    "status": "active",
    "config": {
      "maxUsers": 100,
      "features": ["basic", "analytics"],
      "theme": "default",
      "domain": "tenant1.example.com"
    },
    "adminUserId": "admin-uuid-1",
    "createdAt": "2025-08-31T10:00:00Z",
    "updatedAt": "2025-08-31T10:00:00Z"
  },
  "requestId": "req-20250831-100001",
  "timestamp": "2025-08-31T10:00:00Z"
}
```

### 3. Create Tenant

Create a new tenant with optional admin user.

**Endpoint:** `POST /tenants`

**Request Body:**
```json
{
  "name": "New Tenant",
  "code": "NEWTENANT",
  "config": {
    "maxUsers": 50,
    "features": ["basic"],
    "theme": "default",
    "domain": "newtenant.example.com"
  },
  "adminUser": {
    "email": "admin@newtenant.example.com",
    "name": "Admin User",
    "password": "securePassword123"
  }
}
```

**Example Request:**
```bash
curl -X POST "https://api.example.com/v1/tenants" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "name": "New Tenant",
    "code": "NEWTENANT",
    "config": {
      "maxUsers": 50,
      "features": ["basic"]
    }
  }'
```

**Response Format:**
```json
{
  "success": true,
  "message": "Tenant created successfully",
  "data": {
    "tenant": {
      "id": "tenant-uuid-new",
      "name": "New Tenant",
      "code": "NEWTENANT",
      "status": "active",
      "config": {
        "maxUsers": 50,
        "features": ["basic"]
      },
      "createdAt": "2025-08-31T10:30:00Z",
      "updatedAt": "2025-08-31T10:30:00Z"
    },
    "adminUser": {
      "id": "admin-uuid-new",
      "email": "admin@newtenant.example.com",
      "name": "Admin User",
      "tenantId": "tenant-uuid-new"
    },
    "message": "Tenant 'New Tenant' and admin user created successfully"
  },
  "requestId": "req-20250831-103000",
  "timestamp": "2025-08-31T10:30:00Z"
}
```

### 4. Update Tenant

Update tenant information (partial updates supported).

**Endpoint:** `PUT /tenants/{tenantId}`

**Path Parameters:**
- `tenantId` (string, required): Unique tenant identifier

**Request Body (all fields optional):**
```json
{
  "name": "Updated Tenant Name",
  "status": "suspended",
  "config": {
    "maxUsers": 200,
    "features": ["basic", "premium", "analytics"],
    "theme": "dark",
    "domain": "updated.example.com"
  }
}
```

**Example Request:**
```bash
curl -X PUT "https://api.example.com/v1/tenants/tenant-uuid-1" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "name": "Updated Tenant Name",
    "status": "active"
  }'
```

**Response Format:**
```json
{
  "success": true,
  "message": "Tenant updated successfully",
  "data": {
    "id": "tenant-uuid-1",
    "name": "Updated Tenant Name",
    "code": "TENANT1",
    "status": "active",
    "config": {
      "maxUsers": 100,
      "features": ["basic", "analytics"]
    },
    "updatedAt": "2025-08-31T11:00:00Z"
  },
  "requestId": "req-20250831-110000",
  "timestamp": "2025-08-31T11:00:00Z"
}
```

### 5. Delete Tenant

Permanently delete a tenant and all associated data.

**Endpoint:** `DELETE /tenants/{tenantId}`

**Path Parameters:**
- `tenantId` (string, required): Unique tenant identifier

**Request Body:**
```json
{
  "confirmation": "DELETE_TENANT_tenant-uuid-1",
  "reason": "user_request",
  "createBackup": true
}
```

**Example Request:**
```bash
curl -X DELETE "https://api.example.com/v1/tenants/tenant-uuid-1" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "confirmation": "DELETE_TENANT_tenant-uuid-1",
    "reason": "user_request",
    "createBackup": true
  }'
```

**Response Format:**
```json
{
  "success": true,
  "message": "Tenant deletion initiated successfully",
  "data": {
    "tenant": {
      "id": "tenant-uuid-1",
      "name": "Tenant One",
      "code": "TENANT1"
    },
    "cleanupResult": {
      "id": "cleanup_tenant-uuid-1_1693478400000",
      "tenantId": "tenant-uuid-1",
      "status": "success",
      "reason": "tenant_deletion",
      "erasureType": "hard",
      "startedAt": "2025-08-31T12:00:00Z",
      "completedAt": "2025-08-31T12:00:30Z",
      "recordsDeleted": {
        "users": 25,
        "roles": 5,
        "audit_logs": 150
      },
      "backupCreated": true,
      "backupId": "backup_tenant-uuid-1_20250831"
    }
  },
  "requestId": "req-20250831-120000",
  "timestamp": "2025-08-31T12:00:00Z"
}
```

### 6. Tenant Status Management

#### Activate Tenant
**Endpoint:** `PUT /tenants/{tenantId}/activate`

#### Suspend Tenant  
**Endpoint:** `PUT /tenants/{tenantId}/suspend`

#### Disable Tenant
**Endpoint:** `PUT /tenants/{tenantId}/disable`

## Error Responses

All API endpoints return standardized error responses with the following format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error description",
    "details": {
      "field": "fieldName",
      "additionalInfo": "value"
    },
    "field": "fieldName"
  },
  "requestId": "req-20250831-120000",
  "timestamp": "2025-08-31T12:00:00Z",
  "path": "/v1/tenants"
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `BAD_REQUEST` | 400 | Invalid request format or parameters |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Tenant not found |
| `CONFLICT` | 409 | Tenant code already exists or operation conflict |
| `VALIDATION_FAILED` | 400 | Request validation failed |
| `RATE_LIMIT_EXCEEDED` | 429 | Rate limit exceeded |
| `INTERNAL_SERVER_ERROR` | 500 | Unexpected server error |
| `DATABASE_ERROR` | 500 | Database operation failed |

### Example Error Responses

**Validation Error:**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Tenant name is required",
    "field": "name"
  },
  "requestId": "req-20250831-120001",
  "timestamp": "2025-08-31T12:00:01Z",
  "path": "/v1/tenants"
}
```

**Tenant Not Found:**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Tenant not found",
    "details": {
      "tenantId": "invalid-uuid"
    }
  },
  "requestId": "req-20250831-120002",
  "timestamp": "2025-08-31T12:00:02Z",
  "path": "/v1/tenants/invalid-uuid"
}
```

**Rate Limit Exceeded:**
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Please try again later."
  },
  "requestId": "req-20250831-120003",
  "timestamp": "2025-08-31T12:00:03Z",
  "path": "/v1/tenants"
}
```

## Rate Limiting

The API implements rate limiting to ensure fair usage and system stability:

- **Default Limit:** 60 requests per minute per client
- **Burst Allowance:** Up to 10 requests in a short burst
- **Headers:** Rate limit information is included in response headers
  - `X-RateLimit-Limit`: Maximum requests per window
  - `X-RateLimit-Remaining`: Requests remaining in current window
  - `X-RateLimit-Reset`: Unix timestamp when the rate limit resets

## Authentication

All API endpoints require Bearer token authentication:

```
Authorization: Bearer YOUR_JWT_TOKEN
```

The JWT token must be obtained through the authentication system and must have appropriate permissions for tenant management operations.

## Data Models

### Tenant Model
```typescript
interface Tenant {
  id: string;
  name: string;
  code: string;
  status: 'active' | 'suspended' | 'disabled';
  config: TenantConfig;
  adminUserId?: string;
  createdAt: string; // ISO 8601 date
  updatedAt: string; // ISO 8601 date
}
```

### Tenant Configuration Model
```typescript
interface TenantConfig {
  maxUsers: number;
  features: string[];
  theme?: string;
  domain?: string;
}
```

### Pagination Model
```typescript
interface Pagination {
  page: number;
  limit: number;
  total: number;
  pages: number;
}
```

## Security Considerations

1. **Authentication Required:** All endpoints require valid JWT authentication
2. **Authorization:** Users can only access tenants they have permission for
3. **Confirmation Required:** Tenant deletion requires explicit confirmation string
4. **Data Isolation:** Complete tenant data isolation is enforced
5. **Audit Logging:** All operations are logged for security and compliance
6. **Input Validation:** Comprehensive validation prevents injection attacks
7. **Rate Limiting:** Protects against abuse and DoS attacks

## Usage Examples

See the individual endpoint sections above for curl examples. For more comprehensive testing, refer to the Postman collection in the next section.