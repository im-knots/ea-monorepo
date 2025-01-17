package handlers

import (
	"net/http"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello World!"))
}

// HandleGetNodes will pull a given Agents component nodes
// TODO: Define agent builder node type schema or declarative language?
func HandleGetNode(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("heres your list of agent builder node types! :)"))
}

// HandleGetPresets will populate a list of presets or a specific preset Agent and its associated component nodes
func HandleGetPresets(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Heres your list of Agent Presets and their component nodes!"))
}

// HandleCreateAgent will create an Agent container for nodes
func HandleCreateAgent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("an Agent!"))
}

// HandleCreateNode will create a Node component in a given Agent
func HandleCreateAgentNode(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("an Agent Node!"))
}

// HandleCreateJob will create a Job in a given Agent
func HandleCreateAgentJob(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("an Agent Job!"))
}
