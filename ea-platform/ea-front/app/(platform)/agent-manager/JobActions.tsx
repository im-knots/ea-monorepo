// app/(platform)/agent-manager/server/JobActions.ts
"use server";

import { cookies } from "next/headers";

const AINU_MANAGER_URL = "http://api.erulabs.local/ainu-manager/api/v1";

export async function fetchJobsForAgent(agentId: string, userId: string) {
  const cookieStore = await cookies(); // ✅ Must await cookies() in a server function
  const token = cookieStore.get("token")?.value;
  if (!token) throw new Error("Missing auth token");

  const res = await fetch(`${AINU_MANAGER_URL}/users/${userId}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    cache: "no-store",
  });

  if (!res.ok) throw new Error("Failed to fetch jobs");

  const data = await res.json();
  return (data.jobs || []).filter((job: any) => job.agent_id === agentId);
}
