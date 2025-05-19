# EA Ainu Manager API Documentation

## Overview
The **EA Ainu Manager** is a RESTful API service that fetches users, compute devices, and AI jobs status on the ea platform. It is the primary driver of the frontend. This service also allows users to, register compute devices, and track compute credits.

## Data Model
The API operates with the following primary entities:

### User Definition
```json
{
    "id": "string (UUID)",
    "name": "string",
    "compute_credits": "integer",
    "created_time": "ISO 8601 datetime",
    "compute_devices": [
        {
            "device_name": "string",
            "device_os": "string",
            "compute_type": "string",
            "status": "Active | Offline | Error",
            "compute_rate": "float",
            "id": "string (UUID)",
            "last_active": "ISO 8601 datetime",
            "created_time": "ISO 8601 datetime"
        }
    ],
    "jobs": [
        {
            "job_name": "string",
            "job_type": "string",
            "status": "Active | Offline",
            "last_active": "ISO 8601 datetime",
            "id": "string (UUID)",
            "created_time": "ISO 8601 datetime"
        }
    ]
}
```

## API Endpoints

### Required Headers
All requests to this API coming into the cluster via the api gateway must include an authorization header containing an authenticated user's JWT

```
Authorization: Bearer <YOUR JWT>
```

Internal systems within the cluster (behind the istio gateway) can access this service by providing

(**Note: network level access is restricted in the cluster via NetworkPolicies**)

```
X-Ea-Internal: internal
```


---

### Users

#### Get All Users
**GET** `/api/v1/users`
##### Response
```json
[
    {"id": "UUID", "name": "John Doe"},
    {"id": "UUID", "name": "Jane Smith"}
]
```

#### Get User by ID
**GET** `/api/v1/users/{user_id}`
##### Response
```json
{
    "id": "UUID",
    "name": "John Doe",
    "compute_credits": 1000,
    "compute_devices": [...],
    "jobs": [...],
    "created_time": "ISO 8601 datetime"
}
```

### Compute Devices
#### Add Compute Device
**POST** `/api/v1/users/{user_id}/devices`
##### Request Body
```json
{
    "device_name": "Athena",
    "device_os": "Linux (Ubuntu)",
    "compute_type": "CPU + GPU",
    "status": "Active",
    "compute_rate": 85.0
}
```
##### Response
```json
{
    "message": "Compute device added successfully",
    "user_id": "UUID",
    "device": { ... }
}
```

#### Delete Compute Device
**DELETE** `/api/v1/users/{user_id}/devices/{device_id}`
##### Response
```json
{
    "message": "Compute device removed successfully",
    "user_id": "UUID",
    "device_id": "UUID",
    "device_name": "Athena"
}
```

### Jobs
#### Add Job
**POST** `/api/v1/users/{user_id}/jobs`
##### Request Body
```json
{
    "job_name": "Image Processing",
    "job_type": "ML Model Inference",
    "status": "Active"
}
```
##### Response
```json
{
    "message": "User job added successfully",
    "user_id": "UUID",
    "job": { ... }
}
```

#### Delete Job
**DELETE** `/api/v1/users/{user_id}/jobs/{job_id}`
##### Response
```json
{
    "message": "User job removed successfully",
    "user_id": "UUID",
    "job_id": "UUID",
    "job_name": "Image Processing"
}
```

### Compute Credits
#### Update Compute Credits
**PUT** `/api/v1/users/{user_id}/credits`
##### Request Body
```json
{
    "compute_credits": 5000
}
```
##### Response
```json
{
    "message": "Compute credits updated successfully",
    "user_id": "UUID",
    "compute_credits": 5000
}
```

## Notes
- All IDs are UUIDs.
- All timestamps follow ISO 8601 format.
- Error responses follow the pattern:
  ```json
  {"error": "Description of error"}
  ```

