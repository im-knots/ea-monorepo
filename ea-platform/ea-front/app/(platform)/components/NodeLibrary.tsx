"use client";

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { ChevronUp, ChevronDown } from "lucide-react";
import { Node } from "reactflow";

interface NodeDefinition {
  id: string;
  name: string;
  type: string; // ✅ Ensure this gets passed to WorkflowBuilder
  parameters: { key: string; type: string; default: any; enum?: string[] }[];
  metadata?: { description?: string };
}

const API_BASE_URL = "http://agent-manager.ea.erulabs.local/api/v1/nodes";

export default function NodeLibrary({
    sidebarOpen,
    addNodeToFlow,
  }: {
    sidebarOpen: boolean;
    addNodeToFlow: (node: Node) => void;
  }) {
  const [isOpen, setIsOpen] = useState(true);
  const [nodes, setNodes] = useState<NodeDefinition[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchNodes = async () => {
      try {
        const response = await fetch(API_BASE_URL);
        if (!response.ok) throw new Error(`Failed to fetch node IDs: ${response.status}`);
        const nodeList: { id: string }[] = await response.json();

        const nodeDetails = await Promise.all(
          nodeList.map(async (node) => {
            try {
              const nodeResponse = await fetch(`${API_BASE_URL}/${node.id}`);
              if (!nodeResponse.ok) throw new Error(`Failed to fetch node ${node.id}`);
              return (await nodeResponse.json()) as NodeDefinition;
            } catch (error) {
              console.error(`Error fetching details for node ${node.id}:`, error);
              return null;
            }
          })
        );

        setNodes(nodeDetails.filter((n): n is NodeDefinition => n !== null));
      } catch (err) {
        setError(err instanceof Error ? err.message : "An unknown error occurred.");
      } finally {
        setLoading(false);
      }
    };

    fetchNodes();
  }, []);

  return (
    <div
      className="absolute bottom-0 bg-neutral-900 text-white shadow-2xl transition-all"
      style={{
        left: sidebarOpen ? "15rem" : "4rem", // Adjust the left based on sidebar open state
        width: `calc(100% - ${sidebarOpen ? "15rem" : "4rem"})`, // Ensure the width is calculated dynamically based on sidebar width
        bottom: "0",
      }}
    >
      {/* Node Library Toggle Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-center py-3 bg-neutral-800 hover:bg-neutral-700 transition"
      >
        {isOpen ? <ChevronDown size={20} /> : <ChevronUp size={20} />}
        <span className="ml-2 text-sm font-medium">
          {isOpen ? "Close Node Library" : "Open Node Library"}
        </span>
      </button>

      {/* Node List */}
      <motion.div
        initial={{ height: 0, opacity: 0 }}
        animate={{ height: isOpen ? "auto" : 0, opacity: isOpen ? 1 : 0 }}
        className="overflow-hidden bg-neutral-900 border-t border-neutral-700"
      >
        <div className="p-4 grid grid-cols-2 md:grid-cols-4 gap-4">
          {loading && <p className="text-gray-400 text-center">Loading nodes...</p>}
          {error && <p className="text-red-400 text-center">{error}</p>}
          {!loading &&
            !error &&
            nodes.map((node) => (
              <div
                key={node.id}
                className="bg-neutral-800 p-4 rounded-lg shadow-md hover:bg-neutral-700 transition cursor-pointer"
                onClick={() => {
                  const uniqueId = `${node.id}-${Math.random().toString(36).substr(2, 9)}`; // Generate unique ID
                  const newNode: Node = {
                    id: uniqueId,
                    type: "custom",
                    position: { x: Math.random() * 400, y: Math.random() * 400 },
                    data: {
                      alias: uniqueId,
                      type: node.type,
                      label: node.name,
                      parameters: node.parameters,
                      parametersState: node.parameters.reduce<Record<string, any>>(
                        (acc, param) => {
                          acc[param.key] = param.default ?? "";
                          return acc;
                        },
                        {}
                      ),
                    },
                  };
                  addNodeToFlow(newNode); // Call addNodeToFlow from props
                }}
              >
                <h3 className="text-base font-semibold">{node.name}</h3>
                <p className="text-xs text-neutral-400">{node.metadata?.description || "No description."}</p>
              </div>
            ))}
        </div>
      </motion.div>
    </div>
  );
}
