package operator

// Define structs to store agent definition after lookup
type Agent struct {
	ID          string `json:"_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	User        string `json:"user"`
	Nodes       []Node `json:"nodes"`
	Edges       []Edge `json:"edges"`
}

type Node struct {
	ID            string                 `json:"id"`
	DefinitionRef string                 `json:"definition_ref"`
	Parameters    map[string]interface{} `json:"parameters"`
}

type Edge struct {
	From []string `json:"from"`
	To   []string `json:"to"`
}

type CreateJobRequest struct {
	AgentID string `json:"agentID"`
}
