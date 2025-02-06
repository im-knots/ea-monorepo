# EA Job Executor
## Overview

The EA Job Executor is responsible for executing AI workflows by processing interconnected nodes in a directed acyclic graph (DAG). The executor loads a predefined agent job, constructs an execution graph, and processes each node according to its dependencies and parameters. The execution framework is designed to handle API calls, resolve state references, and support dynamic input/output processing.

This executor runs as part of the ea-job-engine as a Kubernetes Job pod. The agentjob.json file is populated by the ea-job-operator from the AgentJob Custom Resource, ensuring that the executor always receives the latest job configuration and executes within the job's lifecycle.
Table of Contents

## Architecture

### The job executor is structured into several key components:

-  Agent Jobs: JSON-defined workflows that specify nodes and edges.
-  Execution Graph: A DAG representing the execution order of nodes.
-  Execution State: A key-value store holding intermediate node outputs.
-  Node Execution: Processes individual nodes based on type (API call, computation, etc.).
-  Kubernetes Job Execution: The executor runs in a controlled environment, ensuring isolation per execution cycle.

## Execution Flow

-  Load the Agent Job: Reads and parses a JSON file describing the workflow.
-  Load the Node Library: Fetches node definitions from the agent manager.
-  Build the Execution Graph: Constructs a DAG to determine execution order based on dependencies.
-  Execute the Graph: Processes nodes in a topological sequence, resolving dependencies dynamically.
-  Store Results: Outputs are saved in the execution state for reference by subsequent nodes.

## Key Components
### Agent Definition

An Agent Job is defined as: 
```json
{ "id": "job-123", "name": "Example Job", "nodes": [ {"alias": "inputNode", "type": "some.node.address", "parameters": {"text": "Hello"}}, {"alias": "processingNode", "type": "some.node.address", "parameters": {}} ], "edges": [ {"from": "inputNode", "to": "processingNode"} ] } 
```

### Execution Graph
The execution graph is a DAG built from the Agent Job structure. It tracks:
-  Nodes: All node instances mapped by alias.
-  Adjacency List: A mapping of node dependencies.
-  In-degree Count: The number of incoming edges per node (for topological sorting).

### DAG Construction
To ensure nodes execute in the correct order, we employ topological sorting using Kahn's Algorithm:

-  Identify nodes with no incoming edges (roots of the graph).
-  Process nodes in order, removing edges as execution progresses.
-  Ensure no cycles exist, guaranteeing the graph remains acyclic.
-  Store execution order for deterministic execution.

This approach ensures that every node is executed only after all its dependencies have been resolved.

### Execution State
The Execution State stores intermediate outputs from nodes: 
```go
 type ExecutionState struct { Results map[string]interface{} // Stores node outputs }
```

This allows later nodes to reference outputs using placeholders (e.g., {{processingNode.response}}).

### Node Execution
Resolving State References

**Nodes** may reference earlier outputs using double-brace syntax ({{node.outputKey}}). The function resolveStateReference() handles these lookups by parsing the reference, retrieving the stored output, and handling nested key resolution.
Handling API-based Nodes

**API-based nodes** define an endpoint and method, which are executed dynamically. The executor prepares a request payload based on node parameters, performs the API call, and stores the response in the execution state for downstream consumption.
Processing Generic Nodes

**Non-API nodes** (e.g., inputs, transformations) process data internally and pass results to the execution state, supporting input merging and computation logic.

## Configuration

The executor loads configuration from config.LoadConfig(), which includes:

-  Agent Manager URL
-  Logging settings
-  Execution parameters
-  Kubernetes-specific configurations for job lifecycle management

## Logging

The executor uses structured logging with levels:

-  INFO: High-level execution steps.
-  WARN: Potentially incorrect but non-fatal issues.
-  ERROR: Failures requiring termination.
-  DEBUG: Detailed execution flow.

Example:  

```go
logger.Slog.Info("Executing node", "nodeType", node.Type) logger.Slog.Error("Execution failed", "error", err)
```

## TODOs

-  Handle Multiple Outputs:
    -  Nodes should support returning multiple outputs that downstream nodes can reference.
    -  Currently, each node stores only a single output.
-  Improve Error Handling:
    -  Standardize error responses from nodes.
    -  Add retries for API-based nodes.
  Enhance Input Merging:
    -  Improve support for multi-input nodes that concatenate multiple outputs into a single prompt.
  Kubernetes Optimizations:
    -  Improve job monitoring and logging for failed executions.

