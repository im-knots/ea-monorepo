'use server';

import { cookies } from "next/headers";
import { Node } from "reactflow";

type NodeDefinition = {
  id: string;
  name: string;
  type: string;
  parameters: { key: string; type: string; default: any }[];
  metadata?: { description?: string };
};

async function fetchNodes(filter: string) {
  const cookieStore = await cookies();
  const token = cookieStore.get("token")?.value;

  const response = await fetch(`${process.env.API_URL}/agent-manager/api/v1/nodes?filter=${filter}`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${token}`,
    },
    cache: "no-store"
  });

  if (!response.ok) {
    console.error("Failed to fetch node list", response.status);
    return [];
  }

  return await response.json();
}

export default async function NodeList({
  filter,
  addNodeToFlow,
}: {
  filter: string;
  addNodeToFlow: (node: Node) => void;
}) {
  const nodes: NodeDefinition[] = await fetchNodes(filter);

  return (
    <>
      {nodes.map((node) => {
        const handleClick = () => {
          const uniqueId = `${node.id}-${Math.random().toString(36).substr(2, 9)}`;
          const paramState: Record<string, any> = node.parameters.reduce(
            (acc, param) => {
              acc[param.key] = param.default ?? "";
              return acc;
            },
            {} as Record<string, any>
          );
          const newNode: Node = {
            id: uniqueId,
            type: "custom",
            position: { x: Math.random() * 400, y: Math.random() * 400 },
            data: {
              alias: uniqueId,
              type: node.type,
              label: node.name,
              parameters: node.parameters,
              parametersState: paramState,
            },
          };
          addNodeToFlow(newNode);
        };

        return (
          <div
            key={node.id}
            className="bg-neutral-800 p-4 rounded-lg shadow-md hover:bg-neutral-700 transition cursor-pointer"
            onClick={handleClick}
          >
            <h3 className="text-base font-semibold">{node.name}</h3>
            <p className="text-xs text-neutral-400">{node.metadata?.description || "No description."}</p>
          </div>
        );
      })}
    </>
  );
}
