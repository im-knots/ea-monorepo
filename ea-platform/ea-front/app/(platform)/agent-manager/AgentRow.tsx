"use client";

import { useState, useTransition } from "react";
import JobListClient from "./JobListClient";
import { startAgentJob, deleteAgent } from "./AgentActions";

export default function AgentRow({ agent, userId, initialJobs }: { agent: any; userId: string; initialJobs: any[] }) {
  const [expanded, setExpanded] = useState(false);
  const [isStarting, startTransition] = useTransition();
  const [isDeleting, deleteTransition] = useTransition();

  const handleStart = () => {
    startTransition(async () => {
      await startAgentJob(agent.id, userId);
      setExpanded(true);
    });
  };

  const handleDelete = () => {
    deleteTransition(async () => {
      await deleteAgent(agent.id);
      location.reload();
    });
  };

  return (
    <>
      <tr className="hover:bg-neutral-800 transition cursor-pointer" onClick={() => setExpanded(!expanded)}>
        <td className="px-4 py-3">{agent.name}</td>
        <td className="px-4 py-3 text-center">{agent.nodes.length}</td>
        <td className="px-4 py-3 text-center">{agent.jobCount}</td>
        <td className="px-4 py-3 text-center flex justify-center space-x-2">
          <button onClick={(e) => { e.stopPropagation(); handleStart(); }} disabled={isStarting} className="btn bg-green-600">
            {isStarting ? "Starting..." : "Start"}
          </button>
          <button className="btn bg-blue-600">Modify</button>
          <button onClick={(e) => { e.stopPropagation(); handleDelete(); }} disabled={isDeleting} className="btn bg-red-600">
            {isDeleting ? "Deleting..." : "Delete"}
          </button>
        </td>
      </tr>

      {expanded && (
        <tr>
          <td colSpan={4} className="bg-neutral-900 p-4 border-t border-gray-700">
            <JobListClient initialJobs={initialJobs} agentId={agent.id} userId={userId} />
          </td>
        </tr>
      )}
    </>
  );
}
