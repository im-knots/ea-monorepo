// app/(platform)/agent-manager/server/agentActions.ts
"use server";

import { cookies } from "next/headers";

export async function startAgentJob(agentId: string, userId: string) {
  const cookieStore = await cookies(); // ✅ await cookies()
  const token = cookieStore.get("token")?.value;
  if (!token) throw new Error("No token");

  const res = await fetch("http://api.erulabs.local/job-api/api/v1/jobs", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ agent_id: agentId, user_id: userId }),
  });

  if (!res.ok) throw new Error("Failed to start agent job");
}

export async function deleteAgent(agentId: string) {
  const cookieStore = await cookies(); // ✅ await cookies()
  const token = cookieStore.get("token")?.value;
  if (!token) throw new Error("No token");

  const res = await fetch(`http://api.erulabs.local/agent-manager/api/v1/agents/${agentId}`, {
    method: "DELETE",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!res.ok) throw new Error("Failed to delete agent");
}
