"use client";

import { useEffect, useState } from "react";
import AgentTable from "../components/AgentTable";

const AINU_MANAGER_URL = "http://ainu-manager.ea.erulabs.local/api/v1";

export default function AgentManagerPage() {
  const [userId, setUserId] = useState<string | null>(null);

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

  return (
    <div className="flex">
      <main className="flex-grow p-4">
        <AgentTable userId={userId} /> {/* ✅ Remove `agents` prop */}
      </main>
    </div>
  );
}
