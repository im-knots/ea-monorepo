"use client";

import { useState, useEffect } from "react";
import { ChevronLeft, ChevronRight, AlertTriangle, CheckCircle } from "lucide-react";

// API URL for the EA Agent Manager and Job Engine
const EA_AGENT_MANAGER_URL = "http://api.ea.erulabs.local/agent-manager/api/v1/agents";
const EA_JOB_API_URL = "http://api.ea.erulabs.local/job-api/api/v1/jobs";

interface JsonEditorProps {
  isOpen: boolean;
  toggle: () => void;
  jsonText: string;
  onJsonChange: (json: string) => void;
  agentId: string | null;
  updateAgentId: (id: string) => void;
  creatorId: string;
  onJobStarted?: (jobId: string) => void; // New callback for job started events
}

export default function JsonEditor({
  isOpen,
  toggle,
  jsonText,
  onJsonChange,
  agentId,
  updateAgentId,
  creatorId,
  onJobStarted,
}: JsonEditorProps) {
  const [width] = useState(600);
  const [localJson, setLocalJson] = useState(jsonText);
  const [isValidJson, setIsValidJson] = useState(true);
  const [isLoading, setIsLoading] = useState(false);
  const [saveStatus, setSaveStatus] = useState<string | null>(null);
  const [showModal, setShowModal] = useState(false);
  const [jobId, setJobId] = useState<string | null>(null);

  // Set localJson whenever jsonText changes
  useEffect(() => {
    setLocalJson(jsonText);
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

  const handleSave = async () => {
    if (!isValidJson || isLoading) return;

    try {
      setIsLoading(true);
      setSaveStatus(null);

      const jsonPayload = JSON.parse(localJson);
      const finalAgentId = agentId || jsonPayload.id;

      const apiUrl = finalAgentId
        ? `${EA_AGENT_MANAGER_URL}/${finalAgentId}`
        : EA_AGENT_MANAGER_URL;

      const method = finalAgentId ? "PUT" : "POST";

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
          jsonPayload.id = data.agent_id;
        }

        setLocalJson(JSON.stringify(jsonPayload, null, 2));
        onJsonChange(JSON.stringify(jsonPayload, null, 2));
        setSaveStatus("success");

        // Update agentId if the response is a successful POST
        if (method === "POST") {
          updateAgentId(data.agent_id);
        }
      } else {
        setSaveStatus("error");
      }
    } catch (error) {
      setSaveStatus("error");
    } finally {
      setIsLoading(false);
    }
  };

  const handleStartJob = async () => {
    if (!agentId) return;
  
    try {
      setIsLoading(true);
      setSaveStatus(null);
  
      const response = await fetch(EA_JOB_API_URL, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          agent_id: agentId,
          user_id: creatorId,
        }),
      });
  
      const data = await response.json();
  
      if (response.ok) {
        console.log("Job started successfully:", data.job_name);
        setJobId(data.job_name);
        setSaveStatus(`Job submitted successfully`);
  
        // Notify parent component
        if (onJobStarted) {
          onJobStarted(data.job_name);
        }
      } else {
        setSaveStatus("Error starting job");
      }
    } catch (error) {
      setSaveStatus("Error starting job");
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

          {/* Show job status message */}
          {saveStatus && saveStatus !== "success" && saveStatus !== "error" && (
            <div className="flex items-center mt-2">
              {saveStatus === "Job submitted successfully" ? (
                <CheckCircle size={16} className="mr-2 text-green-400" />
              ) : (
                <AlertTriangle size={16} className="mr-2 text-red-400" />
              )}
              {saveStatus}
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
            <button
              onClick={handleStartJob}
              disabled={!agentId || isLoading}
              className={`flex-1 py-2 ${agentId ? "bg-green-600 hover:bg-green-500" : "bg-gray-600 cursor-not-allowed"} transition rounded-lg text-white font-medium`}
            >
              Start
            </button>
          </div>
        </div>
      )}
    </div>
  );
}