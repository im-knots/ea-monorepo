// app/(platform)/agent-manager/JobListClient.tsx
"use client";

import { useState } from "react";

interface Job {
  id: string;
  job_name: string;
  created_time: string;
  last_active?: string;
  status: string;
  agent_id: string;
  nodes?: Node[];
}

interface Node {
  alias: string;
  type?: string;
  output?: string;
  status?: string;
  lastUpdated?: string;
}

export default function JobListClient({
  initialJobs,
  agentId,
  userId,
}: {
  initialJobs: Job[];
  agentId: string;
  userId: string;
}) {
  const [jobs] = useState<Job[]>(initialJobs);
  const [expandedJobId, setExpandedJobId] = useState<string | null>(null);

  const getStatusColor = (status: string | undefined) => {
    const s = status?.toLowerCase();
    return s === "completed" || s === "complete"
      ? "bg-green-600 text-white"
      : s === "pending"
      ? "bg-yellow-500 text-black"
      : s === "executing"
      ? "bg-blue-500 text-white"
      : s === "error"
      ? "bg-red-600 text-white"
      : "bg-gray-500 text-white";
  };

  const handleExpand = (jobId: string) => {
    setExpandedJobId(expandedJobId === jobId ? null : jobId);
  };

  return (
    <div className="bg-neutral-900 text-white p-4 rounded-lg shadow-md">
      <div className="flex justify-between items-center mb-2">
        <h3 className="text-sm font-semibold">Jobs</h3>
        <button
          onClick={() => location.reload()}
          className="text-xs bg-neutral-700 hover:bg-neutral-600 px-2 py-1 rounded"
        >
          Refresh
        </button>
      </div>

      {jobs.length > 0 ? (
        <table className="w-full text-sm border border-gray-700 rounded-lg">
          <thead className="bg-neutral-800 text-gray-300 uppercase">
            <tr>
              <th className="px-4 py-2 text-left">Job Name</th>
              <th className="px-4 py-2 text-left">Created</th>
              <th className="px-4 py-2 text-left">Last Active</th>
              <th className="px-4 py-2 text-left">Status</th>
            </tr>
          </thead>
          <tbody>
            {jobs.map((job) => (
              <>
                <tr
                  key={job.id}
                  className="border-b border-gray-700 hover:bg-neutral-800 cursor-pointer"
                  onClick={() => handleExpand(job.id)}
                >
                  <td className="px-4 py-2">{job.job_name}</td>
                  <td className="px-4 py-2">{new Date(job.created_time).toLocaleString()}</td>
                  <td className="px-4 py-2">
                    {job.last_active ? new Date(job.last_active).toLocaleString() : "N/A"}
                  </td>
                  <td className="px-4 py-2">
                    <span className={`px-2 py-1 rounded-md text-xs font-semibold ${getStatusColor(job.status)}`}>
                      {job.status}
                    </span>
                  </td>
                </tr>

                {expandedJobId === job.id && (
                  <tr key={job.id + "-expanded"}>
                    <td colSpan={4} className="p-4 bg-neutral-800 rounded-lg">
                      <h4 className="text-sm font-semibold text-gray-300 mb-2">Node Outputs</h4>
                      {job.nodes?.length ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
                          {job.nodes.map((node) => (
                            <div
                              key={node.alias}
                              className="bg-neutral-900 p-3 rounded-md shadow-sm relative"
                            >
                              <span
                                className={`absolute top-2 right-2 px-2 py-1 rounded-md text-xs font-semibold ${getStatusColor(node.status)}`}
                              >
                                {node.status}
                              </span>

                              <p className="text-gray-300 text-sm font-semibold">{node.alias}</p>
                              <p className="text-xs text-gray-400">
                                {node.lastUpdated
                                  ? `Last Updated: ${new Date(node.lastUpdated).toLocaleString()}`
                                  : "No timestamp"}
                              </p>

                              <pre className="bg-neutral-950 p-2 text-xs text-gray-300 rounded mt-2 overflow-y-auto max-h-40 whitespace-pre-wrap break-words">
                                {node.output ? JSON.stringify(JSON.parse(node.output), null, 2) : "No output"}
                              </pre>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <p className="text-gray-400 text-sm">No node outputs available.</p>
                      )}
                    </td>
                  </tr>
                )}
              </>
            ))}
          </tbody>
        </table>
      ) : (
        <p className="text-gray-400 text-sm">No jobs found for this agent.</p>
      )}
    </div>
  );
}
