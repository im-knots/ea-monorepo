"use client";

import { useState } from "react";
import JobList from "./JobList";

interface Node {
  alias: string;
  type: string;
}

interface Agent {
  id: string;
  name: string;
  nodes: Node[];
  jobs: any[];
}

export default function AgentRow({ agent, userId }: { agent: Agent; userId: string | null }) {
  const [expanded, setExpanded] = useState(false);

  return (
    <>
      {/* Main Agent Row */}
      <tr
        className="border-b border-gray-700 hover:bg-neutral-800 transition duration-200 cursor-pointer"
        onClick={() => setExpanded(!expanded)}
      >
        <td className="px-4 py-3">{agent.name}</td>
        <td className="px-4 py-3 text-center">{agent.nodes.length}</td>
        <td className="px-4 py-3 text-center">{agent.jobs.length}</td>
        <td className="px-4 py-3 text-center flex justify-center space-x-2">
          <button className="bg-green-600 hover:bg-green-700 text-white px-3 py-1 rounded-md text-sm transition">
            Start
          </button>
          <button className="bg-blue-600 hover:bg-blue-700 text-white px-3 py-1 rounded-md text-sm transition">
            Modify
          </button>
          <button className="bg-red-600 hover:bg-red-700 text-white px-3 py-1 rounded-md text-sm transition">
            Delete
          </button>
        </td>
      </tr>

      {/* Expanded View: Nodes & Jobs */}
      {expanded && (
        <tr>
          <td colSpan={4} className="bg-neutral-900 p-4 border-t border-gray-700">
            {/* Nodes Section */}
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

            {/* Jobs Section */}
            <JobList agentId={agent.id} userId={userId} />
          </td>
        </tr>
      )}
    </>
  );
}
