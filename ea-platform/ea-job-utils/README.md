# Ea Job Utils Service

The **Ea Job Utils Service** provides a set of generic, reusable APIs designed to support various utility functions within the Ea Platform's job execution environment. Currently, it includes endpoints for handling base64 encoding and decoding tasks, essential for data transformation and management in agent workflows.

---

## API Endpoints

### Base64 Decode

**Endpoint:**
```
POST /api/v1/base64decode
```

### Request
The endpoint expects a JSON payload with the following format:

```json
{
  "data": "base64-encoded-string"
}
```

### Example Request

```bash
curl -X POST http://<service-url>/api/v1/base64decode \
-H 'Content-Type: application/json' \
-d '{"data": "SGVsbG8gd29ybGQh"}'
```

### Example Response

```json
{
  "decoded": "Hello, World!"
}
```

### Response Codes

- `200 OK`: Successfully decoded the provided base64 string.
- `400 Bad Request`: Missing or invalid `data` field in the request payload.

---

## Base64 Encode Endpoint

**Endpoint:** `/api/v1/base64encode`

### Request

Encode a plain string to base64 format.

```json
{
  "data": "plain-string"
}
```

### Example Request

```bash
curl -X POST http://<service-url>/api/v1/base64encode \
-H 'Content-Type: application/json' \
-d '{"data": "plain-string-to-encode"}'
```


### Responses

- **Success (200 OK):**

```json
{
  "encoded_value": "base64-encoded-result"
}
```

- **Error (400 Bad Request):**

```json
{
  "error": "Missing 'data' field in request payload"
}
```

---

## Logging and Metrics

All successful operations are logged with details including the encoded or decoded values. Errors and invalid requests are logged clearly for troubleshooting purposes. Metrics for successful encoding and decoding operations are recorded and exposed for monitoring purposes.

---

## Monitoring and Observability

Metrics:
- `StepCounter` increments on every successful encoding or decoding request.
- Labels indicate the endpoint, operation type (`encode` or `decode`), and status (`success`).

---


