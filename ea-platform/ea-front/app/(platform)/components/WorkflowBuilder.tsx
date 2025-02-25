"use client";

import React, { useCallback, useEffect, useState } from "react";
import ReactFlow, {
  Background,
  applyNodeChanges,
  applyEdgeChanges,
  addEdge,
  NodeChange,
  EdgeChange,
  Connection,
  Edge,
  Node,
} from "reactflow";
import CustomNode from "./CustomNode";

// Import necessary styles for React Flow
import "reactflow/dist/style.css";

// API URL for job status
const AINU_MANAGER_URL = "http://api.ea.erulabs.local/ainu-manager/api/v1";

interface JobNode {
  alias: string;
  output?: string;
  status?: string;
}

interface Job {
  job_name: string;
  status: string;
  nodes: JobNode[];
}

interface WorkflowBuilderProps {
  nodes: Node[];
  setNodes: React.Dispatch<React.SetStateAction<Node[]>>;
  edges: Edge[];
  setEdges: React.Dispatch<React.SetStateAction<Edge[]>>;
  setJsonText: (json: string) => void;
  agentId: string | null;
  creatorId: string;
  runningJobId: string | null;
  setRunningJobId: (jobId: string | null) => void;
  sidebarOpen: boolean;
}

const nodeTypes = { custom: CustomNode };

export default function WorkflowBuilder({
  nodes,
  setNodes,
  edges,
  setEdges,
  setJsonText,
  agentId,
  creatorId,
  runningJobId,
  setRunningJobId,
  sidebarOpen,
}: WorkflowBuilderProps) {
  const [token, setToken] = useState<string | null>(null);

  // Fetch JWT token from the server
  const fetchToken = async () => {
    try {
      const res = await fetch("/api/auth/token", { credentials: "include" });
      if (!res.ok) throw new Error("Failed to fetch token");
      const data = await res.json();
      setToken(data.token);
    } catch (error) {
      console.error("Error fetching token:", error);
      setToken(null);
    }
  };

  // 🆕 Function to reset node status and outputs when a new job starts
  const resetNodes = useCallback(() => {
    setNodes((prevNodes) =>
      prevNodes.map((node) => ({
        ...node,
        data: {
          ...node.data,
          status: "Idle", // Reset status
          outputs: {}, // Clear outputs
        },
      }))
    );
  }, [setNodes]);

  // Handle node changes
  const onNodesChange = useCallback(
    (changes: NodeChange[]) => {
      setNodes((nds) => applyNodeChanges(changes, nds));
    },
    [setNodes]
  );

  // Handle edge changes
  const onEdgesChange = useCallback(
    (changes: EdgeChange[]) => {
      setEdges((eds) => applyEdgeChanges(changes, eds));
    },
    [setEdges]
  );

  const onConnect = useCallback(
    (connection: Connection) => {
      setEdges((eds) => addEdge(connection, eds));
    },
    [setEdges]
  );

  // Fetch job status periodically
  const fetchJobStatus = useCallback(async () => {
    if (!creatorId || !runningJobId || !token) return;

    try {
      const response = await fetch(`${AINU_MANAGER_URL}/users/${creatorId}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`, // Attach JWT token
        },
        credentials: "include",
      });

      if (!response.ok) throw new Error(`Failed to fetch job status: ${response.status}`);
      const data = await response.json();

      if (data.jobs) {
        const job: Job | undefined = data.jobs.find((j: Job) => j.job_name === runningJobId);

        if (job) {
          console.log("Job status:", job);

          setNodes((prevNodes) =>
            prevNodes.map((node) => {
              const nodeStatus: JobNode | undefined = job.nodes.find((n) => n.alias === node.data.alias);

              if (nodeStatus) {
                let parsedOutput: Record<string, any> = {};
                try {
                  parsedOutput = nodeStatus.output ? JSON.parse(nodeStatus.output) : {};
                } catch (error) {
                  console.error(`Error parsing output for node ${nodeStatus.alias}:`, error);
                }

                return {
                  ...node,
                  data: {
                    ...node.data,
                    status: nodeStatus.status, // ✅ Use the node's own `status`
                    outputs: parsedOutput, // ✅ Store parsed outputs
                  },
                };
              }
              return node;
            })
          );

          // ✅ Stop polling if the overall job is complete
          if (job.status.toLowerCase() === "completed") {
            console.log("Job finished:", job.status);
            setRunningJobId(null);
          }
        }
      }
    } catch (error) {
      console.error("Error fetching job status:", error);
    }
  }, [creatorId, runningJobId, token, setNodes, setRunningJobId]);

  // 🔄 **Reintroduce polling every 5 seconds when a job is running**
  useEffect(() => {
    if (runningJobId) {
      fetchJobStatus(); // Fetch immediately
      const interval = setInterval(fetchJobStatus, 5000);
      return () => clearInterval(interval); // Cleanup on unmount
    }
  }, [runningJobId, fetchJobStatus]);

  // 🔥 **Reset nodes when a new job starts**
  useEffect(() => {
    if (runningJobId) {
      console.log("Starting new job, resetting node statuses...");
      resetNodes();
    }
  }, [runningJobId, resetNodes]);

  // Fetch token on component mount
  useEffect(() => {
    fetchToken();
  }, []);

  // Generate JSON representation of the workflow
  useEffect(() => {
    const formattedJson = JSON.stringify(
      {
        name: "My Sample Workflow",
        creator: creatorId,
        description: "An automatically generated workflow.",
        nodes: nodes.map((node) => ({
          alias: node.data.alias ?? node.id,
          type: node.data.type,
          parameters: node.data.parametersState || {},
        })),
        edges: edges.map((edge) => ({
          from: [edge.source],
          to: [edge.target],
        })),
        id: agentId,
      },
      null,
      2
    );

    setJsonText(formattedJson);
  }, [nodes, edges, setJsonText, agentId, creatorId]);

  // Ensure nodes have default positions
  const initializedNodes = nodes.map((node) => ({
    ...node,
    position: node.position || { x: Math.random() * 400, y: Math.random() * 400 },
  }));

  // Function to update node data
  const updateNodeData = useCallback(
    (id: string, key: string, value: any) => {
      setNodes((nodes) =>
        nodes.map((node) =>
          node.id === id
            ? {
                ...node,
                data: {
                  ...node.data,
                  [key]: key === "alias" ? value : node.data[key],
                  parametersState:
                    key !== "alias"
                      ? { ...node.data.parametersState, [key]: value }
                      : node.data.parametersState,
                },
              }
            : node
        )
      );
    },
    [setNodes]
  );

  return (
    <div 
      className="relative flex-1 h-full bg-neutral-900 transition-all duration-300"
      style={{ 
        marginTop: "70px" // ✅ Ensure it moves below the header bar
      }} 
    >
      <ReactFlow
        nodes={initializedNodes.map((node) => ({
          ...node,
          data: {
            ...node.data,
            updateNodeData, // Pass the updateNodeData function
          },
        }))}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        fitView
        nodeTypes={nodeTypes}
        style={{ height: "100%", width: "100%" }}
      >
        <Background />
      </ReactFlow>
    </div>
  );
}
