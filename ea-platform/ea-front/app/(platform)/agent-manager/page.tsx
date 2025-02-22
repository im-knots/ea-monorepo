"use client";

import { useEffect, useState } from "react";
import AgentTable from "../components/AgentTable";

const AINU_MANAGER_URL = "http://ainu-manager.ea.erulabs.local/api/v1";
const AGENT_MANAGER_URL = "http://agent-manager.ea.erulabs.local/api/v1";

// Define TypeScript interface for an Agent
interface Agent {
  id: string;
  name: string;
  creator: string;
  nodes: any[]; // Define proper types later if needed
  edges: any[];
  jobs?: any[]; // Placeholder for jobs, will be populated later
}

export default function AgentManagerPage() {
  const [userId, setUserId] = useState<string | null>(null);
  const [agents, setAgents] = useState<Agent[]>([]); // ✅ Explicitly type the state

  useEffect(() => {
    const fetchUserId = async () => {
      try {
        const res = await fetch(`${AINU_MANAGER_URL}/users`);
        const users = await res.json();
        const user = users.find((u: any) => u.name === "marco@erulabs.ai"); // Match by name or email
        if (user) {
          setUserId(user.id);
        }
      } catch (error) {
        console.error("Error fetching user ID:", error);
      }
    };

    fetchUserId();
  }, []);

  useEffect(() => {
    if (!userId) return;

    const fetchAgents = async () => {
      try {
        const res = await fetch(`${AGENT_MANAGER_URL}/agents?creator_id=${userId}`);
        const agentsData: Agent[] = await res.json(); // ✅ Explicitly cast response to `Agent[]`

        // Fetch detailed agent data including nodes and edges
        const detailedAgents = await Promise.all(
          agentsData.map(async (agent) => {
            const agentRes = await fetch(`${AGENT_MANAGER_URL}/agents/${agent.id}`);
            const agentDetails = await agentRes.json();

            return {
              ...agent,
              nodes: agentDetails.nodes || [],
              edges: agentDetails.edges || [],
              jobs: [], // Placeholder for jobs, will populate later
            };
          })
        );

        setAgents(detailedAgents); // ✅ Now TypeScript knows the structure
      } catch (error) {
        console.error("Error fetching agents:", error);
      }
    };

    fetchAgents();
  }, [userId]);

  return (
    <div className="flex">
      <main className="flex-grow p-4">
        <AgentTable agents={agents} userId={userId} />
      </main>
    </div>
  );
}
