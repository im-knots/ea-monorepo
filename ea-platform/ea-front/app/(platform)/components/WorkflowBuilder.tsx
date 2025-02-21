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
import 'reactflow/dist/style.css';

interface WorkflowBuilderProps {
  nodes: Node[];
  setNodes: React.Dispatch<React.SetStateAction<Node[]>>;
  edges: Edge[];
  setEdges: React.Dispatch<React.SetStateAction<Edge[]>>;
  setJsonText: (json: string) => void;
  agentId: string | null;  // Add agentId here
}

const nodeTypes = { custom: CustomNode };

export default function WorkflowBuilder({
  nodes,
  setNodes,
  edges,
  setEdges,
  setJsonText,
  agentId,  // Receive agentId
}: WorkflowBuilderProps) {
  // Handle node changes
  const onNodesChange = useCallback(
    (changes: NodeChange[]) => {
      setNodes((nds: Node[]) => applyNodeChanges(changes, nds));
    },
    [setNodes]
  );

  // Handle edge changes
  const onEdgesChange = useCallback(
    (changes: EdgeChange[]) => {
      setEdges((eds: Edge[]) => applyEdgeChanges(changes, eds));
    },
    [setEdges]
  );

  const onConnect = useCallback(
    (connection: Connection) => {
      setEdges((eds: Edge[]) => addEdge(connection, eds));
    },
    [setEdges]
  );

  useEffect(() => {
    const formattedJson = JSON.stringify(
      {
        name: "My Sample Workflow",
        creator: "<UUID OF CREATOR USER>",
        description: "An automatically generated workflow.",
        nodes: nodes.map((node: Node) => ({
          alias: node.data.alias ?? node.id,
          type: node.data.type,
          parameters: node.data.parametersState || {},
        })),
        edges: edges.map((edge: Edge) => ({
          from: [edge.source],
          to: [edge.target],
        })),
        id: agentId, // Include agentId here if it's available
      },
      null,
      2
    );

    setJsonText(formattedJson);
  }, [nodes, edges, setJsonText, agentId]);  // Depend on agentId

  // Set default node positions if not set
  const initializedNodes = nodes.map((node) => ({
    ...node,
    position: node.position || { x: Math.random() * 400, y: Math.random() * 400 },  // Ensure position
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
                  parametersState: key !== "alias" ? { ...node.data.parametersState, [key]: value } : node.data.parametersState,
                },
              }
            : node
        )
      );
    },
    [setNodes]
  );
  

  return (
    <div className="relative flex-1 w-full h-full bg-neutral-900">
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
        style={{ height: '100%', width: '100%' }}  // Ensure full height
        >
        <Background />
        </ReactFlow>
    </div>
  );
}
