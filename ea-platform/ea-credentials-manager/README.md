# Ea Credentials Manager

The **Ea Credentials Manager** service securely manages and updates user-specific credentials required by third-party APIs utilized within Ea Platform's agent workflows. Credentials are stored securely as Kubernetes secrets, accessible exclusively to the relevant users and services.

---

## API Endpoint

### Add or Update Credentials

**Endpoint:**
```
POST /api/v1/credentials
```

### Request Headers

- **`X-Consumer-Username`** *(Required)*: The user identifier injected by the API gateway (Kong).

### Request Body

Provide credentials as key-value pairs in a JSON object:

```json
{
  "api_key": "your-api-key",
  "secret_token": "your-secret-token"
}
```

### Example Request

```bash
curl -X POST http://<service-url>/api/v1/credentials \
-H 'Content-Type: application/json' \
-H 'X-Consumer-Username: user123' \
-d '{"api_key": "abcd1234", "secret_token": "secret9876"}'
```

### Example Response

- **Success (200 OK)**:

```json
{
  "message": "✅ Credentials updated successfully!"
}
```

### Error Responses

- **Unauthorized (401)**:

```json
{
  "error": "Unauthorized: Missing X-Consumer-User header"
}
```

- **Bad Request (400)**:

```json
{
  "error": "Invalid JSON request body"
}
```

- **Internal Server Error (500)**:

```json
{
  "error": "Failed to update secret"
}
```

---

## How it Works

Upon receiving a valid request, the service:
1. Retrieves the `X-Consumer-Username` header to identify the user.
2. Parses and base64 encodes the provided credentials.
3. Uses Kubernetes API to securely PATCH the corresponding secret in the `ea-platform` namespace.

---



## Security

- All credentials are stored securely as Kubernetes secrets.
- Secrets are scoped to individual users.
- Service interactions are authenticated through an API gateway.

---

## Monitoring and Logging

- Operations, including success and failure states, are extensively logged for auditability.
- Metrics are captured for monitoring credential update operations.

---

## Deployment

This service is containerized and operates within the Ea Platform's Kubernetes infrastructure, leveraging native Kubernetes capabilities for secure credential management.

