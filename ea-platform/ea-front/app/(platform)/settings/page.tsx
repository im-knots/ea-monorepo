"use client";

import { useState, useEffect } from "react";

const API_URL = "http://api.ea.erulabs.local/credentials-manager/api/v1/credentials";

const SUPPORTED_APIS = [
  { name: "OpenAI", key: "openai_api_key" },
  { name: "Anthropic", key: "anthropic_api_key" },
  { name: "GitHub", key: "github_api_key" },
  { name: "Jira", key: "jira_api_key" },
  { name: "Custom", key: "custom" },
];

export default function SettingsPage() {
  const [selectedApi, setSelectedApi] = useState(SUPPORTED_APIS[0]); // Default: OpenAI
  const [apiKey, setApiKey] = useState("");
  const [customKeyName, setCustomKeyName] = useState(""); // Custom API key name
  const [message, setMessage] = useState("");
  const [jwtToken, setJwtToken] = useState<string | null>(null);

  // 🔹 Extract JWT token from cookies
  useEffect(() => {
    const fetchToken = async () => {
      try {
        const res = await fetch("/api/auth/token", { credentials: "include" });
        if (!res.ok) throw new Error("Failed to fetch token");

        const data = await res.json();
        if (data.token) {
          setJwtToken(data.token);
        } else {
          throw new Error("Token not found");
        }
      } catch (error) {
        console.error("Error fetching JWT token:", error);
        setMessage("❌ Unable to authenticate. Please log in again.");
      }
    };

    fetchToken();
  }, []);

  const handleApiChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const selected = SUPPORTED_APIS.find((api) => api.key === e.target.value);
    if (selected) {
      setSelectedApi(selected);
      setApiKey(""); // Reset API key field
      setCustomKeyName(""); // Reset custom key field
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setMessage("");

    if (!jwtToken) {
      setMessage("❌ Unable to authenticate. Please log in again.");
      return;
    }

    if (selectedApi.key === "custom" && !customKeyName.trim()) {
      setMessage("❌ Please provide a key name for your custom API.");
      return;
    }

    const payload =
      selectedApi.key === "custom"
        ? { [customKeyName]: apiKey }
        : { [selectedApi.key]: apiKey };

    try {
      const response = await fetch(API_URL, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json-patch+json",
          Authorization: `Bearer ${jwtToken}`, // ✅ Inject JWT token
        },
        body: JSON.stringify(payload),
      });

      if (response.ok) {
        setMessage("✅ Credentials updated successfully!");
      } else {
        const errorData = await response.json();
        setMessage(`❌ Failed to update credentials: ${errorData.error}`);
      }
    } catch (error) {
      console.error("Error saving credentials:", error);
      setMessage("❌ An error occurred. Please try again.");
    }
  };

  return (
    <div className="flex flex-col items-center justify-center h-full bg-neutral-900 text-white">
      <div className="w-full max-w-lg bg-neutral-800 p-6 rounded-lg shadow-md">
        <h2 className="text-2xl font-semibold text-center mb-6">User Settings</h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* API Dropdown */}
          <div>
            <label className="block text-sm font-medium text-neutral-400">Select API</label>
            <select
              value={selectedApi.key}
              onChange={handleApiChange}
              className="w-full mt-1 px-4 py-2 bg-neutral-700 border border-neutral-600 rounded-md text-white focus:ring focus:ring-blue-500"
            >
              {SUPPORTED_APIS.map((api) => (
                <option key={api.key} value={api.key}>
                  {api.name}
                </option>
              ))}
            </select>
          </div>

          {/* Custom Key Name Input (Only Shows If "Custom" is Selected) */}
          {selectedApi.key === "custom" && (
            <div>
              <label className="block text-sm font-medium text-neutral-400">Custom Key Name</label>
              <input
                type="text"
                value={customKeyName}
                onChange={(e) => setCustomKeyName(e.target.value)}
                className="w-full mt-1 px-4 py-2 bg-neutral-700 border border-neutral-600 rounded-md text-white focus:ring focus:ring-blue-500"
                placeholder="Enter your custom key name"
                required
              />
            </div>
          )}

          {/* API Key Input */}
          <div>
            <label className="block text-sm font-medium text-neutral-400">
              {selectedApi.key === "custom" ? "Custom API Key" : `${selectedApi.name} API Key`}
            </label>
            <input
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              className="w-full mt-1 px-4 py-2 bg-neutral-700 border border-neutral-600 rounded-md text-white focus:ring focus:ring-blue-500"
              placeholder={`Enter your ${selectedApi.key === "custom" ? "custom" : selectedApi.name} API key`}
              required
            />
          </div>

          {/* Submit Button */}
          <button
            type="submit"
            className="w-full bg-blue-500 hover:bg-blue-600 text-white font-semibold py-2 rounded-md shadow-md transition"
          >
            Save Credentials
          </button>

          {/* Status Message */}
          {message && (
            <p className="text-sm text-center mt-2 font-medium text-neutral-300">{message}</p>
          )}
        </form>
      </div>
    </div>
  );
}
