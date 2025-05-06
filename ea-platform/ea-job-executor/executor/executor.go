package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"ea-job-executor/config"
	"ea-job-executor/logger"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//--------------------- Struct Definitions ---------------------//

type NodeInstance struct {
	Alias      string                 `json:"alias,omitempty"`
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type Edge struct {
	From []string `json:"from"`
	To   []string `json:"to"`
}

type Metadata struct {
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	AgentJobID string    `json:"agent_job_id"`
}

type Agent struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Creator     string         `json:"creator"`
	Description string         `json:"description"`
	Nodes       []NodeInstance `json:"nodes"`
	Edges       []Edge         `json:"edges"`
	Metadata    Metadata       `json:"metadata"`
}

type ExecutionGraph struct {
	Nodes          map[string]NodeInstance
	AdjList        map[string][]string
	Indegrees      map[string]int
	ExecutionOrder []string
}

type ExecutionState struct {
	Results map[string]interface{}
	Lock    sync.RWMutex
}

type NodeDefinition struct {
	Type       string          `json:"type"`
	API        APIConfig       `json:"api"`
	Parameters []NodeParameter `json:"parameters"`
	Outputs    []NodeOutput    `json:"outputs"`
}

type APIConfig struct {
	BaseURL  string            `json:"baseurl"`
	Endpoint string            `json:"endpoint"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers"`
}

type NodeParameter struct {
	Key     string      `json:"key"`
	Default interface{} `json:"default"`
}

type NodeOutput struct {
	Key string `json:"key"`
}

// NodesLibrary represents the full set of nodes available from the agent manager.
type NodesLibrary []NodeDefinition

//--------------------- Main Executor Function ---------------------//

func ExecuteAgentJob(filePath string) {
	config := config.LoadConfig()

	agent, err := loadAgentJob(filePath)
	if err != nil {
		handleError(err, "Failed to load agent job")
		os.Exit(1)
	}

	nodesLib, err := loadNodesLibrary(config.AgentManagerUrl)
	if err != nil {
		handleError(err, "Failed to load nodes library")
	}

	graph, err := buildExecutionGraph(agent)
	if err != nil {
		handleError(err, "Failed to build execution graph")
	}

	finalOutput, err := executeGraph(agent.Metadata.AgentJobID, agent.Creator, graph, nodesLib)
	if err != nil {
		handleError(err, "Graph execution failed")
	}

	logger.Slog.Info("Execution completed successfully", "output", finalOutput)

	shutdownIstioSidecar()
	os.Exit(0)
}

//--------------------- Step Functions ---------------------//

func loadAgentJob(filePath string) (Agent, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Agent{}, err
	}
	var agent Agent
	err = json.Unmarshal(data, &agent)
	logger.Slog.Info("Successfully loaded agentjob", "agent", agent)
	return agent, err
}

func loadNodesLibrary(agentManagerURL string) (NodesLibrary, error) {
	// Step 1: Fetch the basic node list
	nodesListURL := agentManagerURL + "/nodes"
	req, err := http.NewRequest("GET", nodesListURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Ea-Internal", "internal") // Use internal header

	client := &http.Client{}
	resp, err := client.Do(req)
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

		req, err := http.NewRequest("GET", nodeDetailURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-Ea-Internal", "internal") // Use internal header

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch node details: %d", resp.StatusCode)
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

func buildExecutionGraph(agent Agent) (ExecutionGraph, error) {
	graph := ExecutionGraph{
		Nodes:     make(map[string]NodeInstance),
		AdjList:   make(map[string][]string),
		Indegrees: make(map[string]int),
	}

	existingEdges := make(map[string]bool)
	logger.Slog.Info("Building execution graph...")

	// Step 1: Add all nodes
	for _, node := range agent.Nodes {
		graph.Nodes[node.Alias] = node
		graph.Indegrees[node.Alias] = 0
		logger.Slog.Info("Added node to graph", "alias", node.Alias)
	}

	// Step 2: Add explicit edges
	for _, edge := range agent.Edges {
		for _, from := range edge.From {
			for _, to := range edge.To {
				edgeKey := from + "->" + to
				if !existingEdges[edgeKey] {
					graph.AdjList[from] = append(graph.AdjList[from], to)
					graph.Indegrees[to]++
					existingEdges[edgeKey] = true
					logger.Slog.Info("Added explicit edge", "from", from, "to", to, "current_indegree", graph.Indegrees[to])
				}
			}
		}
	}

	// Step 3: Add implicit edges (skip if explicit exists)
	for _, node := range agent.Nodes {
		dependencies := extractParameterDependencies(node.Parameters)
		for _, dep := range dependencies {
			if dep != node.Alias {
				edgeKey := dep + "->" + node.Alias
				if dep != node.Alias && !contains(graph.AdjList[dep], node.Alias) {
					graph.AdjList[dep] = append(graph.AdjList[dep], node.Alias)
					graph.Indegrees[node.Alias]++
					existingEdges[edgeKey] = true
					logger.Slog.Info("Added implicit edge", "from", dep, "to", node.Alias, "current_indegree", graph.Indegrees[node.Alias])
				} else {
					logger.Slog.Warn("Duplicate dependency detected (explicit+implicit) - skipping implicit edge", "from", dep, "to", node.Alias)
				}
			}
		}
	}

	// ✅ Preserve original indegrees before topological sort
	originalIndegrees := make(map[string]int)
	for k, v := range graph.Indegrees {
		originalIndegrees[k] = v
	}
	graph.Indegrees = originalIndegrees

	// Step 4: Topological Sort
	order, err := topologicalSort(graph)
	if err != nil {
		return ExecutionGraph{}, err
	}
	graph.ExecutionOrder = order

	// Step 5: Output the execution graph structure
	logger.Slog.Info("Execution graph built successfully")
	for node, neighbors := range graph.AdjList {
		logger.Slog.Info("Node connections", "node", node, "triggers", neighbors)
	}
	for node, indegree := range graph.Indegrees {
		logger.Slog.Info("Final indegree", "node", node, "indegree", indegree)
	}
	logger.Slog.Info("Execution order", "order", graph.ExecutionOrder)

	return graph, nil
}

func executeGraph(agentJobID string, agentCreator string, graph ExecutionGraph, nodesLib []NodeDefinition) (interface{}, error) {
	state := &ExecutionState{Results: make(map[string]interface{})}
	var wg sync.WaitGroup
	errCh := make(chan error, len(graph.ExecutionOrder))

	executedNodes := make(map[string]bool)
	executedNodesLock := sync.Mutex{}

	logger.Slog.Info("Starting graph execution...")

	// ✅ Channel for nodes ready to be executed
	nodeQueue := make(chan NodeInstance, len(graph.ExecutionOrder))
	activeNodes := 0                // Counter for active nodes
	activeNodesLock := sync.Mutex{} // Lock for activeNodes counter

	// ✅ Seed ONLY nodes with no incoming edges (no dependencies)
	for nodeAlias, indegree := range graph.Indegrees {
		if indegree == 0 && !hasIncomingEdges(graph, nodeAlias) {
			logger.Slog.Info("Seeding node with no dependencies", "node", nodeAlias)
			nodeQueue <- graph.Nodes[nodeAlias]

			activeNodesLock.Lock()
			activeNodes++
			activeNodesLock.Unlock()
		}
	}

	// Worker function to process nodes
	worker := func() {
		for node := range nodeQueue {
			logger.Slog.Info("Executing node", "node", node.Alias)

			if err := executeNode(agentJobID, agentCreator, node, nodesLib, state); err != nil {
				logger.Slog.Error("Node execution failed", "node", node.Alias, "error", err)
				errCh <- err
				return
			}

			executedNodesLock.Lock()
			executedNodes[node.Alias] = true
			logger.Slog.Info("Node execution completed", "node", node.Alias)
			executedNodesLock.Unlock()

			// ✅ Resolve dependencies strictly
			for _, dependent := range graph.AdjList[node.Alias] {
				executedNodesLock.Lock()
				if graph.Indegrees[dependent] > 0 {
					graph.Indegrees[dependent]--
					logger.Slog.Info("Dependency resolved", "from", node.Alias, "to", dependent, "remaining_indegree", graph.Indegrees[dependent])
				}
				logger.Slog.Info("Dependency resolved", "from", node.Alias, "to", dependent, "remaining_indegree", graph.Indegrees[dependent])

				// ✅ Check if ALL dependencies have been successfully executed
				allDepsExecuted := true
				for parent, children := range graph.AdjList {
					if contains(children, dependent) && !executedNodes[parent] {
						allDepsExecuted = false
						logger.Slog.Info("Dependency not yet executed", "dependent", dependent, "missing_dependency", parent)
						break
					}
				}

				if graph.Indegrees[dependent] == 0 && allDepsExecuted {
					logger.Slog.Info("All dependencies satisfied, adding to queue", "node", dependent)
					nodeQueue <- graph.Nodes[dependent]

					activeNodesLock.Lock()
					activeNodes++
					activeNodesLock.Unlock()
				}
				executedNodesLock.Unlock()
			}

			// ✅ Decrement active node count
			activeNodesLock.Lock()
			activeNodes--
			if activeNodes == 0 {
				logger.Slog.Info("All nodes executed, closing the queue")
				close(nodeQueue)
			}
			activeNodesLock.Unlock()
		}
	}

	// Launch worker goroutines
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			logger.Slog.Info("Worker started", "worker_id", workerID)
			worker()
			logger.Slog.Info("Worker finished", "worker_id", workerID)
		}(i)
	}

	wg.Wait()
	close(errCh)

	logger.Slog.Info("Graph execution completed successfully")

	// Error handling
	if len(errCh) > 0 {
		return nil, <-errCh
	}

	// ✅ Return the final output
	finalNode := graph.ExecutionOrder[len(graph.ExecutionOrder)-1]
	logger.Slog.Info("Graph execution completed successfully", "final_node", finalNode, "result", state.Results[finalNode])
	return state.Results[finalNode], nil
}

//--------------------- Node Execution ---------------------//

func executeNode(agentJobID string, agentCreator string, node NodeInstance, nodesLib []NodeDefinition, state *ExecutionState) error {
	nodeDef, err := findNodeDefinition(node.Type, nodesLib)
	if err != nil {
		return err
	}

	logger.Slog.Info("Executing node", "alias", node.Alias, "original_parameters", node.Parameters)

	// Inject inputs from state (resolve {{alias.some.output}})
	params, err := injectInputsFromState(node.Parameters, state)
	if err != nil {
		return err
	}
	node.Parameters = params

	logger.Slog.Info("Parameters after dependency injection", "alias", node.Alias, "injected_parameters", node.Parameters)

	var result interface{}
	if nodeDef.API.BaseURL != "" {
		// Execute API node with injected parameters
		result, err = executeAPINode(agentCreator, node, nodeDef, state)
		if err != nil {
			return err
		}
	} else {
		result = node.Parameters
	}

	// Save raw and flattened output
	state.Lock.Lock()
	state.Results[node.Alias] = result

	logger.Slog.Info("Execution result", "alias", node.Alias, "result", result)

	// Flatten the output for easy reference
	flattened := make(map[string]interface{})
	if resMap, ok := result.(map[string]interface{}); ok {
		flattenJSON(node.Alias, resMap, flattened)
		for k, v := range flattened {
			state.Results[k] = v
		}
	}
	state.Lock.Unlock()

	// ✅ Emit Kubernetes Event with flattened JSON
	flattenedJSON, err := json.Marshal(flattened)
	if err != nil {
		logger.Slog.Error("Failed to marshal flattened output", "node", node.Alias, "error", err)
		return err
	}
	emitK8sEvent(agentJobID, node.Alias, "Completed", string(flattenedJSON))

	return nil
}

func executeAPINode(agentCreator string, node NodeInstance, def NodeDefinition, state *ExecutionState) (interface{}, error) {
	// Inject inputs from state to resolve placeholders
	params, err := injectInputsFromState(node.Parameters, state)
	if err != nil {
		return nil, err
	}
	node.Parameters = params

	// Replace placeholders in URL for GET, DELETE, and PUT methods
	url := def.API.BaseURL + def.API.Endpoint
	if def.API.Method == "GET" || def.API.Method == "DELETE" || def.API.Method == "PUT" || def.API.Method == "POST" {
		for key, value := range node.Parameters {
			placeholder := fmt.Sprintf("{%s}", key)
			if strings.Contains(url, placeholder) {
				url = strings.ReplaceAll(url, placeholder, fmt.Sprintf("%v", value))
				delete(node.Parameters, key) // Remove path params from query/body
			}
		}
	}

	logger.Slog.Info("Preparing API request", "alias", node.Alias, "url", url, "method", def.API.Method, "payload", node.Parameters)

	// Prepare API request
	var req *http.Request
	if def.API.Method == "POST" || def.API.Method == "PUT" {
		// Send JSON payload for POST and PUT
		body, _ := json.Marshal(node.Parameters)
		req, err = http.NewRequest(def.API.Method, url, strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
	} else {
		// For GET and DELETE, append query parameters if any remain
		if len(node.Parameters) > 0 {
			queryParams := make([]string, 0)
			for key, value := range node.Parameters {
				queryParams = append(queryParams, fmt.Sprintf("%s=%v", key, value))
			}
			url = fmt.Sprintf("%s?%s", url, strings.Join(queryParams, "&"))
		}
		req, err = http.NewRequest(def.API.Method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	// Handle Authorization header secret replacement
	for key, value := range def.API.Headers {
		updatedHeader := value
		if strings.Contains(value, "((") && strings.Contains(value, "))") {
			// Extract secret key from placeholder ((some_key))
			secretKey := extractSecretKey(value)
			if secretKey != "" {
				// Fetch secret from Kubernetes
				userSecret, err := fetchUserSecret(agentCreator)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch user secret: %w", err)
				}

				// Replace ((some_key)) with actual secret value
				if apiKey, exists := userSecret[secretKey]; exists {
					updatedHeader = strings.ReplaceAll(value, fmt.Sprintf("((%s))", secretKey), apiKey)
					logger.Slog.Info("Injected secret into Authorization header", "alias", node.Alias, "header_key", key)
				} else {
					return nil, fmt.Errorf("missing required API key: %s", secretKey)
				}
			}
		}
		req.Header.Set(key, updatedHeader)
	}

	// Execute the API call
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	logger.Slog.Info("API response received", "alias", node.Alias, "response", result)

	// Flatten the API response for easy referencing
	state.Lock.Lock()
	flattened := make(map[string]interface{})
	flattenJSON(node.Alias, result, flattened)
	for k, v := range flattened {
		state.Results[k] = v
	}
	state.Lock.Unlock()

	return result, nil
}

//--------------------- Helper Functions ---------------------//

// Extract secret key name from placeholder ((some_key))
func extractSecretKey(headerValue string) string {
	re := regexp.MustCompile(`\(\((.*?)\)\)`)
	matches := re.FindStringSubmatch(headerValue)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// Fetch user's API secret from Kubernetes
func fetchUserSecret(agentCreator string) (map[string]string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create Kubernetes config", "error", err)
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create Kubernetes client", "error", err)
		return nil, err
	}

	// Fetch the user's secret
	secretName := fmt.Sprintf("third-party-user-creds-%s", agentCreator)
	secret, err := clientset.CoreV1().Secrets("ea-platform").Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret: %w", err)
	}

	// Convert secret data from base64
	decodedData := make(map[string]string)
	for key, value := range secret.Data {
		decodedData[key] = string(value)
	}
	return decodedData, nil
}

func handleError(err error, msg string) {
	if err != nil {
		logger.Slog.Error(msg, "error", err)
		os.Exit(1)
	}
}

func findNodeDefinition(nodeType string, nodesLib []NodeDefinition) (NodeDefinition, error) {
	for _, node := range nodesLib {
		if node.Type == nodeType {
			return node, nil
		}
	}
	return NodeDefinition{}, errors.New("node definition not found")
}

func injectInputsFromState(params map[string]interface{}, state *ExecutionState) (map[string]interface{}, error) {
	resolved := make(map[string]interface{})

	// Regex to detect all placeholders like {{ alias.key }}
	placeholderRegex := regexp.MustCompile(`{{\s*([^{}]+?)\s*}}`)

	var resolveValue func(interface{}) (interface{}, error)

	resolveValue = func(value interface{}) (interface{}, error) {
		switch v := value.(type) {
		case string:
			matches := placeholderRegex.FindAllStringSubmatch(v, -1)
			if len(matches) == 0 {
				return v, nil
			}

			resolvedStr := v
			for _, match := range matches {
				ref := match[1] // e.g., "gpt.choices"
				parts := strings.Split(ref, ".")
				alias := parts[0]

				// Lock and fetch value from ExecutionState
				state.Lock.RLock()
				data, exists := state.Results[alias]
				state.Lock.RUnlock()

				if !exists {
					return nil, fmt.Errorf("invalid reference: %s", ref)
				}

				// Resolve nested properties
				result := data
				for _, part := range parts[1:] {
					// Handle array index (e.g., messages[0])
					if strings.Contains(part, "[") {
						key := part[:strings.Index(part, "[")]
						idxStr := part[strings.Index(part, "[")+1 : strings.Index(part, "]")]
						index, err := strconv.Atoi(idxStr)
						if err != nil {
							return nil, fmt.Errorf("invalid index in reference: %s", part)
						}

						// Traverse into the array
						if arr, ok := result.(map[string]interface{})[key].([]interface{}); ok {
							if index >= 0 && index < len(arr) {
								result = arr[index]
							} else {
								return nil, fmt.Errorf("index out of bounds in reference: %s", part)
							}
						} else {
							return nil, fmt.Errorf("expected array for key: %s", key)
						}
					} else {
						// Traverse into nested objects
						if nested, ok := result.(map[string]interface{})[part]; ok {
							result = nested
						} else {
							return nil, fmt.Errorf("invalid nested key: %s", part)
						}
					}
				}

				// Replace the placeholder with the resolved value
				resolvedStr = strings.Replace(resolvedStr, match[0], fmt.Sprintf("%v", result), -1)
			}

			return resolvedStr, nil

		case map[string]interface{}:
			newMap := make(map[string]interface{})
			for k, v := range v {
				newVal, err := resolveValue(v)
				if err != nil {
					return nil, err
				}
				newMap[k] = newVal
			}
			return newMap, nil

		case []interface{}:
			newSlice := make([]interface{}, len(v))
			for i, item := range v {
				newVal, err := resolveValue(item)
				if err != nil {
					return nil, err
				}
				newSlice[i] = newVal
			}
			return newSlice, nil

		default:
			return v, nil
		}
	}

	for key, val := range params {
		newVal, err := resolveValue(val)
		if err != nil {
			return nil, err
		}
		resolved[key] = newVal
	}

	return resolved, nil
}

func topologicalSort(graph ExecutionGraph) ([]string, error) {
	var order []string
	queue := []string{}

	for node, indeg := range graph.Indegrees {
		if indeg == 0 {
			queue = append(queue, node)
		}
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		order = append(order, current)

		for _, neighbor := range graph.AdjList[current] {
			graph.Indegrees[neighbor]--
			if graph.Indegrees[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(order) != len(graph.Nodes) {
		return nil, errors.New("cyclic dependency detected")
	}

	return order, nil
}

func emitK8sEvent(agentJobID string, nodeAlias, status, output string) {
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Slog.Error("Failed to create Kubernetes config", "error", err)
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Slog.Error("Failed to create Kubernetes client", "error", err)
		return
	}

	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-", agentJobID, nodeAlias),
			Namespace:    "ea-platform",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:      "AgentJob",
			Namespace: "ea-platform",
			Name:      agentJobID,
		},
		Reason:  "NodeStatusUpdate",
		Message: fmt.Sprintf("Node %s completed", nodeAlias),
		Type:    "Normal",
	}

	// Add output as annotations
	event.Annotations = map[string]string{
		"nodeAlias": nodeAlias,
		"status":    status,
		"output":    string(output), // JSON string instead of raw map
	}

	_, err = clientset.CoreV1().Events("ea-platform").Create(context.TODO(), event, metav1.CreateOptions{})
	if err != nil {
		logger.Slog.Error("Failed to emit Kubernetes event", "error", err)
	}
}

func extractParameterDependencies(parameters map[string]interface{}) []string {
	var dependencies []string
	seen := make(map[string]bool)

	// ✅ Updated regex to capture multiple placeholders in one string
	placeholderRegex := regexp.MustCompile(`{{\s*([^{}]+?)\s*}}`)

	for key, value := range parameters {
		if strVal, ok := value.(string); ok {
			matches := placeholderRegex.FindAllStringSubmatch(strVal, -1)
			for _, match := range matches {
				fullReference := match[1]                     // e.g., "noaa.properties.periods[0].detailedForecast"
				alias := strings.Split(fullReference, ".")[0] // Extract alias before the first dot

				if !seen[alias] {
					dependencies = append(dependencies, alias)
					seen[alias] = true
					logger.Slog.Info("Detected dependency", "parameter_key", key, "alias", alias, "full_reference", match[0])
				}
			}
		}
	}

	return dependencies
}

func flattenJSON(prefix string, data map[string]interface{}, result map[string]interface{}) {
	for key, value := range data {
		fullKey := prefix + "." + key
		switch v := value.(type) {
		case map[string]interface{}:
			flattenJSON(fullKey, v, result)
		default:
			result[fullKey] = v
		}
	}
}

// Helper function to check if an item exists in a slice
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func hasIncomingEdges(graph ExecutionGraph, nodeAlias string) bool {
	for _, targets := range graph.AdjList {
		for _, target := range targets {
			if target == nodeAlias {
				return true
			}
		}
	}
	return false
}

func shutdownIstioSidecar() {
	resp, err := http.Post("http://localhost:15020/quitquitquit", "application/json", nil)
	if err != nil {
		logger.Slog.Error("Failed to shutdown Istio sidecar", "error", err)
		return
	}
	defer resp.Body.Close()
	log.Println("Istio sidecar shutdown initiated")
}
