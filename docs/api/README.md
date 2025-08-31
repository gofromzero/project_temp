# Tenant Management API Documentation

This directory contains comprehensive documentation and examples for the Tenant Management API.

## Files Overview

### 📋 API Specification
- **[tenant-api-specification.md](./tenant-api-specification.md)** - Complete API documentation including endpoints, request/response formats, error codes, and data models

### 🔧 Testing Tools
- **[Tenant-Management-API.postman_collection.json](./Tenant-Management-API.postman_collection.json)** - Postman collection for interactive API testing
- **[curl-examples.sh](./curl-examples.sh)** - Comprehensive bash script with cURL examples for all endpoints

### 🧪 Documentation Tests
- **[backend/tests/documentation/api_documentation_test.go](../../backend/tests/documentation/api_documentation_test.go)** - Automated tests to verify documentation accuracy

## Quick Start

### Using Postman Collection

1. Import the Postman collection: `Tenant-Management-API.postman_collection.json`
2. Set collection variables:
   - `base_url`: Your API base URL (e.g., `https://api.example.com/v1`)
   - `jwt_token`: Your authentication token
3. Run the collection to test all endpoints

### Using cURL Examples

1. Set environment variables:
   ```bash
   export BASE_URL="https://api.example.com/v1"
   export JWT_TOKEN="your-jwt-token-here"
   ```

2. Run the examples script:
   ```bash
   chmod +x curl-examples.sh
   ./curl-examples.sh
   ```

### Manual Testing

Refer to the [API specification](./tenant-api-specification.md) for detailed endpoint documentation with example requests and responses.

## API Endpoints Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/tenants` | List tenants with pagination and filtering |
| `GET` | `/tenants/{id}` | Get specific tenant details |
| `POST` | `/tenants` | Create new tenant |
| `PUT` | `/tenants/{id}` | Update tenant (partial updates supported) |
| `DELETE` | `/tenants/{id}` | Delete tenant and all data |
| `PUT` | `/tenants/{id}/activate` | Activate tenant |
| `PUT` | `/tenants/{id}/suspend` | Suspend tenant |
| `PUT` | `/tenants/{id}/disable` | Disable tenant |

## Authentication

All endpoints require Bearer token authentication:

```
Authorization: Bearer YOUR_JWT_TOKEN
```

## Rate Limiting

- **Default Limit:** 60 requests per minute per client
- **Burst Allowance:** Up to 10 requests in a short burst
- Rate limit headers are included in responses

## Error Handling

The API returns standardized error responses with:
- HTTP status codes
- Error codes (e.g., `VALIDATION_FAILED`, `NOT_FOUND`)
- Detailed error messages
- Request tracking IDs

## Data Models

### Core Models
- **Tenant**: Main tenant entity with ID, name, code, status, and configuration
- **TenantConfig**: Configuration settings including max users, features, theme, and domain
- **Pagination**: Standard pagination information for list endpoints

### Request/Response Types
- **CreateTenantRequest**: For tenant creation with optional admin user
- **UpdateTenantRequest**: For partial tenant updates
- **ListTenantsRequest**: For filtered and paginated tenant listing
- **DeleteTenantRequest**: For secure tenant deletion with confirmation

## Security Features

1. **JWT Authentication**: Required for all endpoints
2. **Tenant Isolation**: Complete data isolation between tenants
3. **Confirmation Required**: Secure deletion requires explicit confirmation
4. **Input Validation**: Comprehensive validation prevents injection attacks
5. **Rate Limiting**: Protects against abuse and DoS attacks
6. **Audit Logging**: All operations are logged for compliance

## Testing Your Integration

1. **Unit Tests**: Use the documentation tests to verify structure compatibility
2. **Integration Tests**: Use Postman collection or cURL examples for end-to-end testing
3. **Error Scenarios**: Test error handling with invalid data and edge cases
4. **Rate Limits**: Test rate limiting behavior with rapid requests

## Support

For additional help with the API:

1. Review the complete [API specification](./tenant-api-specification.md)
2. Check the example requests in the [cURL examples](./curl-examples.sh)
3. Use the [Postman collection](./Tenant-Management-API.postman_collection.json) for interactive testing
4. Run the documentation tests to verify compatibility

## Version Information

- **API Version:** v1
- **Documentation Version:** 1.0.0
- **Last Updated:** 2025-08-31

---

**Note:** This API is part of the multi-tenant management system implementing Story 2.3 requirements for third-party system integration.