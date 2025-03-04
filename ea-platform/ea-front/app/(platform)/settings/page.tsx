"use client";

import { useState, useEffect } from "react";
import Image from "next/image";

const API_URL = "http://api.ea.erulabs.local/credentials-manager/api/v1/credentials";

const SUPPORTED_APIS = [
  { name: "OpenAI", key: "openai_api_key", icon: "/openai-icon.png" },
  { name: "Anthropic", key: "anthropic_api_key", icon: "/anthropic-icon.png" },
  { name: "xAI", key: "xai_api_key", icon: "/xai-icon.png" },
  { name: "Google Gemini", key: "google_api_key", icon: "/google-gemini-icon.png" },
  { name: "GitHub", key: "github_api_key", icon: "/github-icon.png" },
  { name: "Jira", key: "jira_api_key", icon: "/jira-icon.png" },
  { name: "Custom", key: "custom", icon: "/logo.png" },
];

export default function SettingsPage() {
  const [keys, setKeys] = useState<Record<string, string>>({});
  const [keyInputs, setKeyInputs] = useState<Record<string, string>>({});
  const [customKeys, setCustomKeys] = useState<Record<string, string>>({});
  const [jwtToken, setJwtToken] = useState<string | null>(null);
  const [messages, setMessages] = useState<Record<string, string>>({});

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
      }
    };

    fetchToken();
  }, []);

  const handleInputChange = (key: string, value: string) => {
    setKeyInputs((prev) => ({ ...prev, [key]: value }));
  };

  const handleCustomKeyChange = (key: string, value: string) => {
    setCustomKeys((prev) => ({ ...prev, [key]: value }));
  };

  const handleSubmit = async (apiKey: string, isCustom: boolean = false) => {
    if (!jwtToken) {
      setMessages((prev) => ({ ...prev, [apiKey]: "❌ Authentication error" }));
      return;
    }

    const keyName = isCustom ? customKeys[apiKey] || "" : apiKey;
    if (isCustom && !keyName.trim()) {
      setMessages((prev) => ({ ...prev, [apiKey]: "❌ Custom key name required" }));
      return;
    }

    const payload = { [keyName]: keyInputs[apiKey] || "" };

    try {
      const response = await fetch(API_URL, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json-patch+json",
          Authorization: `Bearer ${jwtToken}`,
        },
        body: JSON.stringify(payload),
      });

      if (response.ok) {
        setKeys((prev) => ({ ...prev, [apiKey]: "********" })); // Mask key
        setMessages((prev) => ({ ...prev, [apiKey]: "✅ Saved successfully!" }));
      } else {
        const errorData = await response.json();
        setMessages((prev) => ({ ...prev, [apiKey]: `❌ Error: ${errorData.error}` }));
      }
    } catch (error) {
      console.error("Error saving credentials:", error);
      setMessages((prev) => ({ ...prev, [apiKey]: "❌ An error occurred. Try again." }));
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-neutral-900 text-white p-6">
      <div className="w-full max-w-3xl bg-neutral-800 p-6 rounded-lg shadow-md">
        <h2 className="text-2xl font-semibold text-center mb-6">Third-Party API Keys</h2>

        <div className="space-y-4">
          {SUPPORTED_APIS.map((api) => (
            <div key={api.key} className="bg-neutral-700 p-4 rounded-lg shadow-md flex items-center space-x-4">
              {/* Provider Icon */}
              <Image
                src={api.icon}
                alt={`${api.name} Logo`}
                width={32}
                height={32}
                className="rounded-full"
              />

              <div className="flex-1">
                <h3 className="text-lg font-medium">{api.name}</h3>

                {/* Custom Key Override Field */}
                {api.key === "custom" && (
                  <input
                    type="text"
                    value={customKeys[api.key] || ""}
                    onChange={(e) => handleCustomKeyChange(api.key, e.target.value)}
                    className="w-full mt-1 px-3 py-2 bg-neutral-600 border border-neutral-500 rounded-md text-white text-sm"
                    placeholder="Custom key name"
                    required
                  />
                )}

                {/* API Key Input */}
                <input
                  type="text"
                  value={keys[api.key] ? "********" : keyInputs[api.key] || ""}
                  onChange={(e) => handleInputChange(api.key, e.target.value)}
                  className="w-full mt-1 px-3 py-2 bg-neutral-600 border border-neutral-500 rounded-md text-white text-sm"
                  placeholder={`Enter ${api.name} API key`}
                />
              </div>

              {/* Save Button (Positioned on Right) */}
              <button
                onClick={() => handleSubmit(api.key, api.key === "custom")}
                className="ml-4 bg-blue-500 hover:bg-blue-600 text-white font-semibold py-2 px-4 rounded-md transition"
              >
                {keys[api.key] ? "Update" : "Save"}
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
