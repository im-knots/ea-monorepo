"use client";

import { useState, useEffect } from "react";
import JobList from "./JobList";

const AGENT_MANAGER_URL = "http://api.ea.erulabs.local/agent-manager/api/v1/agents";
const EA_JOB_API_URL = "http://api.ea.erulabs.local/job-api/api/v1";

interface Node {
  alias: string;
  type: string;
}

interface Agent {
  id: string;
  name: string;
  nodes: Node[];
}

export default function AgentRow({
  agent,
  userId,
  refreshAgents, // ✅ Ensure full refresh of agents
}: {
  agent: Agent;
  userId: string | null;
  refreshAgents: () => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [jobs, setJobs] = useState<any[]>([]);
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

  // ✅ DELETE Agent & Refresh Full List
  const deleteAgent = async () => {
    if (!token) {
      console.error("No token available for authentication.");
      return;
    }

    setIsDeleting(true);
    try {
      const response = await fetch(`${AGENT_MANAGER_URL}/${agent.id}`, {
        method: "DELETE",
        headers: {
          "Authorization": `Bearer ${token}`, // ✅ Attach JWT token for authentication
        },
      });

      if (!response.ok) throw new Error("Failed to delete agent");

      refreshAgents(); // ✅ Trigger full refresh
    } catch (error) {
      console.error("Error deleting agent:", error);
    } finally {
      setIsDeleting(false);
    }
  };

  // ✅ Start Agent Job
  const startAgentJob = async () => {
    if (!userId || !token) {
      console.error("Missing userId or token.");
      return;
    }

    setIsStarting(true);
    try {
      const response = await fetch(`${EA_JOB_API_URL}/jobs`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`, // ✅ Attach JWT token for authentication
        },
        body: JSON.stringify({
          agent_id: agent.id,
          user_id: userId,
        }),
      });

      if (!response.ok) throw new Error("Failed to start agent job");

      setIsStarting(false);
      setExpanded(true);
    } catch (error) {
      console.error("Error starting agent job:", error);
      setIsStarting(false);
    }
  };

  return (
    <>
      <tr
        className="border-b border-gray-700 hover:bg-neutral-800 transition duration-200 cursor-pointer"
        onClick={() => setExpanded((prev) => !prev)}
      >
        <td className="px-4 py-3">{agent.name}</td>
        <td className="px-4 py-3 text-center">{agent.nodes.length}</td>
        <td className="px-4 py-3 text-center">{jobs.length}</td>
        <td className="px-4 py-3 text-center flex justify-center space-x-2">
          {/* Start Button */}
          <button
            onClick={(e) => {
              e.stopPropagation();
              startAgentJob();
            }}
            className={`px-3 py-1 rounded-md text-sm transition ${
              isStarting ? "bg-gray-600 cursor-not-allowed" : "bg-green-600 hover:bg-green-700"
            } text-white`}
            disabled={isStarting}
          >
            {isStarting ? "Starting..." : "Start"}
          </button>

          {/* Modify Button */}
          <button className="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded-md text-sm transition">
            Modify
          </button>

          {/* Delete Button */}
          <button
            onClick={(e) => {
              e.stopPropagation();
              deleteAgent();
            }}
            className={`px-3 py-1 rounded-md text-sm transition ${
              isDeleting ? "bg-gray-600 cursor-not-allowed" : "bg-red-600 hover:bg-red-700"
            } text-white`}
            disabled={isDeleting}
          >
            {isDeleting ? "Deleting..." : "Delete"}
          </button>
        </td>
      </tr>

      {expanded && (
        <tr>
          <td colSpan={4} className="bg-neutral-900 p-4 border-t border-gray-700">
            <div className="mb-4">
              <h3 className="text-sm font-semibold text-gray-300">Nodes</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-2 mt-2">
                {agent.nodes.map((node) => (
                  <div
                    key={node.alias}
                    className="flex flex-col bg-neutral-800 p-2 rounded shadow-sm"
                  >
                    <span className="text-sm font-semibold text-gray-200">{node.alias}</span>
                    <span className="text-xs text-gray-400">{node.type}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* JobList with live updates */}
            <JobList agentId={agent.id} userId={userId} refreshJobs={() => setJobs([...jobs])} />
          </td>
        </tr>
      )}
    </>
  );
}
