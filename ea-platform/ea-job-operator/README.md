# EA Job Operator

## Overview
The **EA Job Operator** is a Kubernetes-based operator service responsible for managing the lifecycle of **AgentJob** custom resources (CRs). It ensures smooth orchestration of job execution within the `ea-platform` namespace. The operator consists of multiple independently running controllers (watch loops) to handle different stages of the job lifecycle.

## Features
- **AgentJob State Management**
  - Detects and marks newly created jobs as `inactive`.
  - Moves `inactive` jobs to `executing` and spawns a corresponding Kubernetes Job to execute a users AgentJob workflow.
  - Watches for `executing` jobs that complete and updates their state to `completed`.
  - Cleans up `completed` jobs after a configured duration.
  
- **TODO: Orphaned Job Recovery**
  - Identifies jobs whose assigned operator pod has failed.
  - Resets job state and removes orphaned locks to allow for retrying.
  - Deletes corresponding Kubernetes Jobs to ensure a fresh restart.

- **Scalability & Configurability**
  - Uses feature flags via environment variables to enable/disable specific operators.
  - Supports HA (High Availability) by distributing workloads across multiple pods.

## Architecture
The EA Job Operator runs multiple watch loops, each responsible for handling a specific part of the job lifecycle:

| Watch Function | Description |
|---------------|-------------|
| `WatchNewAgentJobs` | Detects new AgentJobs with blank statuses and marks them as `inactive`. |
| `WatchInactiveAgentJobs` | Moves `inactive` jobs to `pending`, locks them, and spawns Kubernetes Jobs. |
| `WatchCompletedJobs` | Watches Kubernetes Jobs and updates corresponding AgentJobs upon completion. |
| `WatchCompletedAgentJobs` | Cleans up completed jobs older than a certain threshold. |


## Configuration
Configuration is managed through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Port for metrics endpoint | `8080` |
| `GIN_MODE` | Gin mode (`release` or `debug`) | `release` |
| `FEATURE_NEW_AGENT_JOBS` | Enables `WatchNewAgentJobs` | `true` |
| `FEATURE_INACTIVE_AGENT_JOBS` | Enables `WatchInactiveAgentJobs` | `true` |
| `FEATURE_COMPLETED_JOBS` | Enables `WatchCompletedJobs` | `true` |
| `FEATURE_COMPLETED_AGENT_JOBS` | Enables `WatchCompletedAgentJobs` | `true` |

To modify the configuration, update your deployment environment variables in `chart/values.yaml`.

## Observability
### Metrics
The operator exposes Prometheus-compatible metrics at:
```
http://<pod-ip>:8080/metrics
```
Key metrics include:
- `operator_step_counter`: Counts execution steps for each watch function.


### Logs
Structured logs are emitted in JSON format and can be collected using `kubectl logs`:
```sh
kubectl logs -f deployment/ea-job-operator -n ea-platform
```

## TODO improve scaling
Right now this works pretty well for 100s of jobs in a short period of time. If we want to go to millions of jobs a min we will need a more sophisticated mecahnism to allow actual pod performance improvements when we scale out the number of operator pods. 

Today how each operator loop picks up work they just step all over eachother causing update conflict errors. We can fix these with these approaches together probably. 
Its fine for now though. 
- job sharding for multi pod performance improvement
- implement kubernetes finalizers for job cleanup