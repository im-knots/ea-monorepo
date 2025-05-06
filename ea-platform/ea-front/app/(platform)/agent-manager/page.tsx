import AgentTable from "./AgentTable";

export default function AgentManagerPage() {
  return (
    <div className="flex">
      <main className="flex-grow p-4">
        <AgentTable />
      </main>
    </div>
  );
}
