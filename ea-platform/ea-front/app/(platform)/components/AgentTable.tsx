"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import AgentRow from "./AgentRow";

const AGENT_MANAGER_URL = "http://api.ea.erulabs.local/agent-manager/api/v1/agents";

export default function AgentTable({ userId }: { userId: string | null }) {
  const router = useRouter();
  const [agents, setAgents] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [token, setToken] = useState<string | null>(null);

  // 🔹 Fetch JWT token from /api/auth/token
  useEffect(() => {
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

    fetchToken();
  }, []);

  // ✅ Fetch all agent IDs, then fetch full agent details
  const fetchAgents = async () => {
    if (!token) {
      console.error("No token available for authentication.");
      return;
    }

    setLoading(true);
    try {
      // 🔥 Step 1: Get basic agent list (IDs)
      const response = await fetch(AGENT_MANAGER_URL, {
        method: "GET",
        headers: {
          "Authorization": `Bearer ${token}`, // ✅ Attach JWT token
          "Content-Type": "application/json",
        },
        credentials: "include",
      });

      if (!response.ok) throw new Error("Failed to fetch agents");

      const agentList = await response.json();
      if (!Array.isArray(agentList)) throw new Error("Invalid agent data received");

      // 🔥 Step 2: Fetch full details for each agent
      const detailedAgents = await Promise.all(
        agentList.map(async (agent: any) => {
          try {
            const detailsRes = await fetch(`${AGENT_MANAGER_URL}/${agent.id}`, {
              method: "GET",
              headers: {
                "Authorization": `Bearer ${token}`, // ✅ Attach JWT token
                "Content-Type": "application/json",
              },
              credentials: "include",
            });

            if (!detailsRes.ok) throw new Error(`Failed to fetch details for agent ${agent.id}`);
            const details = await detailsRes.json();
            return {
              ...agent, // Keep original data
              nodes: details.nodes ?? [], // Ensure nodes exist
              edges: details.edges ?? [], // Ensure edges exist
            };
          } catch (error) {
            console.error(error);
            return null; // Ignore failed agents
          }
        })
      );

      // 🔥 Step 3: Remove failed requests (null values)
      setAgents(detailedAgents.filter(Boolean));
    } catch (error) {
      console.error("Error fetching agents:", error);
    } finally {
      setLoading(false);
    }
  };

  // ✅ Fetch agents on component mount
  useEffect(() => {
    if (userId && token) fetchAgents();
  }, [userId, token]);

  // ✅ Refresh agent list after deletion
  const refreshAgents = () => {
    fetchAgents();
  };

  return (
    <div className="bg-neutral-900 text-white p-6 rounded-lg shadow-md">
        <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold">My Agents</h2>
            
            {/* Button container to align them together on the right */}
            <div className="flex space-x-2 ml-auto">
            <button
                onClick={() => router.push("/agent-builder")}
                className="flex items-center space-x-2 bg-neutral-800 hover:bg-neutral-900 text-white font-semibold px-4 py-2 rounded-md transition"
            >
                <span className="text-lg">+</span>
                <span>Create Agent</span>
            </button>

            <button
                onClick={() => router.push("/node-builder")}
                className="flex items-center space-x-2 bg-neutral-800 hover:bg-neutral-900 text-white font-semibold px-4 py-2 rounded-md transition"
            >
                <span className="text-lg">+</span>
                <span>Create Node</span>
            </button>
            </div>
        </div>

      <div className="overflow-x-auto">
        <table className="w-full border border-gray-700 rounded-lg text-sm">
          <thead className="bg-neutral-800 text-gray-300 uppercase">
            <tr>
              <th className="px-4 py-3 text-left">Agent Name</th>
              <th className="px-4 py-3 text-center">Nodes</th>
              <th className="px-4 py-3 text-center">Jobs</th>
              <th className="px-4 py-3 text-center">Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={4} className="text-center text-gray-400 py-4">
                  Loading agents...
                </td>
              </tr>
            ) : agents.length > 0 ? (
              agents.map((agent) => (
                <AgentRow
                  key={agent.id}
                  agent={agent}
                  userId={userId}
                  refreshAgents={refreshAgents} // ✅ Ensure proper refresh
                />
              ))
            ) : (
              <tr>
                <td colSpan={4} className="text-center text-gray-400 py-4">
                  No agents found.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
