import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import Link from "next/link";
import AgentRow from "./AgentRow";
import { fetchJobsForAgent } from "./JobActions";

const AGENT_MANAGER_URL = "http://api.erulabs.local/agent-manager/api/v1/agents";

export default async function AgentTable() {
  const cookieStore = await cookies();  // ✅ Add await here
  const token = cookieStore.get("token")?.value;

  if (!token) {
    return redirect("/login");
  }

  const payload = JSON.parse(atob(token.split(".")[1]));
  const userId = payload?.sub;

  const agentsRes = await fetch(AGENT_MANAGER_URL, {
    headers: { Authorization: `Bearer ${token}` },
    cache: "no-store",
  });

  if (!agentsRes.ok) {
    throw new Error("Failed to fetch agents");
  }

  const agents = await agentsRes.json();

  const detailedAgents = await Promise.all(
    agents.map(async (agent: any) => {
      const [detailRes, jobs] = await Promise.all([
        fetch(`${AGENT_MANAGER_URL}/${agent.id}`, {
          headers: { Authorization: `Bearer ${token}` },
          cache: "no-store",
        }),
        fetchJobsForAgent(agent.id, userId),
      ]);

      if (!detailRes.ok) {
        throw new Error("Failed to fetch agent details");
      }

      const details = await detailRes.json();

      return {
        ...agent,
        ...details,
        jobs, 
        jobCount: jobs.length,
      };
    })
  );

  return (
    <div className="bg-neutral-900 text-white p-6 rounded-lg shadow-md">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold">My Agents</h2>
        <div className="flex space-x-2 ml-auto">
          <Link href="/agent-builder" className="btn">+ Create Agent</Link>
          <Link href="/node-builder" className="btn">+ Create Node</Link>
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
            {detailedAgents.map(agent => (
              <AgentRow key={agent.id} agent={agent} userId={userId} initialJobs={agent.jobs} />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
