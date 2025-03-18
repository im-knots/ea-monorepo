"use client";

import { Suspense, useState } from "react";
import { motion } from "framer-motion";
import { ChevronUp, ChevronDown } from "lucide-react";
import NodeList from "./NodeList"; // new server component!

const NODE_TYPES = ["all", "input", "worker", "destination", "utils"];

export default function NodeLibrary({
  sidebarOpen,
  addNodeToFlow,
}: {
  sidebarOpen: boolean;
  addNodeToFlow: (node: any) => void;
}) {
  const [isOpen, setIsOpen] = useState(true);
  const [selectedType, setSelectedType] = useState<string>("all");

  return (
    <div
      className="absolute bottom-0 bg-neutral-900 text-white shadow-2xl transition-all duration-300"
      style={{
        width: `calc(100% - ${sidebarOpen ? "16rem" : "4rem"})`,
      }}
    >
      {/* Toggle */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-center py-3 bg-neutral-800 hover:bg-neutral-700 transition"
      >
        {isOpen ? <ChevronDown size={20} /> : <ChevronUp size={20} />}
        <span className="ml-2 text-sm font-medium">
          {isOpen ? "Close Node Library" : "Open Node Library"}
        </span>
      </button>

      {/* Library */}
      <motion.div
        initial={{ height: 0, opacity: 0 }}
        animate={{ height: isOpen ? "auto" : 0, opacity: isOpen ? 1 : 0 }}
        className="overflow-hidden bg-neutral-900 border-t border-neutral-700"
      >
        <div className="flex">
          {/* Sidebar Filter */}
          <div className="w-40 bg-neutral-800 p-4 border-r border-neutral-700">
            <h3 className="text-sm font-semibold text-gray-300 mb-2">Filter by Type</h3>
            <ul className="space-y-2">
              {NODE_TYPES.map((type) => (
                <li key={type}>
                  <button
                    onClick={() => setSelectedType(type)}
                    className={`w-full text-left px-3 py-2 rounded-md transition ${
                      selectedType === type ? "bg-green-500 text-white" : "bg-neutral-700 hover:bg-neutral-600"
                    }`}
                  >
                    {type.charAt(0).toUpperCase() + type.slice(1)}
                  </button>
                </li>
              ))}
            </ul>
          </div>

          {/* Server-side Node List */}
          <div className="flex-1 p-4 grid grid-cols-2 md:grid-cols-4 gap-4">
            <Suspense fallback={<p className="text-gray-400 text-center">Loading nodes...</p>}>
              <NodeList filter={selectedType} addNodeToFlow={addNodeToFlow} />
            </Suspense>
          </div>
        </div>
      </motion.div>
    </div>
  );
}
