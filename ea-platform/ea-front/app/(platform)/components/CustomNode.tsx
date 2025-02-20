"use client";

import React from "react";
import { Handle, Position, NodeProps } from "reactflow";

interface Parameter {
  key: string;
  type: string;
  default: any;
  enum?: string[];
}

interface CustomNodeData {
  alias: string;
  type: string; // Preserve original node definition type
  label: string;
  parameters: Parameter[];
  parametersState: Record<string, any>;
  updateNodeData: (id: string, key: string, value: any) => void;
}

export default function CustomNode({ id, data }: NodeProps<CustomNodeData>) {
  // Determine header background color based on node type
  let typeBgColor = "bg-gray-700"; // Default
  if (data.type.startsWith("worker")) typeBgColor = "bg-purple-500";
  else if (data.type.startsWith("input")) typeBgColor = "bg-green-500";
  else if (data.type.startsWith("destination")) typeBgColor = "bg-blue-500";

  return (
    <div className="bg-neutral-800 border border-neutral-600 text-white p-4 rounded-lg shadow-md min-w-[220px]">
      {/* Node Type Header with Dynamic Background Color */}
      <div className={`${typeBgColor} text-white text-xs font-bold px-2 py-1 rounded-md mb-2 text-center`}>
        {data.type}
      </div>

      {/* Alias Field */}
      <label className="text-xs text-gray-400 block mb-1">alias</label>
      <input
        type="text"
        value={data.alias || ""} // Ensure alias starts empty
        onChange={(e) => data.updateNodeData(id, "alias", e.target.value)}  // Update alias
        className="bg-neutral-700 text-white text-xs p-1 rounded w-full mb-3 border border-gray-600"
        placeholder="Set alias..."
      />

      {/* Render Parameters Dynamically */}
      <div className="space-y-2">
        {data.parameters.map((param) => (
          <div key={param.key} className="text-xs">
            <label className="block text-gray-400">{param.key}</label>

            {param.type === "bool" ? (
              // Boolean toggle switch with green active state
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={data.parametersState[param.key] ?? param.default}
                  onChange={(e) => data.updateNodeData(id, param.key, e.target.checked)}  // Update parameter state
                  className="sr-only peer"
                />
                <div className="w-9 h-5 bg-gray-600 peer-focus:ring-2 peer-focus:ring-blue-300 rounded-full peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-1 after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-green-500"></div>
              </label>
            ) : param.enum ? (
              // Enum dropdown
              <select
                value={data.parametersState[param.key] ?? param.default}
                onChange={(e) => data.updateNodeData(id, param.key, e.target.value)}  // Update parameter state
                className="bg-neutral-700 text-white text-xs p-1 rounded w-full"
              >
                {param.enum.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
            ) : (
              // Regular input field for other types
              <input
                type="text"
                value={data.parametersState[param.key] ?? param.default}
                onChange={(e) => data.updateNodeData(id, param.key, e.target.value)}  // Update parameter state
                className="bg-neutral-700 text-white text-xs p-1 rounded w-full"
              />
            )}
          </div>
        ))}
      </div>

      {/* Handles for connections */}
      <Handle type="target" position={Position.Left} className="w-2 h-2 bg-blue-500" />
      <Handle type="source" position={Position.Right} className="w-2 h-2 bg-blue-500" />
    </div>
  );
}
