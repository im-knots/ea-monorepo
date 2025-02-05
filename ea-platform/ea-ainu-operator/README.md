# Ea Ainulindale Operator

## Overview
The **Ea Ainulindale Operator** (`ea-ainu-operator`) is a Kubernetes-based operator that watches for changes in **AgentJob** Custom Resources (CRs) and updates the corresponding job entries in the **ea-ainu-manager** MongoDB (`ainuUsers.users` collection). This operator ensures that job states in Kubernetes are reflected in the **ea-ainu-manager** user database to push `AgentJob` status updates to the **ea-frontend**.

## Features
- Watches for **new** `AgentJob` CRs and adds them to the user's job array in MongoDB as **"New"**.
- Watches for **inactive** `AgentJob` CRs and updates the job entry to **"Pending"**.
- Watches for **completed** `AgentJob` CRs and updates the job entry to **"Complete"**.
- Watches for **errored** `AgentJob` CRs and updates the job entry to **"Error"**.
- Uses work queues to efficiently process and update job states without excessive API requests.
- Implements efficient Kubernetes informers for monitoring `AgentJob` state transitions.

## Architecture
The Ea Ainu Operator consists of the following key components:
- **Informer Watchers**: Listens to changes in `AgentJob` CRs.
- **Work Queues**: Throttles updates to avoid excessive load.
- **MongoDB Integration**: Updates the user database to reflect job status changes.
- **Parallel Processing**: Uses Goroutines to handle multiple job updates efficiently.

## AgentJob CRD
The `AgentJob` CRD has the following key fields:
```yaml
spec:
  agentID: "<UUID>"  # ID of the agent executing the job
  name: "Job Name"   # Human-readable job name
  user: "<UUID>"     # User who initiated the job
  creator: "<UUID>"  # User who created the agent
  nodes: [...]       # Workflow definition
  edges: [...]       # Workflow dependencies
status:
  state: "Pending"   # Possible values: "New", "Pending", "Executing", "Complete", "Error"
  message: "Job details"
```

## Configuration
The operator uses environment variables to define MongoDB connection settings and feature flags. The defaults for feature flags configured in `config/config.go` should be sufficient. You can override all configuration settings via the `chart/values.yaml` like so:

```yaml
config:
  GIN_MODE: release
  PORT: 8080
  FEATURE_NEW_AGENT_JOBS: "true"
  FEATURE_INACTIVE_AGENT_JOBS: "true"
  FEATURE_COMPLETED_AGENT_JOBS: "true"
  FEATURE_ERROR_AGENT_JOBS: "true"

secrets:
  DB_URL: mongodb://mongodb.ea-platform.svc.cluster.local
```

## How It Works
### Event Flow
1. A new **AgentJob** CR is created.
2. The `watchNewAgentJobs` informer detects the new job and queues it.
3. The `processNewJobQueue` function extracts the `userID`, `jobID`, and other metadata, then **adds the job to the user's job array in MongoDB**.
4. As the job progresses (`Inactive` → `Executing` → `Complete`/`Error`), different watchers update the job entry in MongoDB accordingly.

## Deploying the Operator Per Feature for Scalability
The **Ea Ainulindale Operator** is designed with feature flags to enable or disable specific job state handlers. This allows the operator to be deployed in multiple instances, each handling a single job state for improved scalability.

### Example Deployment Strategy:
1. **Single Feature Per Deployment:** Deploy separate instances of the operator for each feature (e.g., `ea-ainu-operator-new`, `ea-ainu-operator-inactive`).
2. **Feature Flag Control:** Use Helm values to configure which features are enabled in each deployment:

   ```yaml
   # Deployment for handling new AgentJobs only
   config:
     FEATURE_NEW_AGENT_JOBS: "true"
     FEATURE_INACTIVE_AGENT_JOBS: "false"
     FEATURE_COMPLETED_AGENT_JOBS: "false"
     FEATURE_ERROR_AGENT_JOBS: "false"
   ```

3. **Pod Autoscaling:** With each feature running in its own deployment, Kubernetes **Horizontal Pod Autoscaler (HPA)** can scale up operators that are processing high traffic while keeping inactive ones minimal.
4. **Load Distribution:** Avoids contention by ensuring each pod handles a specific workload, reducing conflicts on MongoDB updates and API rate limits.

## Logs & Debugging
- View logs:
  ```sh
  kubectl logs -f deployment/ea-ainu-operator
  ```
- Check MongoDB updates:
  ```sh
  mongo ainuUsers --eval "db.users.find().pretty()"
  ```

## Next Steps
- Implement retries for MongoDB updates to handle transient failures.
- Add Prometheus metrics for monitoring job processing efficiency.
- Implement role-based access control (RBAC) for security improvements.

