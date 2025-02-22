"use client";

import { useRouter } from "next/navigation";
import AgentRow from "./AgentRow";

export default function AgentTable({ agents, userId }: { agents: any[], userId: string | null }) {
  const router = useRouter();

  return (
    <div className="bg-neutral-900 text-white p-6 rounded-lg shadow-md">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold">My Agents</h2>
        <button
          onClick={() => router.push("/agent-builder")}
          className="flex items-center space-x-2 bg-neutral-800 hover:bg-neutral-900 text-white font-semibold px-4 py-2 rounded-md transition"
        >
          <span className="text-lg">+</span>
          <span>Create Agent</span>
        </button>
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
            {agents.length > 0 ? (
              agents.map((agent) => (
                <AgentRow key={agent.id} agent={agent} userId={userId} />
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
