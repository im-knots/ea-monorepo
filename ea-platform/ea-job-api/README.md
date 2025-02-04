# Ea Job Engine API

## Overview
The **Ea Job Engine API** is a stateless microservice responsible for orchestrating job execution requests in the Ea platform. It handles incoming job requests, fetches the corresponding agent definitions from the Ea Agent Manager, and creates Kubernetes Custom Resources (CRs) representing the jobs. These CRs are defined as instantiations of the CRD found in the chart/crds directory. These CRs are then processed by the Ea Job Operator for execution.

## Features
- Accepts job creation requests via HTTP API.
- Fetches agent definitions from the Ea Agent Manager.
- Creates `AgentJob` CRs in Kubernetes.
- Stateless design for scalability.

## API Endpoints

### **1. Create a Job**
**Endpoint:**
```
POST /api/v1/jobs
```
**Description:**
Submits a new job request, which creates an `AgentJob` CR in Kubernetes.

**Request Body:**
```json
{
  "agent_id": "<AGENT_ID>",
  "user_id": "<USER_ID>"
}
```

**Response:**
```json
{
  "status": "job created",
  "job_name": "agentjob-<AGENT_ID>-<TIMESTAMP>",
  "user_id": "<USER_ID>"
}
```

**Example Request:**
```sh
curl -X POST http://localhost:8084/api/v1/jobs \
     -H "Content-Type: application/json" \
     --data '{
                "agent_id": "ecc21c86-24ee-4b36-803e-c63616325132",
                "user_id": "e40af905-a6bf-4ef9-bb89-30c6d254afd9"
      }'
```

**Response Example:**
```json
{
  "jobName":"agentjob-ecc21c86-24ee-4b36-803e-c63616325132-8c8db6",
  "status":"job created",
  "user":"e40af905-a6bf-4ef9-bb89-30c6d254afd9"
}
```

## Architecture
1. **API receives job request**: A user submits a job creation request via the API.
2. **Fetch agent definition**: The API fetches the agent definition from the Ea Agent Manager.
3. **Create Kubernetes CR**: The API creates an `AgentJob` CR in the Kubernetes cluster.
4. **Ea Job Operator handles execution**: The operator watches for `AgentJob` CRs and executes the job.
5. **Ea Job Operator handles completion and cleanup**: The Ea Job Operator updates the job status and deletes the CR after execution.

## Deployment
The Ea Job API is designed to run as a Kubernetes service and can be deployed using its helm chart in the `/charts` directory

## Environment Variables
| Variable | Description |
|----------|------------|
| `AGENT_MANAGER_URL` | URL of the Ea Agent Manager |
| `PORT` | Port on which the API runs |

## Scalability Considerations
- **Stateless API**: Ensures easy horizontal scaling.
- **Short-lived CRs**: Prevents `etcd` overload by removing completed jobs.
- **Job execution is handled by the operator**: API remains lightweight.

## Future Enhancements
- Support for batch job creation.
- Authentication and rate limiting.
- Improved observability with tracing and logging.

