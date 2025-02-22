"use client";

import { useState } from "react";
import JobList from "./JobList";

const EA_JOB_API_URL = "http://job-api.ea.erulabs.local/api/v1";

interface Node {
  alias: string;
  type: string;
}

interface Agent {
  id: string;
  name: string;
  nodes: Node[];
}

export default function AgentRow({ agent, userId }: { agent: Agent; userId: string | null }) {
  const [expanded, setExpanded] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  const [jobs, setJobs] = useState<any[]>([]); // Track jobs in AgentRow

  // Start Agent Job & Trigger Refresh
  const startAgentJob = async () => {
    if (!userId) return;

    setIsStarting(true);
    try {
      const response = await fetch(`${EA_JOB_API_URL}/jobs`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          agent_id: agent.id,
          user_id: userId,
        }),
      });

      if (!response.ok) throw new Error("Failed to start agent job");

      setIsStarting(false);
      setExpanded(true); // Ensure row expands after starting
      refreshJobs(); // 🔥 Trigger a refresh for the JobList
    } catch (error) {
      console.error("Error starting agent job:", error);
      setIsStarting(false);
    }
  };

  // Function to refresh jobs in JobList
  const refreshJobs = () => {
    setJobs([...jobs]); // Trigger re-render by modifying state
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
          <button
            onClick={(e) => {
              e.stopPropagation(); // Prevent row collapse
              startAgentJob();
            }}
            className={`px-3 py-1 rounded-md text-sm transition ${
              isStarting ? "bg-gray-600 cursor-not-allowed" : "bg-green-600 hover:bg-green-700"
            } text-white`}
            disabled={isStarting}
          >
            {isStarting ? "Starting..." : "Start"}
          </button>
          <button className="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded-md text-sm transition">
            Modify
          </button>
          <button className="bg-red-600 hover:bg-red-700 text-white px-3 py-1 rounded-md text-sm transition">
            Delete
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
            <JobList agentId={agent.id} userId={userId} refreshJobs={refreshJobs} />
          </td>
        </tr>
      )}
    </>
  );
}
