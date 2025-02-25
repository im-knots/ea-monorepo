"use client";

import { useEffect, useState } from "react";
import AgentTable from "../components/AgentTable";

export default function AgentManagerPage() {
  const [userId, setUserId] = useState<string | null>(null);
  const [token, setToken] = useState<string | null>(null);

  // Fetch JWT token from the server
  const fetchToken = async () => {
    try {
      const res = await fetch("/api/auth/token", { credentials: "include" });
      if (!res.ok) throw new Error("Failed to fetch token");
      const data = await res.json();
      setToken(data.token);

      // Decode the token and extract the user ID
      const payload = JSON.parse(atob(data.token.split(".")[1])); // Decode JWT payload
      if (payload.iss) {
        setUserId(payload.iss); // ✅ Extract UUID from `iss` field
      }
    } catch (error) {
      console.error("Error fetching or decoding token:", error);
      setToken(null);
      setUserId(null);
    }
  };

  // Fetch token on component mount
  useEffect(() => {
    fetchToken();
  }, []);

  return (
    <div className="flex">
      <main className="flex-grow p-4">
        <AgentTable userId={userId} /> {/* ✅ Pass extracted userId directly */}
      </main>
    </div>
  );
}
