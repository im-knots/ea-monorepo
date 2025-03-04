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
  type: string;
  label: string;
  parameters: Parameter[];
  parametersState: Record<string, any>;
  outputs?: Record<string, any>;
  updateNodeData: (id: string, key: string, value: any) => void;
  status?: string;
}

export default function CustomNode({ id, data }: NodeProps<CustomNodeData>) {
  let typeBgColor = "bg-gray-700";
  if (data.type.startsWith("worker")) typeBgColor = "bg-purple-500";
  else if (data.type.startsWith("input")) typeBgColor = "bg-green-500";
  else if (data.type.startsWith("destination")) typeBgColor = "bg-blue-500";

  let statusColor = "bg-gray-500";
  if (data.status === "executing") statusColor = "bg-blue-500";
  else if (data.status === "Completed") statusColor = "bg-green-500";
  else if (data.status === "failed") statusColor = "bg-red-500";

  // **Update parameters in place (arrays & objects included)**
  const handleParameterUpdate = (key: string, value: any) => {
    data.updateNodeData(id, key, value);
  };

  // **Handle updates inside nested objects in arrays**
  const handleNestedUpdate = (arrayKey: string, index: number, field: string, value: any) => {
    const existingArray = Array.isArray(data.parametersState[arrayKey]) ? [...data.parametersState[arrayKey]] : [];
    existingArray[index] = { ...existingArray[index], [field]: value }; 
    handleParameterUpdate(arrayKey, existingArray);
  };

  // **Render input fields dynamically**
  const renderParameterField = (param: Parameter) => {
    const paramValue = data.parametersState[param.key] ?? param.default;

    if (typeof paramValue === "boolean") {
      return (
        <label key={param.key} className="flex items-center space-x-2 text-xs">
          <span className="text-gray-400">{param.key}</span>
          <button
            onClick={() => handleParameterUpdate(param.key, !paramValue)}
            className={`w-10 h-5 rounded-full flex items-center p-1 transition-colors ${
              paramValue ? "bg-green-500" : "bg-gray-500"
            }`}
          >
            <div className={`w-4 h-4 bg-white rounded-full shadow-md transform transition-transform ${
              paramValue ? "translate-x-5" : "translate-x-0"
            }`} />
          </button>
        </label>
      );
    }

    if (param.enum) {
      return (
        <div key={param.key} className="text-xs">
          <label className="block text-gray-400">{param.key}</label>
          <select
            value={paramValue}
            onChange={(e) => handleParameterUpdate(param.key, e.target.value)}
            className="bg-neutral-700 text-white text-xs p-1 rounded w-full border border-gray-600"
          >
            {param.enum.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </div>
      );
    }

    if (data.type === "input.internal.text" && typeof paramValue === "string") {
      return (
        <div key={param.key} className="text-xs">
          <label className="block text-gray-400">{param.key}</label>
          <textarea
            value={paramValue}
            onChange={(e) => handleParameterUpdate(param.key, e.target.value)}
            className="bg-neutral-700 text-white text-xs p-2 rounded w-full border border-gray-600 h-20 resize-none"
            placeholder="Enter text..."
          />
        </div>
      );
    }

    if (typeof paramValue === "number" || typeof paramValue === "string") {
      return (
        <div key={param.key} className="text-xs">
          <label className="block text-gray-400">{param.key}</label>
          <input
            type="text"
            value={paramValue}
            onChange={(e) => handleParameterUpdate(param.key, e.target.value)}
            className="bg-neutral-700 text-white text-xs p-1 rounded w-full border border-gray-600"
          />
        </div>
      );
    }

    if (Array.isArray(paramValue)) {
      return (
        <div key={param.key} className="text-xs">
          <label className="block text-gray-400">{param.key} (Array)</label>
          {paramValue.map((item, index) => (
            <div key={index} className="pl-4 border-l border-gray-600 ml-2">
              {typeof item === "object" && item !== null ? (
                Object.entries(item).map(([subKey, subValue]) => (
                  <div key={subKey} className="text-xs">
                    <label className="block text-gray-400">{`${param.key}[${index}].${subKey}`}</label>
                    <input
                      type="text"
                      value={subValue as string}
                      onChange={(e) => handleNestedUpdate(param.key, index, subKey, e.target.value)}
                      className="bg-neutral-700 text-white text-xs p-1 rounded w-full border border-gray-600"
                    />
                  </div>
                ))
              ) : (
                <input
                  type="text"
                  value={item}
                  onChange={(e) => {
                    const updatedArray = [...paramValue];
                    updatedArray[index] = e.target.value;
                    handleParameterUpdate(param.key, updatedArray);
                  }}
                  className="bg-neutral-700 text-white text-xs p-1 rounded w-full border border-gray-600"
                />
              )}
            </div>
          ))}
        </div>
      );
    }

    return null;
  };

  return (
    <div className="bg-neutral-800 border border-neutral-600 text-white p-4 rounded-lg shadow-md min-w-[200px]">
      <div className="flex items-center mb-2">
        <div className={`w-2 h-2 rounded-full ${statusColor} mr-2`} />
        <div className={`${typeBgColor} text-white text-xs font-bold px-3 py-1 rounded-md w-full`}>
          {data.type}
        </div>
      </div>

      <label className="text-xs text-gray-400 block mb-1">Alias</label>
      <input
        type="text"
        value={data.alias || ""}
        onChange={(e) => data.updateNodeData(id, "alias", e.target.value)}
        className="bg-neutral-700 text-white text-xs p-1 rounded w-full mb-3 border border-gray-600"
        placeholder="Set alias..."
      />

      {/* Render parameters dynamically */}
      <div className="space-y-2">
        {data.parameters.map((param) => renderParameterField(param))}
      </div>

      {/* Output Box for destination.internal.text Nodes */}
      {data.type === "destination.internal.text" && (
        <div className="mt-3">
          <label className="text-xs text-gray-400 block">Output</label>
          <div
            className="bg-neutral-700 text-white text-xs p-2 rounded w-full h-32 overflow-auto border border-gray-600"
            style={{
              maxWidth: "600px",
              whiteSpace: "pre-wrap",
            }}
          >
            {data.outputs?.[data.alias + ".input"] || "Waiting for output..."}
          </div>
        </div>
      )}

      <div className="text-xs text-gray-400 mt-2">
        Status: <span className="text-white">{data.status || "Idle"}</span>
      </div>

      <Handle type="target" position={Position.Left} className="w-2 h-2 bg-blue-500" />
      <Handle type="source" position={Position.Right} className="w-2 h-2 bg-blue-500" />
    </div>
  );
}
