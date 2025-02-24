"use client";

import React, { useEffect, useState } from "react";

const AINU_MANAGER_URL = "http://api.ea.erulabs.local/ainu-manager/api/v1";

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

export default function JobList({
  agentId,
  userId,
  refreshJobs, // 🔥 Refreshes job count in AgentRow
}: {
  agentId: string;
  userId: string | null;
  refreshJobs: () => void;
}) {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [expandedJobId, setExpandedJobId] = useState<string | null>(null);
  const [autoRefresh, setAutoRefresh] = useState<string | null>(null);

  // Fetch jobs for the agent
  const fetchJobs = async () => {
    try {
      const res = await fetch(`${AINU_MANAGER_URL}/users/${userId}`);
      const data = await res.json();
      const filteredJobs = (data.jobs || []).filter((job: Job) => job.agent_id === agentId);

      setJobs(filteredJobs);
      refreshJobs(); // 🔥 Notify AgentRow to update job count
    } catch (error) {
      console.error("Error fetching jobs:", error);
    }
  };

  useEffect(() => {
    fetchJobs();
  }, [agentId, userId]);

  // Auto-refresh job details when expanded
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(async () => {
      await fetchJobs();
      const job = jobs.find((job) => job.id === autoRefresh);
      if (job?.status.toLowerCase() === "completed" || job?.status.toLowerCase() === "complete") {
        setAutoRefresh(null);
      }
    }, 3000);

    return () => clearInterval(interval);
  }, [autoRefresh, jobs]);

  const getStatusColor = (status: string | undefined) => {
    if (!status) return "bg-gray-500 text-white";
    const lowerStatus = status.toLowerCase();

    return lowerStatus === "completed" || lowerStatus === "complete"
      ? "bg-green-600 text-white"
      : lowerStatus === "pending"
      ? "bg-yellow-500 text-black"
      : lowerStatus === "error"
      ? "bg-red-600 text-white"
      : lowerStatus === "executing"
      ? "bg-blue-500 text-white"
      : "bg-gray-500 text-white";
  };

  const handleExpand = (jobId: string) => {
    setExpandedJobId(expandedJobId === jobId ? null : jobId);
    setAutoRefresh(jobId);
  };

  return (
    <div className="bg-neutral-900 text-white p-4 rounded-lg shadow-md">
      <h3 className="text-sm font-semibold mb-2">Jobs</h3>
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
              <React.Fragment key={job.id}>
                <tr
                  className="border-b border-gray-700 hover:bg-neutral-800 transition cursor-pointer"
                  onClick={() => handleExpand(job.id)}
                >
                  <td className="px-4 py-2">{job.job_name}</td>
                  <td className="px-4 py-2">{new Date(job.created_time).toLocaleString()}</td>
                  <td className="px-4 py-2">{job.last_active ? new Date(job.last_active).toLocaleString() : "N/A"}</td>
                  <td className="px-4 py-2">
                    <span className={`px-2 py-1 rounded-md text-xs font-semibold ${getStatusColor(job.status)}`}>
                      {job.status}
                    </span>
                  </td>
                </tr>

                {/* Expanded Job Row: Shows Individual Node Statuses */}
                {expandedJobId === job.id && (
                  <tr>
                    <td colSpan={4} className="p-4 bg-neutral-800 rounded-lg">
                      <h4 className="text-sm font-semibold text-gray-300 mb-2">Node Outputs</h4>
                      {job.nodes?.length ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2">
                          {job.nodes.map((node) => (
                            <div
                              key={node.alias}
                              className="bg-neutral-900 p-3 rounded-md shadow-sm relative"
                            >
                              {/* Status in the upper right corner */}
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
              </React.Fragment>
            ))}
          </tbody>
        </table>
      ) : (
        <p className="text-gray-400 text-sm">No jobs found for this agent.</p>
      )}
    </div>
  );
}
