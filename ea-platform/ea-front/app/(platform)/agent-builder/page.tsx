"use client";

import { useState, useEffect } from "react";
import NodeLibrary from "../components/NodeLibrary";
import JsonEditor from "../components/JsonEditor";
import WorkflowBuilder from "../components/WorkflowBuilder";
import { Node, Edge } from "reactflow";
import { ChevronDown, ChevronUp } from "lucide-react";

// Function to fetch JWT from `/api/auth/token`
const fetchToken = async () => {
  try {
    const res = await fetch("/api/auth/token", { credentials: "include" });
    if (!res.ok) throw new Error("Failed to fetch token");
    const data = await res.json();
    return data.token;
  } catch (error) {
    console.error("Error fetching token:", error);
    return null;
  }
};

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
  const [agentId, setAgentId] = useState<string | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState<boolean>(false);
  const [runningJobId, setRunningJobId] = useState<string | null>(null);

  // Fetch and set creator ID from JWT token
  useEffect(() => {
    const fetchCreatorId = async () => {
      try {
        const token = await fetchToken();
        if (!token) {
          console.error("No token found.");
          return;
        }

        // Decode JWT payload and extract the `iss` field (which is the userId)
        const payload = JSON.parse(atob(token.split(".")[1]));
        if (payload.iss) {
          setCreator(payload.iss); // ✅ Extract UUID from `iss` field
        }
      } catch (error) {
        console.error("Error fetching or decoding token:", error);
      }
    };

    fetchCreatorId();
  }, []);

  // Generate JSON representation of the workflow
  const generateJsonSchema = (): string => {
    const nodeAliasMap = new Map(workflowNodes.map(node => [node.id, node.data.alias]));

    return JSON.stringify(
      {
        name: agentName,
        creator: creator,
        description: agentDescription,
        id: agentId,
        nodes: workflowNodes.map((node) => ({
          alias: node.data.alias ?? node.id,
          type: node.data.type,
          parameters: node.data.parametersState || {},
          position: node.position,
        })),
        edges: workflowEdges.map((edge) => ({
          from: [nodeAliasMap.get(edge.source) ?? edge.source],
          to: [nodeAliasMap.get(edge.target) ?? edge.target],
        })),
      },
      null,
      2
    );
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
    setAgentId(id);
  };

  return (
    <div className="relative flex min-h-screen bg-neutral-900 text-white overflow-hidden">
      
      {/* ✅ Title Bar Hack - Only Extends Across Textbox */}
      <div 
        className="absolute top-0 left-4 bg-neutral-900 p-3 flex items-center shadow-md z-50 rounded-lg pointer-events-auto transition-all duration-300"
        style={{ maxWidth: "400px" }}
      >
        <input
          type="text"
          value={agentName}
          onChange={(e) => setAgentName(e.target.value)}
          className="bg-transparent text-lg font-semibold text-white text-center outline-none w-64 px-4 py-2 rounded-md border border-neutral-600 focus:border-blue-500 transition duration-150 pointer-events-auto"
          placeholder="Enter Agent Name..."
        />
        <button
          onClick={() => setDescOpen(!descOpen)}
          className="ml-2 p-2 rounded-lg bg-neutral-800 hover:bg-neutral-700 transition pointer-events-auto"
        >
          {descOpen ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
        </button>
      </div>

      {/* ✅ Description Box */}
      {descOpen && (
        <div className="absolute top-14 left-4 bg-neutral-800 p-3 rounded-lg shadow-lg z-50 w-96 pointer-events-auto">
          <textarea
            value={agentDescription}
            onChange={(e) => setAgentDescription(e.target.value)}
            className="bg-neutral-700 text-white text-sm p-3 w-full rounded-lg outline-none focus:ring-2 focus:ring-blue-500 pointer-events-auto"
            placeholder="Enter Agent Description..."
          />
        </div>
      )}

      {/* ✅ Workflow Builder */}
      <div className={`flex-1 flex flex-col relative transition-all duration-300`}>
        <WorkflowBuilder
          nodes={workflowNodes}
          setNodes={setWorkflowNodes}
          edges={workflowEdges}
          setEdges={setWorkflowEdges}
          setJsonText={setJsonText}
          agentId={agentId}
          creatorId={creator}
          runningJobId={runningJobId}
          setRunningJobId={setRunningJobId}
          sidebarOpen={sidebarOpen} 
        />
      </div>

      {/* ✅ Node Library */}
      <NodeLibrary sidebarOpen={sidebarOpen} addNodeToFlow={addNodeToFlow} />

      {/* ✅ JSON Editor */}
      <div
        className="absolute top-0 right-0 h-full bg-neutral-900 transition-all duration-300 z-40"
        style={{ width: jsonEditorOpen ? jsonEditorWidth : "0" }}
      >
        <JsonEditor
          isOpen={jsonEditorOpen}
          toggle={() => setJsonEditorOpen(!jsonEditorOpen)}
          jsonText={jsonText}
          onJsonChange={handleJsonChange}
          agentId={agentId}
          updateAgentId={updateAgentId}
          creatorId={creator}
          onJobStarted={(jobId) => setRunningJobId(jobId)} 
        />
      </div>
    </div>
  );
}
