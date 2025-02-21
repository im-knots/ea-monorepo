"use client";

import { useState, useEffect } from "react";
import NodeLibrary from "../components/NodeLibrary";
import JsonEditor from "../components/JsonEditor";
import WorkflowBuilder from "../components/WorkflowBuilder";
import { Node, Edge } from "reactflow";
import { ChevronDown, ChevronUp } from "lucide-react";

// API URL
const AINU_MANAGER_URL = "http://ainu-manager.ea.erulabs.local/api/v1";

export default function AgentBuilderPage() {
  const [jsonEditorOpen, setJsonEditorOpen] = useState(false);
  const jsonEditorWidth = jsonEditorOpen ? "20rem" : "4rem";
  const [workflowNodes, setWorkflowNodes] = useState<Node[]>([]);
  const [workflowEdges, setWorkflowEdges] = useState<Edge[]>([]);
  const [jsonText, setJsonText] = useState("");
  const [agentName, setAgentName] = useState("My Agent");
  const [agentDescription, setAgentDescription] = useState("An awesome AI agent");
  const [creator, setCreator] = useState("");  
  const [descOpen, setDescOpen] = useState(false);
  const [agentId, setAgentId] = useState<string | null>(null); // Updated state for agentId
  const [sidebarOpen, setSidebarOpen] = useState<boolean>(false); // Local state for sidebarOpen

  // Fetch first user from API and update the creator field
  const fetchCreatorId = async () => {
    try {
      const response = await fetch(`${AINU_MANAGER_URL}/users`);
      const users = await response.json();
      if (users.length > 0) {
        setCreator(users[0].id); // Set the first user's ID
      }
    } catch (error) {
      console.error("Error fetching users:", error);
    }
  };

  useEffect(() => {
    fetchCreatorId();
  }, []);

  // Define the shape of the agent job schema
  const generateJsonSchema = (): string => {
    const nodeAliasMap = new Map(workflowNodes.map(node => [node.id, node.data.alias]));

    const agentJob = {
      name: agentName,
      creator: creator,  // Use the creator state here
      description: agentDescription,
      id: agentId,
      nodes: workflowNodes.map((node) => ({
        alias: node.data.alias ?? node.id,
        type: node.data.type,
        parameters: node.data.parametersState || {},
      })),
      edges: workflowEdges.map((edge) => ({
        from: [nodeAliasMap.get(edge.source) ?? edge.source],
        to: [nodeAliasMap.get(edge.target) ?? edge.target],
      })),
    };

    if (agentId) {
      agentJob.id = agentId;  // Include agentId in the JSON if it exists
    }

    return JSON.stringify(agentJob, null, 2);
  };

  useEffect(() => {
    setJsonText(generateJsonSchema());
  }, [workflowNodes, workflowEdges, agentName, agentDescription, creator, agentId]);

  const handleJsonChange = (json: string) => {
    setJsonText(json);
  };

  const addNodeToFlow = (node: Node) => {
    setWorkflowNodes((prev) => {
      const newNodes = [...prev, node];
      setJsonText(generateJsonSchema());
      return newNodes;
    });
  };

  const updateAgentId = (id: string) => {
    setAgentId(id);  // Update the agentId when a new agent is created or updated
  };

  return (
    <div className="relative flex min-h-screen bg-neutral-950 text-white">
      {/* Title Bar with Agent Name & Description */}
      <div className="absolute top-0 left-0 w-full bg-neutral-900 p-3 flex items-center shadow-md">
        <div className="ml-4 flex items-center space-x-2">
          <input
            type="text"
            value={agentName}
            onChange={(e) => setAgentName(e.target.value)}
            className="bg-transparent text-lg font-semibold text-white text-center outline-none w-80 px-4 py-2 rounded-md border border-neutral-600 focus:border-blue-500 transition duration-150"
            placeholder="Enter Agent Name..."
          />
          <button
            onClick={() => setDescOpen(!descOpen)}
            className="p-2 rounded-lg bg-neutral-800 hover:bg-neutral-700 transition"
          >
            {descOpen ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
          </button>
        </div>
      </div>

      {descOpen && (
        <div className="absolute top-14 left-4 bg-neutral-800 p-3 rounded-lg shadow-lg z-50 w-96">
          <textarea
            value={agentDescription}
            onChange={(e) => setAgentDescription(e.target.value)}
            className="bg-neutral-700 text-white text-sm p-3 w-full rounded-lg outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Enter Agent Description..."
          />
        </div>
      )}

      <div className="flex-1 flex flex-col relative" style={{ marginRight: jsonEditorWidth, marginTop: "50px" }}>
        <div className="flex-1 flex items-center justify-center">
          <WorkflowBuilder
            nodes={workflowNodes}
            setNodes={setWorkflowNodes}
            edges={workflowEdges}
            setEdges={setWorkflowEdges}
            setJsonText={setJsonText} 
            agentId={agentId} // Pass the current agentId to the workflow builder
          />
        </div>
      </div>

      <NodeLibrary sidebarOpen={sidebarOpen} addNodeToFlow={addNodeToFlow} />
      <JsonEditor
        isOpen={jsonEditorOpen}
        toggle={() => setJsonEditorOpen(!jsonEditorOpen)}
        jsonText={jsonText}
        onJsonChange={handleJsonChange}
        agentId={agentId} 
        updateAgentId={updateAgentId}  // Pass the updateAgentId callback to JsonEditor
        creatorId={creator}  // Pass the creator ID to JsonEditor
      />
    </div>
  );
}
