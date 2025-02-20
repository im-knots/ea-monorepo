"use client";

import { useState, useEffect } from "react";
import { ChevronLeft, ChevronRight, AlertTriangle, CheckCircle } from "lucide-react";

// API URL for the EA Agent Manager
const EA_AGENT_MANAGER_URL = "http://agent-manager.ea.erulabs.local/api/v1/agents";

interface JsonEditorProps {
  isOpen: boolean;
  toggle: () => void;
  jsonText: string;
  onJsonChange: (json: string) => void;
}

export default function JsonEditor({ isOpen, toggle, jsonText, onJsonChange }: JsonEditorProps) {
  const [width] = useState(600);
  const [localJson, setLocalJson] = useState(jsonText);
  const [isValidJson, setIsValidJson] = useState(true);
  const [isLoading, setIsLoading] = useState(false);
  const [saveStatus, setSaveStatus] = useState<string | null>(null); // for notification

  // Set localJson whenever jsonText changes
  useEffect(() => {
    setLocalJson(jsonText);  // Sync the local JSON with the prop passed from parent
  }, [jsonText]);

  const validateJson = (text: string) => {
    try {
      JSON.parse(text);
      setIsValidJson(true);
    } catch {
      setIsValidJson(false);
    }
  };

  const handleTextChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newText = e.target.value;
    setLocalJson(newText);
    validateJson(newText);
    onJsonChange(newText);
  };

  // Function to handle saving or updating the agent
  const handleSave = async () => {
    if (!isValidJson || isLoading) return;

    try {
      setIsLoading(true);
      setSaveStatus(null); // Reset status before starting

      const jsonPayload = JSON.parse(localJson);
      const apiUrl = jsonPayload.id
        ? `${EA_AGENT_MANAGER_URL}/${jsonPayload.id}` // Update agent if agent_id exists
        : EA_AGENT_MANAGER_URL; // Otherwise, create a new agent

      const method = jsonPayload.id ? "PUT" : "POST"; // Determine if we need to create or update

      // Make the API call
      const response = await fetch(apiUrl, {
        method,
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(jsonPayload),
      });

      const data = await response.json();

      if (response.ok) {
        if (method === "POST") {
          // If it's a POST request (new agent), update the JSON payload with the new agent_id
          jsonPayload.id = data.agent_id;
        }

        // Update the JSON editor UI with the response data
        setLocalJson(JSON.stringify(jsonPayload, null, 2));
        onJsonChange(JSON.stringify(jsonPayload, null, 2));

        // Show success notification
        setSaveStatus("success");
      } else {
        console.error("Error creating/updating agent:", data);
        setSaveStatus("error");
      }
    } catch (error) {
      console.error("Error:", error);
      setSaveStatus("error");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div
      className="fixed right-0 top-0 h-screen bg-neutral-800 text-white shadow-2xl transition-all flex flex-col"
      style={{ width: isOpen ? `${width}px` : "4rem" }}
    >
      <button
        onClick={toggle}
        className="p-2 bg-neutral-800 hover:bg-neutral-700 transition flex items-center justify-center"
      >
        {isOpen ? <ChevronRight size={20} /> : <ChevronLeft size={20} />}
      </button>

      {isOpen && (
        <div className="flex-1 p-4 overflow-auto flex flex-col">
          <h3 className="text-lg font-bold mb-2">JSON Editor</h3>
          <textarea
            className={`w-full flex-1 bg-neutral-900 text-white p-2 rounded-lg resize-none border ${
              isValidJson ? "border-neutral-700" : "border-red-500"
            }`}
            placeholder="Edit JSON here..."
            value={localJson}
            onChange={handleTextChange}
          />

          {!isValidJson && (
            <div className="flex items-center text-red-400 mt-2">
              <AlertTriangle size={16} className="mr-2" />
              Invalid JSON syntax!
            </div>
          )}

          {/* Notification Indicator */}
          {saveStatus === "success" && (
            <div className="flex items-center text-green-400 mt-2">
              <CheckCircle size={16} className="mr-2" />
              Agent saved/updated successfully!
            </div>
          )}
          {saveStatus === "error" && (
            <div className="flex items-center text-red-400 mt-2">
              <AlertTriangle size={16} className="mr-2" />
              Error saving/updating the agent.
            </div>
          )}

          <div className="flex gap-2 mt-4">
            <button
              onClick={handleSave}
              disabled={!isValidJson || isLoading}
              className="flex-1 py-2 bg-blue-600 hover:bg-blue-500 transition rounded-lg text-white font-medium"
            >
              {isLoading ? "Saving..." : "Save"}
            </button>
            <button className="flex-1 py-2 bg-green-600 hover:bg-green-500 transition rounded-lg text-white font-medium">
              Run
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
