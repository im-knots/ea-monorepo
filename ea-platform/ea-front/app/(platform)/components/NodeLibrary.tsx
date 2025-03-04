"use client";

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { ChevronUp, ChevronDown } from "lucide-react";
import { Node } from "reactflow";

interface NodeDefinition {
  id: string;
  name: string;
  type: string;
  parameters: { key: string; type: string; default: any; enum?: string[] }[];
  metadata?: { description?: string };
}

const API_BASE_URL = "http://api.ea.erulabs.local/agent-manager/api/v1/nodes";

const NODE_TYPES = ["all", "input", "worker", "destination", "utils"];

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
  const [selectedType, setSelectedType] = useState<string>("all");

  // Fetch JWT token from the server
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

  useEffect(() => {
    const fetchNodes = async () => {
      try {
        const token = await fetchToken();
        if (!token) {
          console.error("No token found.");
          setError("Authentication error: No token found.");
          setLoading(false);
          return;
        }

        // Fetch list of node IDs
        const response = await fetch(API_BASE_URL, {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`,
          },
          credentials: "include",
        });

        if (!response.ok) throw new Error(`Failed to fetch node IDs: ${response.status}`);
        const nodeList: { id: string }[] = await response.json();

        // Fetch node details in parallel
        const nodeDetails = await Promise.all(
          nodeList.map(async (node) => {
            try {
              const nodeResponse = await fetch(`${API_BASE_URL}/${node.id}`, {
                method: "GET",
                headers: {
                  "Content-Type": "application/json",
                  "Authorization": `Bearer ${token}`,
                },
                credentials: "include",
              });

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

  // Filter nodes based on selected type
  const filteredNodes =
    selectedType === "all"
      ? nodes
      : nodes.filter((node) => node.type.startsWith(selectedType));

  return (
    <div
      className="absolute bottom-0 bg-neutral-900 text-white shadow-2xl transition-all duration-300"
      style={{
        width: `calc(100% - ${sidebarOpen ? "16rem" : "4rem"})`,
      }}
    >
      {/* Node Library Toggle Button (Always Visible) */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-center py-3 bg-neutral-800 hover:bg-neutral-700 transition"
      >
        {isOpen ? <ChevronDown size={20} /> : <ChevronUp size={20} />}
        <span className="ml-2 text-sm font-medium">
          {isOpen ? "Close Node Library" : "Open Node Library"}
        </span>
      </button>

      {/* Collapsible Content (Sorting + Nodes) */}
      <motion.div
        initial={{ height: 0, opacity: 0 }}
        animate={{ height: isOpen ? "auto" : 0, opacity: isOpen ? 1 : 0 }}
        className="overflow-hidden bg-neutral-900 border-t border-neutral-700"
      >
        <div className="flex">
          {/* Sidebar Sorting Feature (Collapses with Library) */}
          <div className="w-40 bg-neutral-800 p-4 border-r border-neutral-700">
            <h3 className="text-sm font-semibold text-gray-300 mb-2">Filter by Type</h3>
            <ul className="space-y-2">
              {NODE_TYPES.map((type) => (
                <li key={type}>
                  <button
                    onClick={() => setSelectedType(type)}
                    className={`w-full text-left px-3 py-2 rounded-md transition ${
                      selectedType === type ? "bg-green-500 text-white" : "bg-neutral-700 hover:bg-neutral-600"
                    }`}
                  >
                    {type.charAt(0).toUpperCase() + type.slice(1)}
                  </button>
                </li>
              ))}
            </ul>
          </div>

          {/* Node List */}
          <div className="flex-1 p-4 grid grid-cols-2 md:grid-cols-4 gap-4">
            {loading && <p className="text-gray-400 text-center">Loading nodes...</p>}
            {error && <p className="text-red-400 text-center">{error}</p>}
            {!loading &&
              !error &&
              filteredNodes.map((node) => (
                <div
                  key={node.id}
                  className="bg-neutral-800 p-4 rounded-lg shadow-md hover:bg-neutral-700 transition cursor-pointer"
                  onClick={() => {
                    const uniqueId = `${node.id}-${Math.random().toString(36).substr(2, 9)}`;
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
                    addNodeToFlow(newNode);
                  }}
                >
                  <h3 className="text-base font-semibold">{node.name}</h3>
                  <p className="text-xs text-neutral-400">{node.metadata?.description || "No description."}</p>
                </div>
              ))}
          </div>
        </div>
      </motion.div>
    </div>
  );
}
