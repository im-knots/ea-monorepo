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
| `WatchCleanOrphans` | Detects and resets jobs orphaned by crashed operator pods. |


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
| `FEATURE_CLEAN_ORPHANS` | Enables `WatchCleanOrphans` (for HA recovery) | `false` |

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


## Load testing
Assume Happy Path only with no orphan creation due to crashing operator pods

With CleanOrphans disabled we run 1000 job submissions in a short time

```bash
eru-labs-monorepo/ea-platform/ea-job-api/tests$ for i in {1..1000}; do ./smoke/create-job.sh; done
```

Goal is to observe the time it takes 3 pods of the operator to process 1000 AgentJobs. Then rerun with higher number of pods and observe if there is a linear relationship between processing time or if we get deminishing returns from k8s API saturation. 

### Test 1
Operator pod count: 3
Test start time: 9:00:30 PM 
Job submission completes: 9:01:00 PM
AgentJob complete time: 9:07:45 PM
