"use server";

import { cookies } from "next/headers";

async function fetchNodes(filter: string) {
  const cookieStore =  await cookies();
  const token = cookieStore.get("token");

  const response = await fetch(`${process.env.API_URL}/agent-manager/api/v1/nodes?filter=${filter}`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`,
      }
    }
  );
  if (!response.ok) {
    return [];
  }
  return await response.json();
}

export default async function NodeList(props: { filter: string }) {
  const nodes = await fetchNodes(props.filter);
  return (
    <div>
      {nodes.map((node: any) => (
        <div key={node.id}>
          <h2>{node.name}</h2>
          <p>{node.description}</p>
        </div>
      ))}
    </div>
  )
}