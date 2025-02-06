package executor

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"ea-job-executor/config"
	"ea-job-executor/logger"
)

//
// structs and types for the executor
//

// NodeInstance represents a reference to a node definition.
type NodeInstance struct {
	Alias      string                 `json:"alias,omitempty"`
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// Edge represents a connection between nodes in an agent workflow.
type Edge struct {
	From MultiString `json:"from"`
	To   MultiString `json:"to"`
}

// Metadata holds timestamps for Agents.
type Metadata struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Agent represents an AI workflow with interconnected nodes.
type Agent struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Creator     string         `json:"creator"`
	Description string         `json:"description"`
	Nodes       []NodeInstance `json:"nodes"`
	Edges       []Edge         `json:"edges"`
	Metadata    Metadata       `json:"metadata"`
}

// ExecutionGraph represents a DAG of node execution order.
type ExecutionGraph struct {
	Nodes          map[string]NodeInstance // Maps node types to actual node instances
	AdjList        map[string][]string     // Maps a node to the list of nodes it should trigger next
	Indegrees      map[string]int          // Keeps track of incoming edges for topological sorting
	ExecutionOrder []string                // Stores nodes in execution order
}

// NodesLibrary represents the full set of nodes available from the agent manager.
type NodesLibrary []NodeDefinition

// ExecutionState maintains intermediate results of executed nodes.
type ExecutionState struct {
	Results map[string]interface{} // Stores outputs of nodes
}

// NodeDefinition represents a node type with its parameters and API details.
type NodeDefinition struct {
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	Creator    string          `json:"creator"`
	API        APIConfig       `json:"api"`
	Parameters []NodeParameter `json:"parameters"`
	Outputs    []NodeOutput    `json:"outputs"`
	Metadata   NodeMetadata    `json:"metadata"`
}

// APIConfig represents the API details for API-based nodes.
type APIConfig struct {
	BaseURL  string            `json:"baseurl"`
	Endpoint string            `json:"endpoint"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers"`
}

// NodeParameter represents a parameter a node accepts.
type NodeParameter struct {
	Key         string      `json:"key"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default"`
}

// NodeOutput represents the expected output of a node.
type NodeOutput struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// NodeMetadata contains additional metadata about a node.
type NodeMetadata struct {
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Additional  map[string]interface{} `json:"additional"`
}

//
// executor function handles the overall execution for an AgentJob
//

func ExecuteAgentJob() {
	// Load configuration
	config := config.LoadConfig()

	logger.Slog.Info("Starting job execution")

	// Step 1: Read the file with the agentJob
	filePath := "agentjob.json"
	job, err := stepReadAgentJob(filePath)
	if err != nil {
		logger.Slog.Error("Failed to read job file", "file", filePath, "error", err)
		os.Exit(1)
	} else {
		// Log the parsed job content
		logger.Slog.Info("Agent Job Loaded Successfully", "job", job)
	}

	// Step 2: Load ea-agent-manager nodes library
	nodesLib, err := stepLoadNodesLibrary(config.AgentManagerUrl)
	if err != nil {
		logger.Slog.Error("Failed to fetch nodes library", "error", err)
		os.Exit(1)
	} else {
		logger.Slog.Info("Successfully loaded nodes library", "total_nodes", len(nodesLib))
	}

	// Step 3: Build the exectuion graph
	execGraph, err := stepBuildExecutionGraph(job)
	if err != nil {
		logger.Slog.Error("Failed to build execution graph", "error", err)
		os.Exit(1)
	} else {
		logger.Slog.Info("Successfully built execution graph", "graph", execGraph)
	}

	// Step 4: Execute the graph
	finalNodeOutput, err := stepExecuteGraph(execGraph, nodesLib)
	if err != nil {
		logger.Slog.Error("Execution of the graph failed", "error", err)
		os.Exit(1)
	} else {
		logger.Slog.Info("Final Node Output", finalNodeOutput)
		logger.Slog.Info("Agent job execution completed successfully")
	}

	os.Exit(0)

}

//
// step functions handle the execution steps for an AgentJob
//

// stepReadAgentJob reads and unmarshals an agent job from a JSON file.
func stepReadAgentJob(filePath string) (Agent, error) {
	// Read the agent-job.json file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Agent{}, fmt.Errorf("failed to read job file: %w", err)
	}

	// Parse the JSON content into the Agent struct
	var job Agent
	if err := json.Unmarshal(data, &job); err != nil {
		return Agent{}, fmt.Errorf("failed to parse job JSON: %w", err) // Return an empty Agent and error
	}

	// Return the parsed Agent and no error
	return job, nil
}

// stepLoadNodesLibrary fetches the full library of nodes from the agent-manager API.
func stepLoadNodesLibrary(agentManagerURL string) (NodesLibrary, error) {
	// Step 1: Fetch the basic node list
	nodesListURL := agentManagerURL + "/nodes"
	resp, err := http.Get(nodesListURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nodes list request failed with status %d", resp.StatusCode)
	}

	// Parse response body
	var nodeSummaries []struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Creator string `json:"creator"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&nodeSummaries); err != nil {
		return nil, err
	}

	// Step 2: Fetch full details for each node
	var nodesLib NodesLibrary
	for _, summary := range nodeSummaries {
		nodeDetailURL := fmt.Sprintf("%s/nodes/%s", agentManagerURL, summary.ID)
		resp, err := http.Get(nodeDetailURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Failed to fetch node details: %d", resp.StatusCode)
		}

		// Parse full node definition
		var nodeDef NodeDefinition
		if err := json.NewDecoder(resp.Body).Decode(&nodeDef); err != nil {
			return nil, err
		}

		// Log retrieved node details
		logger.Slog.Info("Fetched full node details", "nodeID", summary.ID, "nodeType", nodeDef.Type)

		nodesLib = append(nodesLib, nodeDef)
	}

	// Final check: Ensure we have nodes
	if len(nodesLib) == 0 {
		return nil, fmt.Errorf("failed to fetch any valid node definitions")
	}

	logger.Slog.Info("Successfully loaded full node definitions", "total_nodes", len(nodesLib))

	return nodesLib, nil
}

func stepBuildExecutionGraph(agent Agent) (ExecutionGraph, error) {
	execGraph := ExecutionGraph{
		Nodes:          make(map[string]NodeInstance),
		AdjList:        make(map[string][]string),
		Indegrees:      make(map[string]int),
		ExecutionOrder: []string{},
	}

	nodeAliases := make(map[string]string) // alias -> alias mapping
	nodeIDs := []string{}                  // Maintain original JSON node order

	// Step 1: Store nodes in the graph
	for _, node := range agent.Nodes {
		nodeID := node.Alias
		if nodeID == "" {
			nodeID = node.Type
		}

		nodeAliases[nodeID] = nodeID
		execGraph.Nodes[nodeID] = node
		execGraph.AdjList[nodeID] = []string{}
		execGraph.Indegrees[nodeID] = 0
		nodeIDs = append(nodeIDs, nodeID)
	}

	// Step 2: Build adjacency list and track incoming connections
	for _, edge := range agent.Edges {
		for _, fromAlias := range edge.From {
			for _, toAlias := range edge.To {
				fromID, fromExists := nodeAliases[fromAlias]
				toID, toExists := nodeAliases[toAlias]

				if fromExists && toExists {
					execGraph.AdjList[fromID] = append(execGraph.AdjList[fromID], toID)
					execGraph.Indegrees[toID]++
				} else {
					return ExecutionGraph{}, fmt.Errorf("edge references unknown alias from %s to %s", fromAlias, toAlias)
				}
			}
		}
	}

	// Step 3: Stable Topological Sorting (preserves JSON order)
	queue := []string{}

	// Identify nodes with no incoming edges (roots)
	for _, nodeID := range nodeIDs {
		if execGraph.Indegrees[nodeID] == 0 {
			queue = append(queue, nodeID)
		}
	}

	// Step 4: Process nodes in a stable topological order
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		execGraph.ExecutionOrder = append(execGraph.ExecutionOrder, current)

		for _, neighbor := range execGraph.AdjList[current] {
			execGraph.Indegrees[neighbor]--
			if execGraph.Indegrees[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Step 5: Validate graph completeness (check for cycles)
	if len(execGraph.ExecutionOrder) != len(execGraph.Nodes) {
		return ExecutionGraph{}, errors.New("cyclic dependency detected in execution graph")
	}

	logger.Slog.Info("Execution order determined", "ExecutionOrder", execGraph.ExecutionOrder)
	return execGraph, nil
}

// stepExecuteGraph executes the graph in order.
func stepExecuteGraph(execGraph ExecutionGraph, nodesLib NodesLibrary) (interface{}, error) {
	state := ExecutionState{Results: make(map[string]interface{})}

	var finalNodeOutput interface{} // Declare finalNodeOutput

	for _, nodeType := range execGraph.ExecutionOrder {
		node, exists := execGraph.Nodes[nodeType]
		if !exists {
			return nil, fmt.Errorf("Node type not found in execution graph %s", nodeType)
		}

		logger.Slog.Info("Executing node", "nodeType", nodeType)

		nodeDef, err := findNodeDefinition(node.Type, nodesLib)
		if err != nil {
			return nil, err
		}

		logger.Slog.Debug("Found node definition", "nodeType", node.Type, "definition", nodeDef)

		result, err := executeNode(node, nodeDef, &state)
		if err != nil {
			return nil, err
		}

		// Store results of the node execution
		state.Results[nodeType] = result
		logger.Slog.Info("Successfully executed node", "nodeType", nodeType)
		logger.Slog.Debug("Successfully executed node", "nodeType", nodeType, "result", result)

		// Store the result of the final node
		if nodeType == execGraph.ExecutionOrder[len(execGraph.ExecutionOrder)-1] {
			finalNodeOutput = result
			logger.Slog.Info("Final node output stored", "nodeAlias", node.Alias, "result", finalNodeOutput)
		}
	}

	logger.Slog.Info("Graph execution completed successfully")
	return finalNodeOutput, nil
}

//
// Node execution functions handle the execution of nodes within the workflow
//

// executeNode determines the node type and delegates execution accordingly.
func executeNode(node NodeInstance, nodeDef NodeDefinition, state *ExecutionState) (interface{}, error) {
	// Only merge inputs for nodes that actually need it
	node = mergeInputsForNode(node, nodeDef, state)

	// Inject values into parameters from execution state
	for key, value := range node.Parameters {
		if strVal, ok := value.(string); ok && strings.Contains(strVal, "{{") {
			resolvedVal, exists, err := resolveStateReference(strVal, state)
			if err != nil {
				logger.Slog.Error("Failed to resolve reference", "key", key, "value", strVal, "error", err)
				return nil, err
			} else if exists {
				node.Parameters[key] = resolvedVal
				logger.Slog.Info("Resolved reference for parameter", "key", key, "value", resolvedVal)
			}
		}
	}

	// Validate and populate parameters
	validatedParams, err := validateAndFillParameters(node.Parameters, nodeDef.Parameters)
	if err != nil {
		return nil, err
	}
	node.Parameters = validatedParams

	// Execute API-based nodes
	var result interface{}
	if nodeDef.API.BaseURL != "" {
		logger.Slog.Info("Executing API Node", "nodeType", node.Type)
		result, err = executeAPINode(node, nodeDef)
	} else {
		logger.Slog.Info("Executing Generic Node", "nodeType", node.Type)
		result, err = executeGenericNode(node, nodeDef, state)
	}

	if err != nil {
		return nil, err
	}

	// Store outputs in execution state for downstream nodes
	state.Results[node.Alias] = result
	logger.Slog.Debug("Execution state after storing result", "state", state.Results)

	logger.Slog.Debug("Stored node result in execution state", "nodeAlias", node.Alias, "result", result)

	return result, nil
}

// executeAPINode makes an HTTP request based on the node definition.
func executeAPINode(node NodeInstance, nodeDef NodeDefinition) (interface{}, error) {
	url := nodeDef.API.BaseURL + nodeDef.API.Endpoint
	bodyData, err := json.Marshal(node.Parameters)
	if err != nil {
		return nil, err
	}

	logger.Slog.Info("Sending API Request", "nodeType", node.Type, "URL", url, "Payload", string(bodyData))

	req, err := http.NewRequest(nodeDef.API.Method, url, bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, err
	}

	for key, value := range nodeDef.API.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Log the full response body
	logger.Slog.Info("Received API Response", "nodeType", node.Type, "RawResponse", string(responseData))

	var result map[string]interface{}
	if err := json.Unmarshal(responseData, &result); err != nil {
		return nil, err
	}

	// Store the full API response so it can be referenced dynamically
	logger.Slog.Info("Storing full API response under node alias", "nodeAlias", node.Alias, "response", result)
	return result, nil
}

// executeGenericNode handles simple input/output storage operations.
func executeGenericNode(node NodeInstance, nodeDef NodeDefinition, state *ExecutionState) (interface{}, error) {
	// If it's an input node, return its parameter values
	if len(node.Parameters) > 0 {
		return node.Parameters, nil
	}

	// If it's an output node, retrieve and store output from execution state
	for _, param := range nodeDef.Parameters {
		if val, exists := state.Results[param.Key]; exists {
			state.Results[node.Type] = val
		}
	}

	return nil, nil
}

//
// HELPER FUNCTIONS
//

// MultiString allows a field to be either a single string or an array of strings.
type MultiString []string

// UnmarshalJSON for MultiString allows handling single strings as arrays.
func (m *MultiString) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*m = []string{single}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err == nil {
		*m = multiple
		return nil
	}

	return json.Unmarshal(data, m)
}

// FindNodeDefinition retrieves the node definition from the nodes library.
func findNodeDefinition(nodeType string, nodesLib NodesLibrary) (NodeDefinition, error) {
	for _, node := range nodesLib {
		if node.Type == nodeType {
			logger.Slog.Debug("Found node definition", "nodeType", nodeType, "nodeDef", node)
			return node, nil
		}
	}
	return NodeDefinition{}, fmt.Errorf("node definition not found for type %s", nodeType)
}

// validateAndFillParameters ensures parameters match the definition and fills in missing ones
func validateAndFillParameters(providedParams map[string]interface{}, paramDefs []NodeParameter) (map[string]interface{}, error) {
	validatedParams := make(map[string]interface{})

	// Build a map of expected parameters
	expectedParams := make(map[string]NodeParameter)
	for _, param := range paramDefs {
		expectedParams[param.Key] = param
	}

	// Check for invalid parameters
	for key := range providedParams {
		if _, exists := expectedParams[key]; !exists {
			return nil, fmt.Errorf("unexpected parameter: %s", key)
		}
	}

	// Fill missing parameters with defaults
	for key, paramDef := range expectedParams {
		if value, exists := providedParams[key]; exists {
			validatedParams[key] = value
		} else {
			validatedParams[key] = paramDef.Default // Use default if not provided
			logger.Slog.Warn("Missing parameter, using default", "key", key, "default", paramDef.Default)
		}
	}

	return validatedParams, nil
}

// injectInputsFromState replaces placeholders in parameters with values from execution state.
func injectInputsFromState(parameters map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	updatedParams := make(map[string]interface{})

	for key, value := range parameters {
		// Check if value is a string reference (e.g., "{{response.Result}}")
		if strVal, ok := value.(string); ok {
			if strings.HasPrefix(strVal, "{{") && strings.HasSuffix(strVal, "}}") {
				trimmedRef := strings.TrimPrefix(strings.TrimSuffix(strVal, "}}"), "{{")
				resolvedValue, exists, err := resolveStateReference(trimmedRef, state)
				if err != nil {
					return nil, err
				} else if exists {
					updatedParams[key] = resolvedValue
					logger.Slog.Info("Injected value from execution state", "paramKey", key, "value", resolvedValue)
					continue
				}
			}
		}
		// Otherwise, keep the original parameter
		updatedParams[key] = value
	}

	return updatedParams, nil
}

func resolveStateReference(reference string, state *ExecutionState) (interface{}, bool, error) {
	parts := strings.Split(strings.Trim(reference, "{}"), ".")
	if len(parts) < 2 {
		return nil, false, fmt.Errorf("Invalid reference format: %s", reference)
	}

	nodeAlias, keys := parts[0], parts[1:]

	// Retrieve the node result from execution state
	nodeResult, exists := state.Results[nodeAlias]
	if !exists {
		logger.Slog.Warn("Failed to resolve state reference: node not found", "reference", reference)
		logger.Slog.Warn("Current execution state", "state", state.Results)
		return nil, false, fmt.Errorf("Failed to resolve state reference: node not found. Reference: %s", reference)
	}

	// Ensure nodeResult is a map before extracting keys
	nodeMap, ok := nodeResult.(map[string]interface{})
	if !ok {
		logger.Slog.Warn("Node result is not a map", "nodeAlias", nodeAlias, "actualType", fmt.Sprintf("%T", nodeResult))
		return nil, false, fmt.Errorf("Node result is not a map: %T", nodeResult)
	}

	// Recursively retrieve the nested value
	value, found := getNestedValue(nodeMap, keys)
	if !found {
		logger.Slog.Warn("Failed to resolve state reference: key not found", "reference", reference)
		return nil, false, fmt.Errorf("Failed to resolve state reference: key not found. Reference: %s", reference)
	}

	logger.Slog.Info("Successfully resolved reference", "reference", reference, "value", value)
	return value, true, nil
}

// getNestedValue recursively retrieves a nested value from a JSON-like map.
func getNestedValue(data interface{}, keys []string) (interface{}, bool) {
	if len(keys) == 0 {
		return data, true
	}

	currentKey := keys[0]

	// Ensure the data is a map[string]interface{} before accessing deeper keys
	if nestedMap, ok := data.(map[string]interface{}); ok {
		if nextValue, exists := nestedMap[currentKey]; exists {
			return getNestedValue(nextValue, keys[1:]) // Recurse deeper
		}
	}

	return nil, false
}

func mergeInputsForNode(targetNode NodeInstance, nodeDef NodeDefinition, state *ExecutionState) NodeInstance {
	// Ensure the node actually expects a "prompt"
	hasPrompt := false
	for _, param := range nodeDef.Parameters {
		if param.Key == "prompt" {
			hasPrompt = true
			break
		}
	}

	// If the node does not expect "prompt", return as-is
	if !hasPrompt {
		logger.Slog.Debug("Skipping input merge for node without 'prompt' parameter", "nodeAlias", targetNode.Alias)
		return targetNode
	}

	var mergedInputs []string
	logger.Slog.Info("Checking for inputs to merge for node", "nodeAlias", targetNode.Alias)

	for source, result := range state.Results {
		// Ensure the result is a map with an "input" key
		if inputMap, ok := result.(map[string]interface{}); ok {
			if userInput, found := inputMap["input"].(string); found {
				// Only merge the actual user input text, not the placeholder
				mergedInputs = append(mergedInputs, userInput)
				logger.Slog.Debug("Appending user input from source", "source", source, "value", userInput)
			} else {
				logger.Slog.Warn("Source input does not contain a string", "source", source, "value", inputMap)
			}
		} else {
			logger.Slog.Warn("Unexpected input format", "source", source, "value", result)
		}
	}

	// Join all user inputs into a single prompt
	if len(mergedInputs) > 0 {
		finalPrompt := strings.Join(mergedInputs, " ")
		targetNode.Parameters["prompt"] = finalPrompt
		logger.Slog.Info("Merged inputs for node", "nodeAlias", targetNode.Alias, "mergedPrompt", finalPrompt)
	}

	return targetNode
}
